// Copyright 2020 The Terraformer Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package azure

import (
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
)

type PrivateDNSGenerator struct {
	AzureService
}

// privateDNSRecordResourceType maps a private DNS record-set ARM type to its
// azurerm resource type, or "" for types terraformer does not import.
func privateDNSRecordResourceType(armType string) string {
	parts := strings.Split(armType, "/")
	recordType := parts[len(parts)-1]
	return map[string]string{
		"A":     "azurerm_private_dns_a_record",
		"AAAA":  "azurerm_private_dns_aaaa_record",
		"CNAME": "azurerm_private_dns_cname_record",
		"MX":    "azurerm_private_dns_mx_record",
		"PTR":   "azurerm_private_dns_ptr_record",
		"SRV":   "azurerm_private_dns_srv_record",
		"TXT":   "azurerm_private_dns_txt_record",
	}[recordType]
}

// InitResources imports azurerm_private_dns_zone, its record sets and virtual
// network links. Migrated to the Track 2 armprivatedns SDK.
func (g *PrivateDNSGenerator) InitResources() error {
	subscriptionID, cred, opts := g.getClientOptions()
	if cred == nil {
		return nil
	}
	zonesClient, err := armprivatedns.NewPrivateZonesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	recordsClient, err := armprivatedns.NewRecordSetsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	linksClient, err := armprivatedns.NewVirtualNetworkLinksClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}

	zones, err := g.listPrivateZones(zonesClient)
	if err != nil {
		return err
	}
	for _, zone := range zones {
		zoneID := valueOrEmpty(zone.ID)
		if zoneID == "" {
			continue
		}
		g.AppendSimpleResource(zoneID, valueOrEmpty(zone.Name), "azurerm_private_dns_zone")
		parsed, err := ParseAzureResourceID(zoneID)
		if err != nil {
			return err
		}
		rg, zoneName := parsed.ResourceGroup, valueOrEmpty(zone.Name)

		recordsPager := recordsClient.NewListPager(rg, zoneName, nil)
		for recordsPager.More() {
			page, err := recordsPager.NextPage(context.TODO())
			if err != nil {
				return err
			}
			for _, rs := range page.Value {
				if rs == nil {
					continue
				}
				tfType := privateDNSRecordResourceType(valueOrEmpty(rs.Type))
				if tfType == "" {
					continue
				}
				if id := valueOrEmpty(rs.ID); id != "" {
					g.AppendSimpleResource(id, valueOrEmpty(rs.Name), tfType)
				}
			}
		}

		if err := appendFromPager(&g.AzureService, linksClient.NewListPager(rg, zoneName, nil),
			func(p armprivatedns.VirtualNetworkLinksClientListResponse) []*armprivatedns.VirtualNetworkLink {
				return p.Value
			},
			func(i *armprivatedns.VirtualNetworkLink) string { return valueOrEmpty(i.ID) },
			func(i *armprivatedns.VirtualNetworkLink) string { return valueOrEmpty(i.Name) },
			"azurerm_private_dns_zone_virtual_network_link"); err != nil {
			return err
		}
	}
	return nil
}

func (g *PrivateDNSGenerator) listPrivateZones(client *armprivatedns.PrivateZonesClient) ([]*armprivatedns.PrivateZone, error) {
	var zones []*armprivatedns.PrivateZone
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return nil, err
			}
			zones = append(zones, page.Value...)
		}
		return zones, nil
	}
	for _, rg := range rgs {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return nil, err
			}
			zones = append(zones, page.Value...)
		}
	}
	return zones, nil
}
