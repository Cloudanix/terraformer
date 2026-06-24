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
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v7"
)

type NetworkSecurityGroupGenerator struct {
	AzureService
}

// InitResources imports azurerm_network_security_group (+ nested
// azurerm_network_security_rule). Migrated to the Track 2 armnetwork SDK;
// preserves duplicate-name handling.
func (g *NetworkSecurityGroupGenerator) InitResources() error {
	subscriptionID, cred, opts := g.getClientOptions()
	if cred == nil {
		return nil
	}
	nsgClient, err := armnetwork.NewSecurityGroupsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	rulesClient, err := armnetwork.NewSecurityRulesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}

	groups, err := g.listSecurityGroups(nsgClient)
	if err != nil {
		return err
	}
	for _, nsg := range groups {
		nsgID := valueOrEmpty(nsg.ID)
		if nsgID == "" {
			continue
		}
		g.AppendSimpleResourceWithDuplicateCheck(nsgID, valueOrEmpty(nsg.Name), "azurerm_network_security_group")
		parsed, err := ParseAzureResourceID(nsgID)
		if err != nil {
			return err
		}
		if err := g.appendRules(rulesClient, parsed.ResourceGroup, valueOrEmpty(nsg.Name)); err != nil {
			return err
		}
	}
	return nil
}

func (g *NetworkSecurityGroupGenerator) listSecurityGroups(client *armnetwork.SecurityGroupsClient) ([]*armnetwork.SecurityGroup, error) {
	var groups []*armnetwork.SecurityGroup
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		pager := client.NewListAllPager(nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return nil, err
			}
			groups = append(groups, page.Value...)
		}
		return groups, nil
	}
	for _, rg := range rgs {
		pager := client.NewListPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return nil, err
			}
			groups = append(groups, page.Value...)
		}
	}
	return groups, nil
}

func (g *NetworkSecurityGroupGenerator) appendRules(client *armnetwork.SecurityRulesClient, resourceGroup, nsgName string) error {
	pager := client.NewListPager(resourceGroup, nsgName, nil)
	for pager.More() {
		page, err := pager.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, rule := range page.Value {
			if rule == nil {
				continue
			}
			if id := valueOrEmpty(rule.ID); id != "" {
				g.AppendSimpleResourceWithDuplicateCheck(id, valueOrEmpty(rule.Name), "azurerm_network_security_rule")
			}
		}
	}
	return nil
}
