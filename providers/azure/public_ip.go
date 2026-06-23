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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v7"
)

type PublicIPGenerator struct {
	AzureService
}

// InitResources imports azurerm_public_ip and azurerm_public_ip_prefix.
// Migrated to the Track 2 armnetwork SDK (was Track 1 services/network).
func (g *PublicIPGenerator) InitResources() error {
	if _, cred, _ := g.getClientOptions(); cred == nil {
		return nil
	}
	if err := g.initPublicIPAddresses(); err != nil {
		return err
	}
	return g.initPublicIPPrefixes()
}

func (g *PublicIPGenerator) initPublicIPAddresses() error {
	subscriptionID, cred, opts := g.getClientOptions()
	client, err := armnetwork.NewPublicIPAddressesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	id := func(i *armnetwork.PublicIPAddress) string { return valueOrEmpty(i.ID) }
	name := func(i *armnetwork.PublicIPAddress) string { return valueOrEmpty(i.Name) }
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		return appendFromPager(&g.AzureService, client.NewListAllPager(nil),
			func(p armnetwork.PublicIPAddressesClientListAllResponse) []*armnetwork.PublicIPAddress {
				return p.Value
			},
			id, name, "azurerm_public_ip")
	}
	for _, rg := range rgs {
		if err := appendFromPager(&g.AzureService, client.NewListPager(rg, nil),
			func(p armnetwork.PublicIPAddressesClientListResponse) []*armnetwork.PublicIPAddress {
				return p.Value
			},
			id, name, "azurerm_public_ip"); err != nil {
			return err
		}
	}
	return nil
}

func (g *PublicIPGenerator) initPublicIPPrefixes() error {
	subscriptionID, cred, opts := g.getClientOptions()
	client, err := armnetwork.NewPublicIPPrefixesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	id := func(i *armnetwork.PublicIPPrefix) string { return valueOrEmpty(i.ID) }
	name := func(i *armnetwork.PublicIPPrefix) string { return valueOrEmpty(i.Name) }
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		return appendFromPager(&g.AzureService, client.NewListAllPager(nil),
			func(p armnetwork.PublicIPPrefixesClientListAllResponse) []*armnetwork.PublicIPPrefix {
				return p.Value
			},
			id, name, "azurerm_public_ip_prefix")
	}
	for _, rg := range rgs {
		if err := appendFromPager(&g.AzureService, client.NewListPager(rg, nil),
			func(p armnetwork.PublicIPPrefixesClientListResponse) []*armnetwork.PublicIPPrefix {
				return p.Value
			},
			id, name, "azurerm_public_ip_prefix"); err != nil {
			return err
		}
	}
	return nil
}
