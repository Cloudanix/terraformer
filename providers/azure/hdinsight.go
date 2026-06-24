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
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/hdinsight/armhdinsight"
)

type HDInsightGenerator struct {
	AzureService
}

// hdinsightClusterResourceType maps an HDInsight cluster's definition kind to its
// azurerm resource type (there is no generic azurerm_hdinsight_cluster). Unknown
// kinds return "" and are skipped.
func hdinsightClusterResourceType(kind string) string {
	switch strings.ToLower(kind) {
	case "hadoop":
		return "azurerm_hdinsight_hadoop_cluster"
	case "hbase":
		return "azurerm_hdinsight_hbase_cluster"
	case "interactivehive":
		return "azurerm_hdinsight_interactive_query_cluster"
	case "kafka":
		return "azurerm_hdinsight_kafka_cluster"
	case "spark":
		return "azurerm_hdinsight_spark_cluster"
	case "storm":
		return "azurerm_hdinsight_storm_cluster"
	case "mlservices":
		return "azurerm_hdinsight_ml_services_cluster"
	case "rserver":
		return "azurerm_hdinsight_rserver_cluster"
	default:
		return ""
	}
}

func hdinsightKind(cluster *armhdinsight.Cluster) string {
	if cluster.Properties == nil || cluster.Properties.ClusterDefinition == nil {
		return ""
	}
	return valueOrEmpty(cluster.Properties.ClusterDefinition.Kind)
}

func (g *HDInsightGenerator) InitResources() error {
	subscriptionID, cred, opts := g.getClientOptions()
	if cred == nil {
		return nil
	}
	client, err := armhdinsight.NewClustersClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}

	var clusters []*armhdinsight.Cluster
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return err
			}
			clusters = append(clusters, page.Value...)
		}
	} else {
		for _, rg := range rgs {
			pager := client.NewListByResourceGroupPager(rg, nil)
			for pager.More() {
				page, err := pager.NextPage(context.TODO())
				if err != nil {
					return err
				}
				clusters = append(clusters, page.Value...)
			}
		}
	}

	for _, cluster := range clusters {
		if cluster == nil {
			continue
		}
		id := valueOrEmpty(cluster.ID)
		if id == "" {
			continue
		}
		resourceType := hdinsightClusterResourceType(hdinsightKind(cluster))
		if resourceType == "" {
			log.Printf("azurerm hdinsight: cluster %s kind %q not mapped to an azurerm type", id, hdinsightKind(cluster))
			continue
		}
		g.AppendSimpleResource(id, valueOrEmpty(cluster.Name), resourceType)
	}
	return nil
}
