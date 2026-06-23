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

	var workgroupNames []string
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
			workgroupNames = append(workgroupNames, name)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_athena_workgroup", "aws", defaultAllowEmptyValues))
		}
	}

	for _, wgName := range workgroupNames {
		wg := wgName
		for nq := athena.NewListNamedQueriesPaginator(svc, &athena.ListNamedQueriesInput{WorkGroup: &wg}); nq.HasMorePages(); {
			page, err := nq.NextPage(ctx)
			if err != nil {
				break
			}
			for _, id := range page.NamedQueryIds {
				if id == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, id, "aws_athena_named_query", "aws", defaultAllowEmptyValues))
			}
		}
		for ps := athena.NewListPreparedStatementsPaginator(svc, &athena.ListPreparedStatementsInput{WorkGroup: &wg}); ps.HasMorePages(); {
			page, err := ps.NextPage(ctx)
			if err != nil {
				break
			}
			for _, s := range page.PreparedStatements {
				name := StringValue(s.StatementName)
				if name == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					wg+"/"+name, wg+"_"+name, "aws_athena_prepared_statement", "aws", defaultAllowEmptyValues))
			}
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
			catName := name
			for dp := athena.NewListDatabasesPaginator(svc, &athena.ListDatabasesInput{CatalogName: &catName}); dp.HasMorePages(); {
				dpage, err := dp.NextPage(ctx)
				if err != nil {
					break
				}
				for _, db := range dpage.DatabaseList {
					dbName := StringValue(db.Name)
					if dbName == "" {
						continue
					}
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						catName+"/"+dbName, catName+"_"+dbName, "aws_athena_database", "aws", defaultAllowEmptyValues))
				}
			}
		}
	}

	for cr := athena.NewListCapacityReservationsPaginator(svc, &athena.ListCapacityReservationsInput{}); cr.HasMorePages(); {
		page, err := cr.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, c := range page.CapacityReservations {
			name := StringValue(c.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_athena_capacity_reservation", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
