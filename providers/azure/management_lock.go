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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armlocks"
)

type ManagementLockGenerator struct {
	AzureService
}

// InitResources imports azurerm_management_lock. Migrated to the Track 2
// armlocks SDK (was Track 1 services/resources/locks).
func (g *ManagementLockGenerator) InitResources() error {
	subscriptionID, cred, opts := g.getClientOptions()
	if cred == nil {
		return nil
	}
	client, err := armlocks.NewManagementLocksClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	id := func(i *armlocks.ManagementLockObject) string { return valueOrEmpty(i.ID) }
	name := func(i *armlocks.ManagementLockObject) string { return valueOrEmpty(i.Name) }
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		return appendFromPager(&g.AzureService, client.NewListAtSubscriptionLevelPager(nil),
			func(p armlocks.ManagementLocksClientListAtSubscriptionLevelResponse) []*armlocks.ManagementLockObject {
				return p.Value
			},
			id, name, "azurerm_management_lock")
	}
	for _, rg := range rgs {
		if err := appendFromPager(&g.AzureService, client.NewListAtResourceGroupLevelPager(rg, nil),
			func(p armlocks.ManagementLocksClientListAtResourceGroupLevelResponse) []*armlocks.ManagementLockObject {
				return p.Value
			},
			id, name, "azurerm_management_lock"); err != nil {
			return err
		}
	}
	return nil
}
