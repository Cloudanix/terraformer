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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/servicenetworking/armservicenetworking"
)

// ServiceNetworkingGenerator imports azurerm_application_load_balancer (Traffic
// Controllers).
type ServiceNetworkingGenerator struct {
	AzureService
}

func (g *ServiceNetworkingGenerator) InitResources() error {
	subscriptionID, cred, opts := g.getClientOptions()
	if cred == nil {
		return nil
	}
	client, err := armservicenetworking.NewTrafficControllerInterfaceClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	id := func(i *armservicenetworking.TrafficController) string { return valueOrEmpty(i.ID) }
	name := func(i *armservicenetworking.TrafficController) string { return valueOrEmpty(i.Name) }
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		return appendFromPager(&g.AzureService, client.NewListBySubscriptionPager(nil),
			func(p armservicenetworking.TrafficControllerInterfaceClientListBySubscriptionResponse) []*armservicenetworking.TrafficController {
				return p.Value
			},
			id, name, "azurerm_application_load_balancer")
	}
	for _, rg := range rgs {
		if err := appendFromPager(&g.AzureService, client.NewListByResourceGroupPager(rg, nil),
			func(p armservicenetworking.TrafficControllerInterfaceClientListByResourceGroupResponse) []*armservicenetworking.TrafficController {
				return p.Value
			},
			id, name, "azurerm_application_load_balancer"); err != nil {
			return err
		}
	}
	return nil
}
