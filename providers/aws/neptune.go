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

	"github.com/aws/aws-sdk-go-v2/service/neptune"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type NeptuneGenerator struct {
	AWSService
}

// InitResources enumerates Neptune DB clusters. Import ID is the cluster
// identifier.
func (g *NeptuneGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := neptune.NewFromConfig(config)

	p := neptune.NewDescribeDBClustersPaginator(svc, &neptune.DescribeDBClustersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, cluster := range page.DBClusters {
			id := StringValue(cluster.DBClusterIdentifier)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_neptune_cluster", "aws", defaultAllowEmptyValues))
		}
	}

	add := func(name, tfType string) {
		if name != "" {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, tfType, "aws", defaultAllowEmptyValues))
		}
	}
	ctx := context.TODO()
	for p := neptune.NewDescribeDBClusterParameterGroupsPaginator(svc, &neptune.DescribeDBClusterParameterGroupsInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range pg.DBClusterParameterGroups {
			add(StringValue(x.DBClusterParameterGroupName), "aws_neptune_cluster_parameter_group")
		}
	}
	for p := neptune.NewDescribeDBParameterGroupsPaginator(svc, &neptune.DescribeDBParameterGroupsInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range pg.DBParameterGroups {
			add(StringValue(x.DBParameterGroupName), "aws_neptune_parameter_group")
		}
	}
	for p := neptune.NewDescribeDBSubnetGroupsPaginator(svc, &neptune.DescribeDBSubnetGroupsInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range pg.DBSubnetGroups {
			add(StringValue(x.DBSubnetGroupName), "aws_neptune_subnet_group")
		}
	}
	for p := neptune.NewDescribeEventSubscriptionsPaginator(svc, &neptune.DescribeEventSubscriptionsInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range pg.EventSubscriptionsList {
			add(StringValue(x.CustSubscriptionId), "aws_neptune_event_subscription")
		}
	}
	for p := neptune.NewDescribeGlobalClustersPaginator(svc, &neptune.DescribeGlobalClustersInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range pg.GlobalClusters {
			add(StringValue(x.GlobalClusterIdentifier), "aws_neptune_global_cluster")
		}
	}
	return nil
}
