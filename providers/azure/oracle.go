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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/oracledatabase/armoracledatabase"
)

type OracleGenerator struct {
	AzureService
}

func (g *OracleGenerator) InitResources() error {
	if _, cred, _ := g.getClientOptions(); cred == nil {
		return nil
	}
	if err := g.initExadataInfrastructures(); err != nil {
		return err
	}
	if err := g.initCloudVMClusters(); err != nil {
		return err
	}
	return g.initAutonomousDatabases()
}

func (g *OracleGenerator) initExadataInfrastructures() error {
	subscriptionID, cred, opts := g.getClientOptions()
	client, err := armoracledatabase.NewCloudExadataInfrastructuresClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	id := func(i *armoracledatabase.CloudExadataInfrastructure) string { return valueOrEmpty(i.ID) }
	name := func(i *armoracledatabase.CloudExadataInfrastructure) string { return valueOrEmpty(i.Name) }
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		return appendFromPager(&g.AzureService, client.NewListBySubscriptionPager(nil),
			func(p armoracledatabase.CloudExadataInfrastructuresClientListBySubscriptionResponse) []*armoracledatabase.CloudExadataInfrastructure {
				return p.Value
			},
			id, name, "azurerm_oracle_exadata_infrastructure")
	}
	for _, rg := range rgs {
		if err := appendFromPager(&g.AzureService, client.NewListByResourceGroupPager(rg, nil),
			func(p armoracledatabase.CloudExadataInfrastructuresClientListByResourceGroupResponse) []*armoracledatabase.CloudExadataInfrastructure {
				return p.Value
			},
			id, name, "azurerm_oracle_exadata_infrastructure"); err != nil {
			return err
		}
	}
	return nil
}

func (g *OracleGenerator) initCloudVMClusters() error {
	subscriptionID, cred, opts := g.getClientOptions()
	client, err := armoracledatabase.NewCloudVMClustersClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	id := func(i *armoracledatabase.CloudVMCluster) string { return valueOrEmpty(i.ID) }
	name := func(i *armoracledatabase.CloudVMCluster) string { return valueOrEmpty(i.Name) }
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		return appendFromPager(&g.AzureService, client.NewListBySubscriptionPager(nil),
			func(p armoracledatabase.CloudVMClustersClientListBySubscriptionResponse) []*armoracledatabase.CloudVMCluster {
				return p.Value
			},
			id, name, "azurerm_oracle_cloud_vm_cluster")
	}
	for _, rg := range rgs {
		if err := appendFromPager(&g.AzureService, client.NewListByResourceGroupPager(rg, nil),
			func(p armoracledatabase.CloudVMClustersClientListByResourceGroupResponse) []*armoracledatabase.CloudVMCluster {
				return p.Value
			},
			id, name, "azurerm_oracle_cloud_vm_cluster"); err != nil {
			return err
		}
	}
	return nil
}

func (g *OracleGenerator) initAutonomousDatabases() error {
	subscriptionID, cred, opts := g.getClientOptions()
	client, err := armoracledatabase.NewAutonomousDatabasesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	id := func(i *armoracledatabase.AutonomousDatabase) string { return valueOrEmpty(i.ID) }
	name := func(i *armoracledatabase.AutonomousDatabase) string { return valueOrEmpty(i.Name) }
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		return appendFromPager(&g.AzureService, client.NewListBySubscriptionPager(nil),
			func(p armoracledatabase.AutonomousDatabasesClientListBySubscriptionResponse) []*armoracledatabase.AutonomousDatabase {
				return p.Value
			},
			id, name, "azurerm_oracle_autonomous_database")
	}
	for _, rg := range rgs {
		if err := appendFromPager(&g.AzureService, client.NewListByResourceGroupPager(rg, nil),
			func(p armoracledatabase.AutonomousDatabasesClientListByResourceGroupResponse) []*armoracledatabase.AutonomousDatabase {
				return p.Value
			},
			id, name, "azurerm_oracle_autonomous_database"); err != nil {
			return err
		}
	}
	return nil
}
