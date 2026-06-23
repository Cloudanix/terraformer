// Copyright 2018 The Terraformer Authors.
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

package aws

import (
	"context"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	gluetypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
)

type GlueGenerator struct {
	AWSService
}

func (g *GlueGenerator) loadGlueCrawlers(svc *glue.Client) error {
	var GlueCrawlerAllowEmptyValues = []string{"tags."}
	p := glue.NewGetCrawlersPaginator(svc, &glue.GetCrawlersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, crawler := range page.Crawlers {
			resource := terraformutils.NewSimpleResource(*crawler.Name, *crawler.Name,
				"aws_glue_crawler",
				"aws",
				GlueCrawlerAllowEmptyValues)
			g.Resources = append(g.Resources, resource)
		}
	}
	return nil
}

func (g *GlueGenerator) loadGlueCatalogDatabase(svc *glue.Client, account *string) (databaseNames []*string, err error) {
	var GlueCatalogDatabaseAllowEmptyValues = []string{"tags."}
	p := glue.NewGetDatabasesPaginator(svc, &glue.GetDatabasesInput{})
	for p.HasMorePages() {
		page, pErr := p.NextPage(context.TODO())
		if pErr != nil {
			return databaseNames, pErr
		}
		for _, catalogDatabase := range page.DatabaseList {
			// format of ID is "CATALOG-ID:DATABASE-NAME".
			// CATALOG-ID is AWS Account ID
			// https://docs.aws.amazon.com/cli/latest/reference/glue/create-database.html#options
			id := *account + ":" + *catalogDatabase.Name
			resource := terraformutils.NewSimpleResource(id, *catalogDatabase.Name,
				"aws_glue_catalog_database",
				"aws",
				GlueCatalogDatabaseAllowEmptyValues)
			g.Resources = append(g.Resources, resource)
			databaseNames = append(databaseNames, catalogDatabase.Name)
		}
	}
	return databaseNames, nil
}

func (g *GlueGenerator) loadGlueCatalogTable(svc *glue.Client, account *string, databaseName *string) error {
	// format of ID is "CATALOG-ID:DATABASE-NAME:TABLE-NAME".
	// CATALOG-ID is AWS Account ID
	// https://docs.aws.amazon.com/cli/latest/reference/glue/create-database.html#options
	var GlueCatalogTableAllowEmptyValues = []string{"tags."}
	p := glue.NewGetTablesPaginator(svc, &glue.GetTablesInput{DatabaseName: databaseName})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, catalogTable := range page.TableList {
			databaseTable := *databaseName + ":" + *catalogTable.Name
			id := *account + ":" + databaseTable
			resource := terraformutils.NewSimpleResource(id, databaseTable,
				"aws_glue_catalog_table",
				"aws",
				GlueCatalogTableAllowEmptyValues)
			g.Resources = append(g.Resources, resource)

			for ip := glue.NewGetPartitionIndexesPaginator(svc, &glue.GetPartitionIndexesInput{
				CatalogId: account, DatabaseName: databaseName, TableName: catalogTable.Name,
			}); ip.HasMorePages(); {
				ipage, err := ip.NextPage(context.TODO())
				if err != nil {
					break
				}
				for _, idx := range ipage.PartitionIndexDescriptorList {
					indexName := StringValue(idx.IndexName)
					if indexName == "" {
						continue
					}
					// import: catalog:db:table:index
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id+":"+indexName, databaseTable+"_"+indexName, "aws_glue_partition_index", "aws", GlueCatalogTableAllowEmptyValues))
				}
			}

			// Partitions (import "catalog:db:table:val1#val2").
			for pp := glue.NewGetPartitionsPaginator(svc, &glue.GetPartitionsInput{
				CatalogId: account, DatabaseName: databaseName, TableName: catalogTable.Name,
			}); pp.HasMorePages(); {
				ppage, err := pp.NextPage(context.TODO())
				if err != nil {
					break
				}
				for _, part := range ppage.Partitions {
					if len(part.Values) == 0 {
						continue
					}
					vals := strings.Join(part.Values, "#")
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id+":"+vals, databaseTable+"_"+vals, "aws_glue_partition", "aws", GlueCatalogTableAllowEmptyValues))
				}
			}

			// Table optimizers (Iceberg compaction/retention/orphan-file-deletion).
			for _, optType := range gluetypes.TableOptimizerType("").Values() {
				if _, err := svc.GetTableOptimizer(context.TODO(), &glue.GetTableOptimizerInput{
					CatalogId: account, DatabaseName: databaseName, TableName: catalogTable.Name, Type: optType,
				}); err == nil {
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						*account+","+*databaseName+","+StringValue(catalogTable.Name)+","+string(optType),
						databaseTable+"_"+string(optType), "aws_glue_catalog_table_optimizer", "aws", GlueCatalogTableAllowEmptyValues))
				}
			}
		}
	}
	return nil
}

func (g *GlueGenerator) loadGlueJobs(svc *glue.Client) error {
	var GlueJobAllowEmptyValues = []string{"tags."}
	p := glue.NewGetJobsPaginator(svc, &glue.GetJobsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, job := range page.Jobs {
			resource := terraformutils.NewSimpleResource(*job.Name, *job.Name,
				"aws_glue_job",
				"aws",
				GlueJobAllowEmptyValues)
			g.Resources = append(g.Resources, resource)
		}
	}
	return nil
}

func (g *GlueGenerator) loadGlueTriggers(svc *glue.Client) error {
	var GlueTriggerAllowEmptyValues = []string{"tags."}
	p := glue.NewGetTriggersPaginator(svc, &glue.GetTriggersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, trigger := range page.Triggers {
			resource := terraformutils.NewSimpleResource(*trigger.Name, *trigger.Name,
				"aws_glue_trigger",
				"aws",
				GlueTriggerAllowEmptyValues)
			g.Resources = append(g.Resources, resource)
		}
	}
	return nil
}

// Generate TerraformResources from AWS API,
// from each database create 1 TerraformResource.
// Need only database name as ID for terraform resource
// AWS api support paging
func (g *GlueGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := glue.NewFromConfig(config)

	account, err := g.getAccountNumber(config)
	if err != nil {
		return err
	}

	if err := g.loadGlueCrawlers(svc); err != nil {
		return err
	}
	var DatabaseNames []*string
	if DatabaseNames, err = g.loadGlueCatalogDatabase(svc, account); err != nil {
		return err
	}
	for _, DatabaseName := range DatabaseNames {
		if err := g.loadGlueCatalogTable(svc, account, DatabaseName); err != nil {
			return err
		}
	}

	if err := g.loadGlueJobs(svc); err != nil {
		return err
	}

	if err := g.loadGlueTriggers(svc); err != nil {
		return err
	}

	if err := g.loadGlueConnections(svc, account); err != nil {
		return err
	}

	if err := g.loadGlueWorkflows(svc); err != nil {
		return err
	}

	if err := g.loadGlueSecurityConfigurations(svc); err != nil {
		return err
	}

	if err := g.loadGlueRegistries(svc); err != nil {
		return err
	}

	if err := g.loadGlueDevEndpoints(svc); err != nil {
		return err
	}

	if err := g.loadGlueDataQualityRulesets(svc); err != nil {
		return err
	}

	if err := g.loadGlueMLTransforms(svc); err != nil {
		return err
	}

	if err := g.loadGlueSchemas(svc); err != nil {
		return err
	}

	if err := g.loadGlueClassifiers(svc); err != nil {
		return err
	}

	for _, DatabaseName := range DatabaseNames {
		if err := g.loadGlueUserDefinedFunctions(svc, account, DatabaseName); err != nil {
			return err
		}
	}

	// Account-level singletons keyed by catalog (account) ID.
	catalogID := StringValue(account)
	if catalogID != "" {
		if _, err := svc.GetDataCatalogEncryptionSettings(context.TODO(), &glue.GetDataCatalogEncryptionSettingsInput{CatalogId: account}); err == nil {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				catalogID, catalogID, "aws_glue_data_catalog_encryption_settings", "aws", defaultAllowEmptyValues))
		}
		if pol, err := svc.GetResourcePolicy(context.TODO(), &glue.GetResourcePolicyInput{}); err == nil && StringValue(pol.PolicyInJson) != "" {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				catalogID, catalogID, "aws_glue_resource_policy", "aws", defaultAllowEmptyValues))
		}
	}

	return nil
}

func (g *GlueGenerator) loadGlueClassifiers(svc *glue.Client) error {
	p := glue.NewGetClassifiersPaginator(svc, &glue.GetClassifiersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, c := range page.Classifiers {
			var name string
			switch {
			case c.GrokClassifier != nil:
				name = StringValue(c.GrokClassifier.Name)
			case c.XMLClassifier != nil:
				name = StringValue(c.XMLClassifier.Name)
			case c.JsonClassifier != nil:
				name = StringValue(c.JsonClassifier.Name)
			case c.CsvClassifier != nil:
				name = StringValue(c.CsvClassifier.Name)
			}
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_glue_classifier", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}

func (g *GlueGenerator) loadGlueUserDefinedFunctions(svc *glue.Client, account, databaseName *string) error {
	catalogID := StringValue(account)
	dbName := StringValue(databaseName)
	p := glue.NewGetUserDefinedFunctionsPaginator(svc, &glue.GetUserDefinedFunctionsInput{
		CatalogId: account, DatabaseName: databaseName, Pattern: aws.String("*"),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, f := range page.UserDefinedFunctions {
			name := StringValue(f.FunctionName)
			if name == "" {
				continue
			}
			id := catalogID + ":" + dbName + ":" + name
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, name, "aws_glue_user_defined_function", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}

func (g *GlueGenerator) loadGlueSchemas(svc *glue.Client) error {
	p := glue.NewListSchemasPaginator(svc, &glue.ListSchemasInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, sc := range page.Schemas {
			arn := StringValue(sc.SchemaArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(sc.SchemaName), "aws_glue_schema", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}

func (g *GlueGenerator) loadGlueMLTransforms(svc *glue.Client) error {
	p := glue.NewGetMLTransformsPaginator(svc, &glue.GetMLTransformsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, t := range page.Transforms {
			id := StringValue(t.TransformId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(t.Name), "aws_glue_ml_transform", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}

func (g *GlueGenerator) loadGlueDevEndpoints(svc *glue.Client) error {
	p := glue.NewGetDevEndpointsPaginator(svc, &glue.GetDevEndpointsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, ep := range page.DevEndpoints {
			name := StringValue(ep.EndpointName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_glue_dev_endpoint", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}

func (g *GlueGenerator) loadGlueDataQualityRulesets(svc *glue.Client) error {
	p := glue.NewListDataQualityRulesetsPaginator(svc, &glue.ListDataQualityRulesetsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, rs := range page.Rulesets {
			name := StringValue(rs.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_glue_data_quality_ruleset", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}

func (g *GlueGenerator) loadGlueConnections(svc *glue.Client, account *string) error {
	catalogID := StringValue(account)
	p := glue.NewGetConnectionsPaginator(svc, &glue.GetConnectionsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, conn := range page.ConnectionList {
			name := StringValue(conn.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				catalogID+":"+name, name, "aws_glue_connection", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}

func (g *GlueGenerator) loadGlueWorkflows(svc *glue.Client) error {
	p := glue.NewListWorkflowsPaginator(svc, &glue.ListWorkflowsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, name := range page.Workflows {
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_glue_workflow", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}

func (g *GlueGenerator) loadGlueSecurityConfigurations(svc *glue.Client) error {
	p := glue.NewGetSecurityConfigurationsPaginator(svc, &glue.GetSecurityConfigurationsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, sc := range page.SecurityConfigurations {
			name := StringValue(sc.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_glue_security_configuration", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}

func (g *GlueGenerator) loadGlueRegistries(svc *glue.Client) error {
	p := glue.NewListRegistriesPaginator(svc, &glue.ListRegistriesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, reg := range page.Registries {
			arn := StringValue(reg.RegistryArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(reg.RegistryName), "aws_glue_registry", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
