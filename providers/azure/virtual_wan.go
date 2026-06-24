// Copyright 2019 The Terraformer Authors.
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

type VirtualWanGenerator struct {
	AzureService
}

func (g *VirtualWanGenerator) InitResources() error {
	if _, cred, _ := g.getClientOptions(); cred == nil {
		return nil
	}
	if err := g.initVirtualWans(); err != nil {
		return err
	}
	return g.initVirtualHubs()
}

func (g *VirtualWanGenerator) initVirtualWans() error {
	subscriptionID, cred, opts := g.getClientOptions()
	client, err := armnetwork.NewVirtualWansClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	id := func(i *armnetwork.VirtualWAN) string { return valueOrEmpty(i.ID) }
	name := func(i *armnetwork.VirtualWAN) string { return valueOrEmpty(i.Name) }
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		return appendFromPager(&g.AzureService, client.NewListPager(nil),
			func(p armnetwork.VirtualWansClientListResponse) []*armnetwork.VirtualWAN { return p.Value },
			id, name, "azurerm_virtual_wan")
	}
	for _, rg := range rgs {
		if err := appendFromPager(&g.AzureService, client.NewListByResourceGroupPager(rg, nil),
			func(p armnetwork.VirtualWansClientListByResourceGroupResponse) []*armnetwork.VirtualWAN {
				return p.Value
			},
			id, name, "azurerm_virtual_wan"); err != nil {
			return err
		}
	}
	return nil
}

func (g *VirtualWanGenerator) initVirtualHubs() error {
	subscriptionID, cred, opts := g.getClientOptions()
	client, err := armnetwork.NewVirtualHubsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	id := func(i *armnetwork.VirtualHub) string { return valueOrEmpty(i.ID) }
	name := func(i *armnetwork.VirtualHub) string { return valueOrEmpty(i.Name) }
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		return appendFromPager(&g.AzureService, client.NewListPager(nil),
			func(p armnetwork.VirtualHubsClientListResponse) []*armnetwork.VirtualHub { return p.Value },
			id, name, "azurerm_virtual_hub")
	}
	for _, rg := range rgs {
		if err := appendFromPager(&g.AzureService, client.NewListByResourceGroupPager(rg, nil),
			func(p armnetwork.VirtualHubsClientListByResourceGroupResponse) []*armnetwork.VirtualHub {
				return p.Value
			},
			id, name, "azurerm_virtual_hub"); err != nil {
			return err
		}
	}
	return nil
}
