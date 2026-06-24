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
	"github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type TimestreamInfluxDBGenerator struct {
	AWSService
}

// InitResources enumerates Timestream for InfluxDB instances. Import ID is the id.
func (g *TimestreamInfluxDBGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := timestreaminfluxdb.NewFromConfig(config)

	p := timestreaminfluxdb.NewListDbInstancesPaginator(svc, &timestreaminfluxdb.ListDbInstancesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, inst := range page.Items {
			id := StringValue(inst.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(inst.Name), "aws_timestreaminfluxdb_db_instance", "aws", defaultAllowEmptyValues))
		}
	}

	for cp := timestreaminfluxdb.NewListDbClustersPaginator(svc, &timestreaminfluxdb.ListDbClustersInput{}); cp.HasMorePages(); {
		page, err := cp.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, c := range page.Items {
			id := StringValue(c.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(c.Name), "aws_timestreaminfluxdb_db_cluster", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
