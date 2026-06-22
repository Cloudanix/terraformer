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

	"github.com/aws/aws-sdk-go-v2/service/finspace"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type FinspaceGenerator struct {
	AWSService
}

// InitResources enumerates FinSpace Kx environments. Import ID is the environment id.
func (g *FinspaceGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := finspace.NewFromConfig(config)

	p := finspace.NewListKxEnvironmentsPaginator(svc, &finspace.ListKxEnvironmentsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, env := range page.Environments {
			id := StringValue(env.EnvironmentId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(env.Name), "aws_finspace_kx_environment", "aws", defaultAllowEmptyValues))
			g.loadKxChildren(svc, id)
		}
	}
	return nil
}

// loadKxChildren enumerates a Kx environment's databases (and their dataviews),
// clusters, users, scaling groups, and volumes. Import IDs are comma-separated.
func (g *FinspaceGenerator) loadKxChildren(svc *finspace.Client, envID string) {
	ctx := context.TODO()
	env := envID
	add := func(id, name, tfType string) {
		if id != "" {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, name, tfType, "aws", defaultAllowEmptyValues))
		}
	}

	for dp := finspace.NewListKxDatabasesPaginator(svc, &finspace.ListKxDatabasesInput{EnvironmentId: &env}); dp.HasMorePages(); {
		page, err := dp.NextPage(ctx)
		if err != nil {
			break
		}
		for _, db := range page.KxDatabases {
			dbName := StringValue(db.DatabaseName)
			if dbName == "" {
				continue
			}
			add(env+","+dbName, dbName, "aws_finspace_kx_database")
			dn := dbName
			for vp := finspace.NewListKxDataviewsPaginator(svc, &finspace.ListKxDataviewsInput{EnvironmentId: &env, DatabaseName: &dn}); vp.HasMorePages(); {
				vpage, err := vp.NextPage(ctx)
				if err != nil {
					break
				}
				for _, dv := range vpage.KxDataviews {
					name := StringValue(dv.DataviewName)
					if name == "" {
						continue
					}
					add(env+","+dn+","+name, name, "aws_finspace_kx_dataview")
				}
			}
		}
	}

	for sp := finspace.NewListKxScalingGroupsPaginator(svc, &finspace.ListKxScalingGroupsInput{EnvironmentId: &env}); sp.HasMorePages(); {
		page, err := sp.NextPage(ctx)
		if err != nil {
			break
		}
		for _, sg := range page.ScalingGroups {
			name := StringValue(sg.ScalingGroupName)
			add(env+","+name, name, "aws_finspace_kx_scaling_group")
		}
	}

	var token *string
	for {
		out, err := svc.ListKxClusters(ctx, &finspace.ListKxClustersInput{EnvironmentId: &env, NextToken: token})
		if err != nil {
			break
		}
		for _, c := range out.KxClusterSummaries {
			name := StringValue(c.ClusterName)
			add(env+","+name, name, "aws_finspace_kx_cluster")
		}
		if out.NextToken == nil {
			break
		}
		token = out.NextToken
	}

	token = nil
	for {
		out, err := svc.ListKxUsers(ctx, &finspace.ListKxUsersInput{EnvironmentId: &env, NextToken: token})
		if err != nil {
			break
		}
		for _, u := range out.Users {
			name := StringValue(u.UserName)
			add(env+","+name, name, "aws_finspace_kx_user")
		}
		if out.NextToken == nil {
			break
		}
		token = out.NextToken
	}

	token = nil
	for {
		out, err := svc.ListKxVolumes(ctx, &finspace.ListKxVolumesInput{EnvironmentId: &env, NextToken: token})
		if err != nil {
			break
		}
		for _, v := range out.KxVolumeSummaries {
			name := StringValue(v.VolumeName)
			add(env+","+name, name, "aws_finspace_kx_volume")
		}
		if out.NextToken == nil {
			break
		}
		token = out.NextToken
	}
}
