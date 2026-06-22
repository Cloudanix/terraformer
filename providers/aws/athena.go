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

	"github.com/aws/aws-sdk-go-v2/service/athena"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type AthenaGenerator struct {
	AWSService
}

// InitResources enumerates Athena workgroups and data catalogs. Import IDs are
// the workgroup / data catalog name.
func (g *AthenaGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := athena.NewFromConfig(config)
	ctx := context.TODO()

	workgroups := athena.NewListWorkGroupsPaginator(svc, &athena.ListWorkGroupsInput{})
	for workgroups.HasMorePages() {
		page, err := workgroups.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, wg := range page.WorkGroups {
			name := StringValue(wg.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_athena_workgroup", "aws", defaultAllowEmptyValues))
		}
	}

	catalogs := athena.NewListDataCatalogsPaginator(svc, &athena.ListDataCatalogsInput{})
	for catalogs.HasMorePages() {
		page, err := catalogs.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, cat := range page.DataCatalogsSummary {
			name := StringValue(cat.CatalogName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_athena_data_catalog", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
