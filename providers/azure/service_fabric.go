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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/servicefabric/armservicefabric"
)

type ServiceFabricGenerator struct {
	AzureService
}

// InitResources imports azurerm_service_fabric_cluster. The Clusters API returns
// a flat list (no pager).
func (g *ServiceFabricGenerator) InitResources() error {
	subscriptionID, cred, opts := g.getClientOptions()
	if cred == nil {
		return nil
	}
	client, err := armservicefabric.NewClustersClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}

	var clusters []*armservicefabric.Cluster
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		resp, err := client.List(context.TODO(), nil)
		if err != nil {
			return err
		}
		clusters = resp.Value
	} else {
		for _, rg := range rgs {
			resp, err := client.ListByResourceGroup(context.TODO(), rg, nil)
			if err != nil {
				return err
			}
			clusters = append(clusters, resp.Value...)
		}
	}

	for _, cluster := range clusters {
		if cluster == nil {
			continue
		}
		if id := valueOrEmpty(cluster.ID); id != "" {
			g.AppendSimpleResource(id, valueOrEmpty(cluster.Name), "azurerm_service_fabric_cluster")
		}
	}
	return nil
}
