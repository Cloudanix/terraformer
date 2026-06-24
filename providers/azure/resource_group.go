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
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

type ResourceGroupGenerator struct {
	AzureService
}

// InitResources imports azurerm_resource_group. Migrated to the Track 2
// armresources SDK (was Track 1 services/resources). When -R names specific
// resource groups it Gets each; otherwise it lists all groups in the subscription.
func (g *ResourceGroupGenerator) InitResources() error {
	subscriptionID, cred, opts := g.getClientOptions()
	if cred == nil {
		return nil
	}
	client, err := armresources.NewResourceGroupsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}

	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		return appendFromPager(&g.AzureService, client.NewListPager(nil),
			func(p armresources.ResourceGroupsClientListResponse) []*armresources.ResourceGroup {
				return p.Value
			},
			func(i *armresources.ResourceGroup) string { return valueOrEmpty(i.ID) },
			func(i *armresources.ResourceGroup) string { return valueOrEmpty(i.Name) },
			"azurerm_resource_group")
	}
	for _, rg := range rgs {
		resp, err := client.Get(context.TODO(), rg, nil)
		if err != nil {
			return err
		}
		if id := valueOrEmpty(resp.ID); id != "" {
			g.AppendSimpleResource(id, valueOrEmpty(resp.Name), "azurerm_resource_group")
		}
	}
	return nil
}
