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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armpolicy"
)

// PolicyGenerator imports azurerm_policy_definition, _policy_set_definition and
// _policy_assignment. Policy is subscription-scoped, so this ignores -R. Only
// Custom definitions/set-definitions are imported — the list APIs also return
// the hundreds of built-in policies, which are not user-managed resources.
type PolicyGenerator struct {
	AzureService
}

// isCustomPolicy reports whether a policy (set-)definition is user-authored
// (PolicyType == Custom) and thus an importable azurerm resource. Built-in and
// Static policies are excluded.
func isCustomPolicy(policyType *armpolicy.PolicyType) bool {
	return policyType != nil && *policyType == armpolicy.PolicyTypeCustom
}

func (g *PolicyGenerator) InitResources() error {
	if _, cred, _ := g.getClientOptions(); cred == nil {
		return nil
	}
	if err := g.initDefinitions(); err != nil {
		return err
	}
	if err := g.initSetDefinitions(); err != nil {
		return err
	}
	return g.initAssignments()
}

func (g *PolicyGenerator) initDefinitions() error {
	subscriptionID, cred, opts := g.getClientOptions()
	client, err := armpolicy.NewDefinitionsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, def := range page.Value {
			if def == nil || def.Properties == nil || !isCustomPolicy(def.Properties.PolicyType) {
				continue
			}
			if id := valueOrEmpty(def.ID); id != "" {
				g.AppendSimpleResource(id, valueOrEmpty(def.Name), "azurerm_policy_definition")
			}
		}
	}
	return nil
}

func (g *PolicyGenerator) initSetDefinitions() error {
	subscriptionID, cred, opts := g.getClientOptions()
	client, err := armpolicy.NewSetDefinitionsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, def := range page.Value {
			if def == nil || def.Properties == nil || !isCustomPolicy(def.Properties.PolicyType) {
				continue
			}
			if id := valueOrEmpty(def.ID); id != "" {
				g.AppendSimpleResource(id, valueOrEmpty(def.Name), "azurerm_policy_set_definition")
			}
		}
	}
	return nil
}

func (g *PolicyGenerator) initAssignments() error {
	subscriptionID, cred, opts := g.getClientOptions()
	client, err := armpolicy.NewAssignmentsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	return appendFromPager(&g.AzureService, client.NewListPager(nil),
		func(p armpolicy.AssignmentsClientListResponse) []*armpolicy.Assignment { return p.Value },
		func(i *armpolicy.Assignment) string { return valueOrEmpty(i.ID) },
		func(i *armpolicy.Assignment) string { return valueOrEmpty(i.Name) },
		"azurerm_policy_assignment")
}
