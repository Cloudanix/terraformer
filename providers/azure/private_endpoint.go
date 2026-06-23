// Copyright 2021 The Terraformer Authors.
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

type PrivateEndpointGenerator struct {
	AzureService
}

// InitResources imports azurerm_private_link_service and azurerm_private_endpoint.
// Migrated to the Track 2 armnetwork SDK (was Track 1 services/network).
func (g *PrivateEndpointGenerator) InitResources() error {
	if _, cred, _ := g.getClientOptions(); cred == nil {
		return nil
	}
	if err := g.initPrivateLinkServices(); err != nil {
		return err
	}
	return g.initPrivateEndpoints()
}

func (g *PrivateEndpointGenerator) initPrivateLinkServices() error {
	subscriptionID, cred, opts := g.getClientOptions()
	client, err := armnetwork.NewPrivateLinkServicesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	id := func(i *armnetwork.PrivateLinkService) string { return valueOrEmpty(i.ID) }
	name := func(i *armnetwork.PrivateLinkService) string { return valueOrEmpty(i.Name) }
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		return appendFromPager(&g.AzureService, client.NewListBySubscriptionPager(nil),
			func(p armnetwork.PrivateLinkServicesClientListBySubscriptionResponse) []*armnetwork.PrivateLinkService {
				return p.Value
			},
			id, name, "azurerm_private_link_service")
	}
	for _, rg := range rgs {
		if err := appendFromPager(&g.AzureService, client.NewListPager(rg, nil),
			func(p armnetwork.PrivateLinkServicesClientListResponse) []*armnetwork.PrivateLinkService {
				return p.Value
			},
			id, name, "azurerm_private_link_service"); err != nil {
			return err
		}
	}
	return nil
}

func (g *PrivateEndpointGenerator) initPrivateEndpoints() error {
	subscriptionID, cred, opts := g.getClientOptions()
	client, err := armnetwork.NewPrivateEndpointsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	id := func(i *armnetwork.PrivateEndpoint) string { return valueOrEmpty(i.ID) }
	name := func(i *armnetwork.PrivateEndpoint) string { return valueOrEmpty(i.Name) }
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		return appendFromPager(&g.AzureService, client.NewListBySubscriptionPager(nil),
			func(p armnetwork.PrivateEndpointsClientListBySubscriptionResponse) []*armnetwork.PrivateEndpoint {
				return p.Value
			},
			id, name, "azurerm_private_endpoint")
	}
	for _, rg := range rgs {
		if err := appendFromPager(&g.AzureService, client.NewListPager(rg, nil),
			func(p armnetwork.PrivateEndpointsClientListResponse) []*armnetwork.PrivateEndpoint {
				return p.Value
			},
			id, name, "azurerm_private_endpoint"); err != nil {
			return err
		}
	}
	return nil
}
