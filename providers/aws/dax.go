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

	"github.com/aws/aws-sdk-go-v2/service/dax"
	"github.com/aws/aws-sdk-go-v2/service/dax/types"
)

type DaxGenerator struct {
	AWSService
}

// InitResources enumerates DynamoDB Accelerator (DAX) clusters. DescribeClusters
// has no generated paginator, so we page manually via NextToken. The Terraform
// import ID for aws_dax_cluster is the cluster name.
func (g *DaxGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := dax.NewFromConfig(config)

	var nextToken *string
	for {
		output, err := svc.DescribeClusters(context.TODO(), &dax.DescribeClustersInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, output.Clusters, "aws_dax_cluster",
			defaultAllowEmptyValues,
			func(c types.Cluster) string { return StringValue(c.ClusterName) },
			func(c types.Cluster) string { return StringValue(c.ClusterName) })
		nextToken = output.NextToken
		if nextToken == nil {
			break
		}
	}

	nextToken = nil
	for {
		out, err := svc.DescribeParameterGroups(context.TODO(), &dax.DescribeParameterGroupsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, out.ParameterGroups, "aws_dax_parameter_group",
			defaultAllowEmptyValues,
			func(p types.ParameterGroup) string { return StringValue(p.ParameterGroupName) },
			func(p types.ParameterGroup) string { return StringValue(p.ParameterGroupName) })
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}

	nextToken = nil
	for {
		out, err := svc.DescribeSubnetGroups(context.TODO(), &dax.DescribeSubnetGroupsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, out.SubnetGroups, "aws_dax_subnet_group",
			defaultAllowEmptyValues,
			func(s types.SubnetGroup) string { return StringValue(s.SubnetGroupName) },
			func(s types.SubnetGroup) string { return StringValue(s.SubnetGroupName) })
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
