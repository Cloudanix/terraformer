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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/managedservices/armmanagedservices"
)

// LighthouseGenerator imports azurerm_lighthouse_definition and
// azurerm_lighthouse_assignment. Lighthouse delegations are subscription-scoped,
// so this lists at the subscription scope and ignores -R.
type LighthouseGenerator struct {
	AzureService
}

func (g *LighthouseGenerator) InitResources() error {
	subscriptionID, cred, opts := g.getClientOptions()
	if cred == nil {
		return nil
	}
	scope := "subscriptions/" + subscriptionID

	defsClient, err := armmanagedservices.NewRegistrationDefinitionsClient(cred, opts)
	if err != nil {
		return err
	}
	if err := appendFromPager(&g.AzureService, defsClient.NewListPager(scope, nil),
		func(p armmanagedservices.RegistrationDefinitionsClientListResponse) []*armmanagedservices.RegistrationDefinition {
			return p.Value
		},
		func(i *armmanagedservices.RegistrationDefinition) string { return valueOrEmpty(i.ID) },
		func(i *armmanagedservices.RegistrationDefinition) string { return valueOrEmpty(i.Name) },
		"azurerm_lighthouse_definition"); err != nil {
		return err
	}

	assignClient, err := armmanagedservices.NewRegistrationAssignmentsClient(cred, opts)
	if err != nil {
		return err
	}
	return appendFromPager(&g.AzureService, assignClient.NewListPager(scope, nil),
		func(p armmanagedservices.RegistrationAssignmentsClientListResponse) []*armmanagedservices.RegistrationAssignment {
			return p.Value
		},
		func(i *armmanagedservices.RegistrationAssignment) string { return valueOrEmpty(i.ID) },
		func(i *armmanagedservices.RegistrationAssignment) string { return valueOrEmpty(i.Name) },
		"azurerm_lighthouse_assignment")
}
