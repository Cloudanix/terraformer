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
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/datafactory/armdatafactory"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// SupportedResources maps the Data Factory polymorphic Type discriminator
// (dataset / linked-service / trigger) to the azurerm resource type.
var SupportedResources = map[string]string{
	"AzureBlob":                "azurerm_data_factory_dataset_azure_blob",
	"Binary":                   "azurerm_data_factory_dataset_binary",
	"CosmosDbSqlApiCollection": "azurerm_data_factory_dataset_cosmosdb_sqlapi",
	"CustomDataset":            "azurerm_data_factory_custom_dataset",
	"DelimitedText":            "azurerm_data_factory_dataset_delimited_text",
	"HttpFile":                 "azurerm_data_factory_dataset_http",
	"Json":                     "azurerm_data_factory_dataset_json",
	"MySqlTable":               "azurerm_data_factory_dataset_mysql",
	"Parquet":                  "azurerm_data_factory_dataset_parquet",
	"PostgreSqlTable":          "azurerm_data_factory_dataset_postgresql",
	"SnowflakeTable":           "azurerm_data_factory_dataset_snowflake",
	"SqlServerTable":           "azurerm_data_factory_dataset_sql_server_table",
	"AzureBlobStorage":         "azurerm_data_factory_linked_service_azure_blob_storage",
	"AzureDatabricks":          "azurerm_data_factory_linked_service_azure_databricks",
	"AzureFileStorage":         "azurerm_data_factory_linked_service_azure_file_storage",
	"AzureFunction":            "azurerm_data_factory_linked_service_azure_function",
	"AzureSearch":              "azurerm_data_factory_linked_service_azure_search",
	"AzureSqlDatabase":         "azurerm_data_factory_linked_service_azure_sql_database",
	"AzureTableStorage":        "azurerm_data_factory_linked_service_azure_table_storage",
	"CosmosDb":                 "azurerm_data_factory_linked_service_cosmosdb",
	"CustomDataSource":         "azurerm_data_factory_linked_custom_service",
	"AzureBlobFS":              "azurerm_data_factory_linked_service_data_lake_storage_gen2",
	"AzureKeyVault":            "azurerm_data_factory_linked_service_key_vault",
	"AzureDataExplore":         "azurerm_data_factory_linked_service_kusto",
	"MySql":                    "azurerm_data_factory_linked_service_mysql",
	"OData":                    "azurerm_data_factory_linked_service_odata",
	"PostgreSql":               "azurerm_data_factory_linked_service_postgresql",
	"Sftp":                     "azurerm_data_factory_linked_service_sftp",
	"Snowflake":                "azurerm_data_factory_linked_service_snowflake",
	"SqlServer":                "azurerm_data_factory_linked_service_sql_server",
	"AzureSqlDW":               "azurerm_data_factory_linked_service_synapse",
	"Web":                      "azurerm_data_factory_linked_service_web",
	"BlobEventsTrigger":        "azurerm_data_factory_trigger_blob_event",
	"ScheduleTrigger":          "azurerm_data_factory_trigger_schedule",
	"TumblingWindowTrigger":    "azurerm_data_factory_trigger_tumbling_window",
}

type DataFactoryGenerator struct {
	AzureService
}

func (g *AzureService) appendResourceAs(itemID, itemName, resourceType, abbreviation string) {
	resourceName := abbreviation + "_" + strings.ReplaceAll(itemName, "-", "_")
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(itemID, resourceName, resourceType, g.ProviderName, []string{}))
}

// appendByType maps a polymorphic Type discriminator to an azurerm resource and
// appends it; unknown types are logged and skipped (same as Track 1).
func (g *DataFactoryGenerator) appendByType(id, name, azureType string) {
	if azureType == "" {
		return
	}
	resourceType := SupportedResources[azureType]
	if resourceType == "" {
		log.Printf(`azurerm_data_factory: resource "%s" id: %s type: %s not handled yet by terraform or terraformer`, name, id, azureType)
		return
	}
	g.appendResourceAs(id, name, resourceType, "adf")
}

func (g *DataFactoryGenerator) listFactories(client *armdatafactory.FactoriesClient) ([]*armdatafactory.Factory, error) {
	var factories []*armdatafactory.Factory
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return nil, err
			}
			factories = append(factories, page.Value...)
		}
		return factories, nil
	}
	for _, rg := range rgs {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return nil, err
			}
			factories = append(factories, page.Value...)
		}
	}
	return factories, nil
}

// integrationRuntimeType resolves the azurerm type for an integration runtime,
// distinguishing Managed (Azure) vs Managed-SSIS vs SelfHosted.
func integrationRuntimeType(props armdatafactory.IntegrationRuntimeClassification) string {
	if props == nil {
		return ""
	}
	if managed, ok := props.(*armdatafactory.ManagedIntegrationRuntime); ok {
		if managed.TypeProperties != nil && managed.TypeProperties.SsisProperties != nil {
			return "azurerm_data_factory_integration_runtime_azure_ssis"
		}
		return "azurerm_data_factory_integration_runtime_azure"
	}
	if _, ok := props.(*armdatafactory.SelfHostedIntegrationRuntime); ok {
		return "azurerm_data_factory_integration_runtime_self_hosted"
	}
	return "azurerm_data_factory_integration_runtime_azure"
}

func (g *DataFactoryGenerator) InitResources() error {
	subscriptionID, cred, opts := g.getClientOptions()
	if cred == nil {
		return nil
	}
	factoriesClient, err := armdatafactory.NewFactoriesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	datasetsClient, err := armdatafactory.NewDatasetsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	linkedClient, err := armdatafactory.NewLinkedServicesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	pipelinesClient, err := armdatafactory.NewPipelinesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	triggersClient, err := armdatafactory.NewTriggersClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	dataflowsClient, err := armdatafactory.NewDataFlowsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	irClient, err := armdatafactory.NewIntegrationRuntimesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}

	factories, err := g.listFactories(factoriesClient)
	if err != nil {
		return err
	}
	for _, factory := range factories {
		factoryID := valueOrEmpty(factory.ID)
		if factoryID == "" {
			continue
		}
		g.appendResourceAs(factoryID, valueOrEmpty(factory.Name), "azurerm_data_factory", "adf")
		parsed, err := ParseAzureResourceID(factoryID)
		if err != nil {
			return err
		}
		rg, name := parsed.ResourceGroup, valueOrEmpty(factory.Name)

		// datasets
		dsPager := datasetsClient.NewListByFactoryPager(rg, name, nil)
		for dsPager.More() {
			page, err := dsPager.NextPage(context.TODO())
			if err != nil {
				return err
			}
			for _, ds := range page.Value {
				if ds == nil || ds.Properties == nil {
					continue
				}
				g.appendByType(valueOrEmpty(ds.ID), valueOrEmpty(ds.Name), valueOrEmpty(ds.Properties.GetDataset().Type))
			}
		}
		// linked services
		lsPager := linkedClient.NewListByFactoryPager(rg, name, nil)
		for lsPager.More() {
			page, err := lsPager.NextPage(context.TODO())
			if err != nil {
				return err
			}
			for _, ls := range page.Value {
				if ls == nil || ls.Properties == nil {
					continue
				}
				g.appendByType(valueOrEmpty(ls.ID), valueOrEmpty(ls.Name), valueOrEmpty(ls.Properties.GetLinkedService().Type))
			}
		}
		// triggers
		trPager := triggersClient.NewListByFactoryPager(rg, name, nil)
		for trPager.More() {
			page, err := trPager.NextPage(context.TODO())
			if err != nil {
				return err
			}
			for _, tr := range page.Value {
				if tr == nil || tr.Properties == nil {
					continue
				}
				g.appendByType(valueOrEmpty(tr.ID), valueOrEmpty(tr.Name), valueOrEmpty(tr.Properties.GetTrigger().Type))
			}
		}
		// integration runtimes
		irPager := irClient.NewListByFactoryPager(rg, name, nil)
		for irPager.More() {
			page, err := irPager.NextPage(context.TODO())
			if err != nil {
				return err
			}
			for _, ir := range page.Value {
				if ir == nil || ir.Properties == nil {
					continue
				}
				if rt := integrationRuntimeType(ir.Properties); rt != "" {
					g.appendResourceAs(valueOrEmpty(ir.ID), valueOrEmpty(ir.Name), rt, "adfr")
				}
			}
		}
		// pipelines
		plPager := pipelinesClient.NewListByFactoryPager(rg, name, nil)
		for plPager.More() {
			page, err := plPager.NextPage(context.TODO())
			if err != nil {
				return err
			}
			for _, pl := range page.Value {
				if pl == nil {
					continue
				}
				g.appendResourceAs(valueOrEmpty(pl.ID), valueOrEmpty(pl.Name), "azurerm_data_factory_pipeline", "adfp")
			}
		}
		// data flows
		dfPager := dataflowsClient.NewListByFactoryPager(rg, name, nil)
		for dfPager.More() {
			page, err := dfPager.NextPage(context.TODO())
			if err != nil {
				return err
			}
			for _, df := range page.Value {
				if df == nil {
					continue
				}
				g.appendResourceAs(valueOrEmpty(df.ID), valueOrEmpty(df.Name), "azurerm_data_factory_data_flow", "adfl")
			}
		}
	}
	return nil
}

// PostConvertHook formats the pipeline activities_json property as a heredoc.
func (g *DataFactoryGenerator) PostConvertHook() error {
	for i, resource := range g.Resources {
		if resource.InstanceInfo.Type == "azurerm_data_factory_pipeline" {
			if val, ok := g.Resources[i].Item["activities_json"]; ok && val != nil {
				g.Resources[i].Item["activities_json"] = asHereDoc(val.(string))
			}
		}
	}
	return nil
}
