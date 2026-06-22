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

	"github.com/aws/aws-sdk-go-v2/service/memorydb"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type MemoryDBGenerator struct {
	AWSService
}

// InitResources enumerates MemoryDB clusters. Import ID is the cluster name.
func (g *MemoryDBGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := memorydb.NewFromConfig(config)

	p := memorydb.NewDescribeClustersPaginator(svc, &memorydb.DescribeClustersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, cluster := range page.Clusters {
			name := StringValue(cluster.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_memorydb_cluster", "aws", defaultAllowEmptyValues))
		}
	}

	ctx := context.TODO()
	for acls := memorydb.NewDescribeACLsPaginator(svc, &memorydb.DescribeACLsInput{}); acls.HasMorePages(); {
		page, err := acls.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, a := range page.ACLs {
			g.addNamed(StringValue(a.Name), "aws_memorydb_acl")
		}
	}
	for pg := memorydb.NewDescribeParameterGroupsPaginator(svc, &memorydb.DescribeParameterGroupsInput{}); pg.HasMorePages(); {
		page, err := pg.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, p := range page.ParameterGroups {
			g.addNamed(StringValue(p.Name), "aws_memorydb_parameter_group")
		}
	}
	for sn := memorydb.NewDescribeSnapshotsPaginator(svc, &memorydb.DescribeSnapshotsInput{}); sn.HasMorePages(); {
		page, err := sn.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, s := range page.Snapshots {
			g.addNamed(StringValue(s.Name), "aws_memorydb_snapshot")
		}
	}
	for sg := memorydb.NewDescribeSubnetGroupsPaginator(svc, &memorydb.DescribeSubnetGroupsInput{}); sg.HasMorePages(); {
		page, err := sg.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, s := range page.SubnetGroups {
			g.addNamed(StringValue(s.Name), "aws_memorydb_subnet_group")
		}
	}
	for us := memorydb.NewDescribeUsersPaginator(svc, &memorydb.DescribeUsersInput{}); us.HasMorePages(); {
		page, err := us.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, u := range page.Users {
			g.addNamed(StringValue(u.Name), "aws_memorydb_user")
		}
	}
	return nil
}

func (g *MemoryDBGenerator) addNamed(name, tfType string) {
	if name == "" {
		return
	}
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
		name, name, tfType, "aws", defaultAllowEmptyValues))
}
