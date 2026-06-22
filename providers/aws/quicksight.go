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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type QuickSightGenerator struct {
	AWSService
}

// InitResources enumerates QuickSight data sets and data sources for the current
// account. All QuickSight APIs are scoped by AwsAccountId. Import IDs are
// "<account-id>,<resource-id>".
func (g *QuickSightGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	account, err := g.getAccountNumber(config)
	if err != nil {
		return err
	}
	accountID := StringValue(account)
	svc := quicksight.NewFromConfig(config)
	ctx := context.TODO()

	dataSets := quicksight.NewListDataSetsPaginator(svc, &quicksight.ListDataSetsInput{AwsAccountId: aws.String(accountID)})
	for dataSets.HasMorePages() {
		page, err := dataSets.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, ds := range page.DataSetSummaries {
			id := StringValue(ds.DataSetId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				accountID+","+id, id, "aws_quicksight_data_set", "aws", defaultAllowEmptyValues))
		}
	}

	dataSources := quicksight.NewListDataSourcesPaginator(svc, &quicksight.ListDataSourcesInput{AwsAccountId: aws.String(accountID)})
	for dataSources.HasMorePages() {
		page, err := dataSources.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, src := range page.DataSources {
			id := StringValue(src.DataSourceId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				accountID+","+id, id, "aws_quicksight_data_source", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
