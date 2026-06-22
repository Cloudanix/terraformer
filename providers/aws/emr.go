// Copyright 2019 The Terraformer Authors.
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

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/emr"
)

var emrAllowEmptyValues = []string{"tags."}

type EmrGenerator struct {
	AWSService
}

func (g *EmrGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	client := emr.NewFromConfig(config)

	err := g.addClusters(client)
	if err != nil {
		return err
	}
	err = g.addSecurityConfigurations(client)
	if err != nil {
		return err
	}
	return g.addStudios(client)
}

func (g *EmrGenerator) addStudios(client *emr.Client) error {
	p := emr.NewListStudiosPaginator(client, &emr.ListStudiosInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, s := range page.Studios {
			id := StringValue(s.StudioId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(s.Name), "aws_emr_studio", "aws", emrAllowEmptyValues))
		}
	}
	return nil
}

func (g *EmrGenerator) addClusters(client *emr.Client) error {
	p := emr.NewListClustersPaginator(client, &emr.ListClustersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, cluster := range page.Clusters {
			clusterID := StringValue(cluster.Id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				clusterID,
				*cluster.Name,
				"aws_emr_cluster",
				"aws",
				emrAllowEmptyValues,
			))
			g.addInstanceGroupsAndFleets(client, clusterID)
		}
	}
	return nil
}

// addInstanceGroupsAndFleets enumerates a cluster's instance groups and fleets.
// A cluster uses one or the other; the unused list simply returns empty.
// Import IDs are "<cluster-id>/<id>".
func (g *EmrGenerator) addInstanceGroupsAndFleets(client *emr.Client, clusterID string) {
	if clusterID == "" {
		return
	}
	ctx := context.TODO()
	for fp := emr.NewListInstanceFleetsPaginator(client, &emr.ListInstanceFleetsInput{ClusterId: &clusterID}); fp.HasMorePages(); {
		page, err := fp.NextPage(ctx)
		if err != nil {
			break
		}
		for _, f := range page.InstanceFleets {
			id := StringValue(f.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				clusterID+"/"+id, clusterID+"_"+id, "aws_emr_instance_fleet", "aws", emrAllowEmptyValues))
		}
	}
	for gp := emr.NewListInstanceGroupsPaginator(client, &emr.ListInstanceGroupsInput{ClusterId: &clusterID}); gp.HasMorePages(); {
		page, err := gp.NextPage(ctx)
		if err != nil {
			break
		}
		for _, grp := range page.InstanceGroups {
			id := StringValue(grp.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				clusterID+"/"+id, clusterID+"_"+id, "aws_emr_instance_group", "aws", emrAllowEmptyValues))
		}
	}
}

func (g *EmrGenerator) addSecurityConfigurations(client *emr.Client) error {
	p := emr.NewListSecurityConfigurationsPaginator(client, &emr.ListSecurityConfigurationsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, securityConfiguration := range page.SecurityConfigurations {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*securityConfiguration.Name,
				*securityConfiguration.Name,
				"aws_emr_security_configuration",
				"aws",
				emrAllowEmptyValues,
			))
		}
	}
	return nil
}
