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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
)

// RoleAssignmentGenerator imports azurerm_role_assignment. RBAC role
// assignments are subscription-scoped, so this lists the whole subscription and
// ignores -R. (azurerm_role_definition is deferred: it uses a composite
// "<id>|<scope>" import ID and the list returns built-in roles that should not
// be imported.)
type RoleAssignmentGenerator struct {
	AzureService
}

func (g *RoleAssignmentGenerator) InitResources() error {
	subscriptionID, cred, opts := g.getClientOptions()
	if cred == nil {
		return nil
	}
	client, err := armauthorization.NewRoleAssignmentsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	return appendFromPager(&g.AzureService, client.NewListForSubscriptionPager(nil),
		func(p armauthorization.RoleAssignmentsClientListForSubscriptionResponse) []*armauthorization.RoleAssignment {
			return p.Value
		},
		func(i *armauthorization.RoleAssignment) string { return valueOrEmpty(i.ID) },
		func(i *armauthorization.RoleAssignment) string { return valueOrEmpty(i.Name) },
		"azurerm_role_assignment")
}
