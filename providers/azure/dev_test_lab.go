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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/devtestlabs/armdevtestlabs"
)

type DevTestLabGenerator struct {
	AzureService
}

func (g *DevTestLabGenerator) InitResources() error {
	subscriptionID, cred, opts := g.getClientOptions()
	if cred == nil {
		return nil
	}
	client, err := armdevtestlabs.NewLabsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	id := func(i *armdevtestlabs.Lab) string { return valueOrEmpty(i.ID) }
	name := func(i *armdevtestlabs.Lab) string { return valueOrEmpty(i.Name) }
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		return appendFromPager(&g.AzureService, client.NewListBySubscriptionPager(nil),
			func(p armdevtestlabs.LabsClientListBySubscriptionResponse) []*armdevtestlabs.Lab { return p.Value },
			id, name, "azurerm_dev_test_lab")
	}
	for _, rg := range rgs {
		if err := appendFromPager(&g.AzureService, client.NewListByResourceGroupPager(rg, nil),
			func(p armdevtestlabs.LabsClientListByResourceGroupResponse) []*armdevtestlabs.Lab { return p.Value },
			id, name, "azurerm_dev_test_lab"); err != nil {
			return err
		}
	}
	return nil
}
