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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/managementgroups/armmanagementgroups"
)

// ManagementGroupGenerator imports azurerm_management_group. Management groups
// are tenant-scoped, so this lists all of them and ignores -R.
type ManagementGroupGenerator struct {
	AzureService
}

func (g *ManagementGroupGenerator) InitResources() error {
	_, cred, opts := g.getClientOptions()
	if cred == nil {
		return nil
	}
	client, err := armmanagementgroups.NewClient(cred, opts)
	if err != nil {
		return err
	}
	return appendFromPager(&g.AzureService, client.NewListPager(nil),
		func(p armmanagementgroups.ClientListResponse) []*armmanagementgroups.ManagementGroupInfo {
			return p.Value
		},
		func(i *armmanagementgroups.ManagementGroupInfo) string { return valueOrEmpty(i.ID) },
		func(i *armmanagementgroups.ManagementGroupInfo) string { return valueOrEmpty(i.Name) },
		"azurerm_management_group")
}
