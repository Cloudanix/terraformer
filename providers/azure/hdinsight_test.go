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

import "testing"

func TestHDInsightClusterResourceType(t *testing.T) {
	cases := map[string]string{
		"Hadoop":          "azurerm_hdinsight_hadoop_cluster",
		"HBase":           "azurerm_hdinsight_hbase_cluster",
		"INTERACTIVEHIVE": "azurerm_hdinsight_interactive_query_cluster",
		"Kafka":           "azurerm_hdinsight_kafka_cluster",
		"Spark":           "azurerm_hdinsight_spark_cluster",
		"Storm":           "azurerm_hdinsight_storm_cluster",
		"MLServices":      "azurerm_hdinsight_ml_services_cluster",
		"RServer":         "azurerm_hdinsight_rserver_cluster",
		"spark":           "azurerm_hdinsight_spark_cluster", // case-insensitive
		"":                "",                                // unknown -> skip
		"Unknown":         "",
	}
	for kind, want := range cases {
		if got := hdinsightClusterResourceType(kind); got != want {
			t.Errorf("hdinsightClusterResourceType(%q) = %q, want %q", kind, got, want)
		}
	}
}
