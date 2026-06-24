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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dns/armdns"
)

type DNSGenerator struct {
	AzureService
}

// dnsRecordResourceType maps a DNS record-set ARM type
// (e.g. "Microsoft.Network/dnszones/AAAA") to its azurerm resource type, or ""
// for record types terraformer does not import (SOA).
func dnsRecordResourceType(armType string) string {
	parts := strings.Split(armType, "/")
	recordType := parts[len(parts)-1]
	return map[string]string{
		"A":     "azurerm_dns_a_record",
		"AAAA":  "azurerm_dns_aaaa_record",
		"CAA":   "azurerm_dns_caa_record",
		"CNAME": "azurerm_dns_cname_record",
		"MX":    "azurerm_dns_mx_record",
		"NS":    "azurerm_dns_ns_record",
		"PTR":   "azurerm_dns_ptr_record",
		"SRV":   "azurerm_dns_srv_record",
		"TXT":   "azurerm_dns_txt_record",
	}[recordType]
}

// InitResources imports azurerm_dns_zone and its record sets. Migrated to the
// Track 2 armdns SDK (was Track 1 services/dns).
func (g *DNSGenerator) InitResources() error {
	subscriptionID, cred, opts := g.getClientOptions()
	if cred == nil {
		return nil
	}
	zonesClient, err := armdns.NewZonesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	recordsClient, err := armdns.NewRecordSetsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}

	zones, err := g.listZones(zonesClient)
	if err != nil {
		return err
	}
	for _, zone := range zones {
		zoneID := valueOrEmpty(zone.ID)
		if zoneID == "" {
			continue
		}
		g.AppendSimpleResource(zoneID, valueOrEmpty(zone.Name), "azurerm_dns_zone")
		parsed, err := ParseAzureResourceID(zoneID)
		if err != nil {
			return err
		}
		pager := recordsClient.NewListAllByDNSZonePager(parsed.ResourceGroup, valueOrEmpty(zone.Name), nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return err
			}
			for _, rs := range page.Value {
				if rs == nil {
					continue
				}
				tfType := dnsRecordResourceType(valueOrEmpty(rs.Type))
				if tfType == "" {
					continue
				}
				if id := valueOrEmpty(rs.ID); id != "" {
					g.AppendSimpleResource(id, valueOrEmpty(rs.Name), tfType)
				}
			}
		}
	}
	return nil
}

func (g *DNSGenerator) listZones(client *armdns.ZonesClient) ([]*armdns.Zone, error) {
	var zones []*armdns.Zone
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
