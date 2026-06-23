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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/synapse/armsynapse"
)

type SynapseGenerator struct {
	AzureService
}

// InitResources imports azurerm_synapse_workspace, sql_pool, spark_pool,
// firewall_rule and private_link_hub. Migrated to the Track 2 armsynapse SDK.
// (azurerm_synapse_managed_private_endpoint is a data-plane resource served by
// the synapse managed-virtual-network dev endpoint, not the management SDK, so
// it is deferred with the other synapse data-plane resources — see STATUS.md.)
func (g *SynapseGenerator) InitResources() error {
	subscriptionID, cred, opts := g.getClientOptions()
	if cred == nil {
		return nil
	}
	wsClient, err := armsynapse.NewWorkspacesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	sqlClient, err := armsynapse.NewSQLPoolsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	sparkClient, err := armsynapse.NewBigDataPoolsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	fwClient, err := armsynapse.NewIPFirewallRulesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	plhClient, err := armsynapse.NewPrivateLinkHubsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}

	workspaces, err := g.listWorkspaces(wsClient)
	if err != nil {
		return err
	}
	for _, ws := range workspaces {
		wsID := valueOrEmpty(ws.ID)
		if wsID == "" {
			continue
		}
		g.AppendSimpleResource(wsID, valueOrEmpty(ws.Name), "azurerm_synapse_workspace")
		parsed, err := ParseAzureResourceID(wsID)
		if err != nil {
			return err
		}
		rg, wsName := parsed.ResourceGroup, valueOrEmpty(ws.Name)

		if err := appendFromPager(&g.AzureService, sqlClient.NewListByWorkspacePager(rg, wsName, nil),
			func(p armsynapse.SQLPoolsClientListByWorkspaceResponse) []*armsynapse.SQLPool { return p.Value },
			func(i *armsynapse.SQLPool) string { return valueOrEmpty(i.ID) },
			func(i *armsynapse.SQLPool) string { return valueOrEmpty(i.Name) },
			"azurerm_synapse_sql_pool"); err != nil {
			return err
		}
		if err := appendFromPager(&g.AzureService, sparkClient.NewListByWorkspacePager(rg, wsName, nil),
			func(p armsynapse.BigDataPoolsClientListByWorkspaceResponse) []*armsynapse.BigDataPoolResourceInfo {
				return p.Value
			},
			func(i *armsynapse.BigDataPoolResourceInfo) string { return valueOrEmpty(i.ID) },
			func(i *armsynapse.BigDataPoolResourceInfo) string { return valueOrEmpty(i.Name) },
			"azurerm_synapse_spark_pool"); err != nil {
			return err
		}
		if err := appendFromPager(&g.AzureService, fwClient.NewListByWorkspacePager(rg, wsName, nil),
			func(p armsynapse.IPFirewallRulesClientListByWorkspaceResponse) []*armsynapse.IPFirewallRuleInfo {
				return p.Value
			},
			func(i *armsynapse.IPFirewallRuleInfo) string { return valueOrEmpty(i.ID) },
			func(i *armsynapse.IPFirewallRuleInfo) string { return valueOrEmpty(i.Name) },
			"azurerm_synapse_firewall_rule"); err != nil {
			return err
		}
	}

	return g.appendPrivateLinkHubs(plhClient)
}

func (g *SynapseGenerator) listWorkspaces(client *armsynapse.WorkspacesClient) ([]*armsynapse.Workspace, error) {
	var workspaces []*armsynapse.Workspace
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return nil, err
			}
			workspaces = append(workspaces, page.Value...)
		}
		return workspaces, nil
	}
	for _, rg := range rgs {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return nil, err
			}
			workspaces = append(workspaces, page.Value...)
		}
	}
	return workspaces, nil
}

func (g *SynapseGenerator) appendPrivateLinkHubs(client *armsynapse.PrivateLinkHubsClient) error {
	id := func(i *armsynapse.PrivateLinkHub) string { return valueOrEmpty(i.ID) }
	name := func(i *armsynapse.PrivateLinkHub) string { return valueOrEmpty(i.Name) }
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		return appendFromPager(&g.AzureService, client.NewListPager(nil),
			func(p armsynapse.PrivateLinkHubsClientListResponse) []*armsynapse.PrivateLinkHub { return p.Value },
			id, name, "azurerm_synapse_private_link_hub")
	}
	for _, rg := range rgs {
		if err := appendFromPager(&g.AzureService, client.NewListByResourceGroupPager(rg, nil),
			func(p armsynapse.PrivateLinkHubsClientListByResourceGroupResponse) []*armsynapse.PrivateLinkHub {
				return p.Value
			},
			id, name, "azurerm_synapse_private_link_hub"); err != nil {
			return err
		}
	}
	return nil
}
