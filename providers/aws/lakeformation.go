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

	"github.com/aws/aws-sdk-go-v2/service/lakeformation"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type LakeFormationGenerator struct {
	AWSService
}

// InitResources enumerates Lake Formation registered resources. Import ID is
// the resource ARN.
func (g *LakeFormationGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := lakeformation.NewFromConfig(config)

	ctx := context.TODO()
	p := lakeformation.NewListResourcesPaginator(svc, &lakeformation.ListResourcesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, r := range page.ResourceInfoList {
			arn := StringValue(r.ResourceArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_lakeformation_resource", "aws", defaultAllowEmptyValues))
		}
	}

	for tp := lakeformation.NewListLFTagsPaginator(svc, &lakeformation.ListLFTagsInput{}); tp.HasMorePages(); {
		page, err := tp.NextPage(ctx)
		if err != nil {
			break
		}
		for _, t := range page.LFTags {
			key := StringValue(t.TagKey)
			if key == "" {
				continue
			}
			catalog := StringValue(t.CatalogId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				catalog+":"+key, catalog+"_"+key, "aws_lakeformation_lf_tag", "aws", defaultAllowEmptyValues))
		}
	}

	for fp := lakeformation.NewListDataCellsFilterPaginator(svc, &lakeformation.ListDataCellsFilterInput{}); fp.HasMorePages(); {
		page, err := fp.NextPage(ctx)
		if err != nil {
			break
		}
		for _, f := range page.DataCellsFilters {
			name := StringValue(f.Name)
			db := StringValue(f.DatabaseName)
			table := StringValue(f.TableName)
			catalog := StringValue(f.TableCatalogId)
			if name == "" || db == "" || table == "" {
				continue
			}
			id := db + "," + name + "," + catalog + "," + table
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, name, "aws_lakeformation_data_cells_filter", "aws", defaultAllowEmptyValues))
		}
	}

	account, err := g.getAccountNumber(config)
	if err != nil {
		return err
	}
	if accountID := StringValue(account); accountID != "" {
		if _, err := svc.GetDataLakeSettings(ctx, &lakeformation.GetDataLakeSettingsInput{}); err == nil {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				accountID, accountID, "aws_lakeformation_data_lake_settings", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
