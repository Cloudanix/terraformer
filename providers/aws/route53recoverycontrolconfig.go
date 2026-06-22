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

	"github.com/aws/aws-sdk-go-v2/service/route53recoverycontrolconfig"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type Route53RecoveryControlConfigGenerator struct {
	AWSService
}

// InitResources enumerates Route 53 Application Recovery Controller clusters.
// Import ID is the cluster ARN.
func (g *Route53RecoveryControlConfigGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := route53recoverycontrolconfig.NewFromConfig(config)
	p := route53recoverycontrolconfig.NewListClustersPaginator(svc, &route53recoverycontrolconfig.ListClustersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, c := range page.Clusters {
			arn := StringValue(c.ClusterArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(c.Name), "aws_route53recoverycontrolconfig_cluster", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
