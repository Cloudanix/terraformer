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
	"github.com/aws/aws-sdk-go-v2/service/timestreamwrite"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type TimestreamWriteGenerator struct {
	AWSService
}

// InitResources enumerates Timestream databases. Import ID is the database name.
func (g *TimestreamWriteGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := timestreamwrite.NewFromConfig(config)

	p := timestreamwrite.NewListDatabasesPaginator(svc, &timestreamwrite.ListDatabasesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, db := range page.Databases {
			name := StringValue(db.DatabaseName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_timestreamwrite_database", "aws", defaultAllowEmptyValues))
			dbName := name
			for tp := timestreamwrite.NewListTablesPaginator(svc, &timestreamwrite.ListTablesInput{DatabaseName: &dbName}); tp.HasMorePages(); {
				tpage, err := tp.NextPage(awsContext())
				if err != nil {
					break
				}
				for _, t := range tpage.Tables {
					table := StringValue(t.TableName)
					if table == "" {
						continue
					}
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						dbName+":"+table, dbName+"_"+table, "aws_timestreamwrite_table", "aws", defaultAllowEmptyValues))
				}
			}
		}
	}
	return nil
}
