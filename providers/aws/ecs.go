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
	"fmt"
	"strconv"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

var ecsAllowEmptyValues = []string{"tags."}

type EcsGenerator struct {
	AWSService
}

func (g *EcsGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := ecs.NewFromConfig(config)

	if settings, err := svc.ListAccountSettings(context.TODO(), &ecs.ListAccountSettingsInput{}); err == nil {
		for _, s := range settings.Settings {
			name := string(s.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_ecs_account_setting_default", "aws", ecsAllowEmptyValues))
		}
	}

	p := ecs.NewListClustersPaginator(svc, &ecs.ListClustersInput{})
	for p.HasMorePages() {
		page, e := p.NextPage(context.TODO())
		if e != nil {
			return e
		}
		for _, clusterArn := range page.ClusterArns {
			arnParts := strings.Split(clusterArn, "/")
			clusterName := arnParts[len(arnParts)-1]

			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				clusterArn,
				clusterName,
				"aws_ecs_cluster",
				"aws",
				ecsAllowEmptyValues,
			))

			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				clusterName,
				clusterName,
				"aws_ecs_cluster_capacity_providers",
				"aws",
				ecsAllowEmptyValues,
			))

			servicePage := ecs.NewListServicesPaginator(svc, &ecs.ListServicesInput{
				Cluster: &clusterArn,
			})
			for servicePage.HasMorePages() {
				serviceNextPage, err := servicePage.NextPage(context.TODO())
				if err != nil {
					fmt.Println(err.Error())
					continue
				}
				for _, serviceArn := range serviceNextPage.ServiceArns {
					arnParts := strings.Split(serviceArn, "/")
					serviceName := arnParts[len(arnParts)-1]

					serResp, err := svc.DescribeServices(context.TODO(), &ecs.DescribeServicesInput{
						Services: []string{
							serviceName,
						},
						Cluster: &clusterArn,
					})
					if err != nil {
						fmt.Println(err.Error())
						continue
					}
					serviceDetails := serResp.Services[0]

					g.Resources = append(g.Resources, terraformutils.NewResource(
						serviceArn,
						clusterName+"_"+serviceName,
						"aws_ecs_service",
						"aws",
						map[string]string{
							"task_definition": StringValue(serviceDetails.TaskDefinition),
							"cluster":         clusterName,
							"name":            serviceName,
							"id":              serviceArn,
						},
						ecsAllowEmptyValues,
						map[string]interface{}{},
					))

					taskSets, err := svc.DescribeTaskSets(context.TODO(), &ecs.DescribeTaskSetsInput{
						Cluster: &clusterArn,
						Service: &serviceArn,
					})
					if err != nil {
						continue
					}
					for _, ts := range taskSets.TaskSets {
						id := StringValue(ts.Id)
						if id == "" {
							continue
						}
						g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
							id+","+serviceName+","+clusterName,
							clusterName+"_"+serviceName+"_"+id,
							"aws_ecs_task_set",
							"aws",
							ecsAllowEmptyValues,
						))
					}
				}
			}
		}
	}

	taskDefinitionsMap := map[string]terraformutils.Resource{}
	taskDefinitionsPage := ecs.NewListTaskDefinitionsPaginator(svc, &ecs.ListTaskDefinitionsInput{})
	for taskDefinitionsPage.HasMorePages() {
		taskDefinitionsNextPage, e := taskDefinitionsPage.NextPage(context.TODO())
		if e != nil {
			fmt.Println(e.Error())
			continue
		}
		for _, taskDefinitionArn := range taskDefinitionsNextPage.TaskDefinitionArns {
			arnParts := strings.Split(taskDefinitionArn, ":")
			definitionWithFamily := arnParts[len(arnParts)-2]
			revision, _ := strconv.Atoi(arnParts[len(arnParts)-1])

			// fetch only latest revision of task definitions
			if val, ok := taskDefinitionsMap[definitionWithFamily]; !ok || val.AdditionalFields["revision"].(int) < revision {
				taskDefinitionsMap[definitionWithFamily] = terraformutils.NewResource(
					taskDefinitionArn,
					definitionWithFamily,
					"aws_ecs_task_definition",
					"aws",
					map[string]string{
						"task_definition":       taskDefinitionArn,
						"container_definitions": "{}",
						"family":                "test-task",
						"arn":                   taskDefinitionArn,
					},
					[]string{},
					map[string]interface{}{
						"revision": revision,
					},
				)
			}
		}
	}
	for _, v := range taskDefinitionsMap {
		delete(v.AdditionalFields, "revision")
		g.Resources = append(g.Resources, v)
	}

	// Capacity providers. DescribeCapacityProviders with no names returns all,
	// including the AWS-managed FARGATE / FARGATE_SPOT which aren't importable
	// as aws_ecs_capacity_provider — skip those.
	capacityProviders, err := svc.DescribeCapacityProviders(context.TODO(), &ecs.DescribeCapacityProvidersInput{})
	if err != nil {
		fmt.Println(err.Error())
	} else {
		for _, cp := range capacityProviders.CapacityProviders {
			name := StringValue(cp.Name)
			if name == "" || name == "FARGATE" || name == "FARGATE_SPOT" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_ecs_capacity_provider", "aws", ecsAllowEmptyValues))
		}
	}

	return nil
}

func (g *EcsGenerator) PostConvertHook() error {
	for _, r := range g.Resources {
		if r.InstanceInfo.Type != "aws_ecs_service" {
			continue
		}
		if r.InstanceState.Attributes["propagate_tags"] == "NONE" {
			delete(r.Item, "propagate_tags")
		}
		delete(r.Item, "iam_role")
	}

	return nil
}
