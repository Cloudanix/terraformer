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

type ApplicationGatewayGenerator struct {
	AzureService
}

// InitResources imports azurerm_application_gateway. Migrated to the Track 2
// armnetwork SDK (was Track 1 services/network).
func (g *ApplicationGatewayGenerator) InitResources() error {
	subscriptionID, cred, opts := g.getClientOptions()
	if cred == nil {
		return nil
	}
	client, err := armnetwork.NewApplicationGatewaysClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	id := func(i *armnetwork.ApplicationGateway) string { return valueOrEmpty(i.ID) }
	name := func(i *armnetwork.ApplicationGateway) string { return valueOrEmpty(i.Name) }
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		return appendFromPager(&g.AzureService, client.NewListAllPager(nil),
			func(p armnetwork.ApplicationGatewaysClientListAllResponse) []*armnetwork.ApplicationGateway {
				return p.Value
			},
			id, name, "azurerm_application_gateway")
	}
	for _, rg := range rgs {
		if err := appendFromPager(&g.AzureService, client.NewListPager(rg, nil),
			func(p armnetwork.ApplicationGatewaysClientListResponse) []*armnetwork.ApplicationGateway {
				return p.Value
			},
			id, name, "azurerm_application_gateway"); err != nil {
			return err
		}
	}
	return nil
}
