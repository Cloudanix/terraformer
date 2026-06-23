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
	"github.com/aws/aws-sdk-go-v2/service/docdbelastic"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type DocDBElasticGenerator struct {
	AWSService
}

// InitResources enumerates DocumentDB Elastic clusters. Import ID is the cluster ARN.
func (g *DocDBElasticGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := docdbelastic.NewFromConfig(config)

	p := docdbelastic.NewListClustersPaginator(svc, &docdbelastic.ListClustersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, cluster := range page.Clusters {
			arn := StringValue(cluster.ClusterArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(cluster.ClusterName), "aws_docdbelastic_cluster", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
