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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v6"
)

type KubernetesGenerator struct {
	AzureService
}

// userAgentPools returns the agent pools that are standalone
// azurerm_kubernetes_cluster_node_pool resources, i.e. User-mode pools. The
// System pool is represented inline as default_node_pool on
// azurerm_kubernetes_cluster, so importing it as a node_pool would conflict.
func userAgentPools(pools []*armcontainerservice.AgentPool) []*armcontainerservice.AgentPool {
	var out []*armcontainerservice.AgentPool
	for _, p := range pools {
		if p == nil || p.Properties == nil || p.Properties.Mode == nil {
			continue
		}
		if *p.Properties.Mode == armcontainerservice.AgentPoolModeUser {
			out = append(out, p)
		}
	}
	return out
}

func (g *KubernetesGenerator) InitResources() error {
	subscriptionID, cred, opts := g.getClientOptions()
	if cred == nil {
		return nil
	}
	client, err := armcontainerservice.NewManagedClustersClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	poolsClient, err := armcontainerservice.NewAgentPoolsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}

	clusters, err := g.listClusters(client)
	if err != nil {
		return err
	}
	for _, cluster := range clusters {
		clusterID := valueOrEmpty(cluster.ID)
		if clusterID == "" {
			continue
		}
		g.AppendSimpleResource(clusterID, valueOrEmpty(cluster.Name), "azurerm_kubernetes_cluster")

		parsed, err := ParseAzureResourceID(clusterID)
		if err != nil {
			return err
		}
		if err := g.appendNodePools(poolsClient, parsed.ResourceGroup, valueOrEmpty(cluster.Name)); err != nil {
			return err
		}
	}
	return nil
}

func (g *KubernetesGenerator) listClusters(client *armcontainerservice.ManagedClustersClient) ([]*armcontainerservice.ManagedCluster, error) {
	var clusters []*armcontainerservice.ManagedCluster
	collect := func(values []*armcontainerservice.ManagedCluster) {
		for _, c := range values {
			if c != nil {
				clusters = append(clusters, c)
			}
		}
	}
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return nil, err
			}
			collect(page.Value)
		}
		return clusters, nil
	}
	for _, rg := range rgs {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return nil, err
			}
			collect(page.Value)
		}
	}
	return clusters, nil
}

func (g *KubernetesGenerator) appendNodePools(client *armcontainerservice.AgentPoolsClient, resourceGroup, clusterName string) error {
	var pools []*armcontainerservice.AgentPool
	pager := client.NewListPager(resourceGroup, clusterName, nil)
	for pager.More() {
		page, err := pager.NextPage(context.TODO())
		if err != nil {
			return err
		}
		pools = append(pools, page.Value...)
	}
	for _, pool := range userAgentPools(pools) {
		id := valueOrEmpty(pool.ID)
		if id == "" {
			continue
		}
		g.AppendSimpleResource(id, clusterName+"_"+valueOrEmpty(pool.Name), "azurerm_kubernetes_cluster_node_pool")
	}
	return nil
}
