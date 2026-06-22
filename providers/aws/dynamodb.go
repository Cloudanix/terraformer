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

package aws

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var dynamodbAllowEmptyValues = []string{"tags."}

type DynamoDbGenerator struct {
	AWSService
}

func (g *DynamoDbGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := dynamodb.NewFromConfig(config)
	var tableNames []string
	p := dynamodb.NewListTablesPaginator(svc, &dynamodb.ListTablesInput{})
	for p.HasMorePages() {
		page, e := p.NextPage(context.TODO())
		if e != nil {
			return e
		}
		for _, tableName := range page.TableNames {
			tableNames = append(tableNames, tableName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				tableName,
				tableName,
				"aws_dynamodb_table",
				"aws",
				dynamodbAllowEmptyValues,
			))
		}
	}

	for _, tableName := range tableNames {
		if out, err := svc.DescribeKinesisStreamingDestination(context.TODO(),
			&dynamodb.DescribeKinesisStreamingDestinationInput{TableName: &tableName}); err == nil {
			for _, d := range out.KinesisDataStreamDestinations {
				streamArn := StringValue(d.StreamArn)
				if streamArn == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					tableName+","+streamArn, tableName, "aws_dynamodb_kinesis_streaming_destination", "aws", dynamodbAllowEmptyValues))
			}
		}
		if out, err := svc.DescribeContributorInsights(context.TODO(),
			&dynamodb.DescribeContributorInsightsInput{TableName: &tableName}); err == nil &&
			out.ContributorInsightsStatus != "" && out.ContributorInsightsStatus != "DISABLED" {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				tableName, tableName, "aws_dynamodb_contributor_insights", "aws", dynamodbAllowEmptyValues))
		}
		if out, err := svc.DescribeTable(context.TODO(),
			&dynamodb.DescribeTableInput{TableName: &tableName}); err == nil && out.Table != nil {
			for _, r := range out.Table.Replicas {
				region := StringValue(r.RegionName)
				if region == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					tableName+":"+region, tableName+"_"+region, "aws_dynamodb_table_replica", "aws", dynamodbAllowEmptyValues))
			}
		}
	}

	// Legacy (v2017.11.29) global tables. Newer global tables are modeled as
	// table replicas, not returned here.
	var startName *string
	for {
		out, err := svc.ListGlobalTables(context.TODO(), &dynamodb.ListGlobalTablesInput{ExclusiveStartGlobalTableName: startName})
		if err != nil {
			return err
		}
		for _, gt := range out.GlobalTables {
			name := StringValue(gt.GlobalTableName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_dynamodb_global_table", "aws", dynamodbAllowEmptyValues))
		}
		if out.LastEvaluatedGlobalTableName == nil {
			break
		}
		startName = out.LastEvaluatedGlobalTableName
	}
	return nil
}

func (g *DynamoDbGenerator) PostConvertHook() error {
	for _, r := range g.Resources {
		if r.InstanceInfo.Type != "aws_dynamodb_table" {
			continue
		}
		if val, ok := r.InstanceState.Attributes["ttl.0.enabled"]; ok && val == "false" {
			delete(r.Item, "ttl")
		}
	}
	return nil
}
