// Copyright 2020 The Terraformer Authors.
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
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codedeploy"
)

var codedeployAllowEmptyValues = []string{"tags."}

type CodeDeployGenerator struct {
	AWSService
}

func (g *CodeDeployGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := codedeploy.NewFromConfig(config)
	p := codedeploy.NewListApplicationsPaginator(svc, &codedeploy.ListApplicationsInput{})
	var resources []terraformutils.Resource
	for p.HasMorePages() {
		page, e := p.NextPage(context.TODO())
		if e != nil {
			return e
		}
		for _, application := range page.Applications {
			resources = append(resources, terraformutils.NewSimpleResource(
				fmt.Sprintf(":%s", application),
				application,
				"aws_codedeploy_app",
				"aws",
				codedeployAllowEmptyValues))

			gp := codedeploy.NewListDeploymentGroupsPaginator(svc, &codedeploy.ListDeploymentGroupsInput{
				ApplicationName: aws.String(application),
			})
			for gp.HasMorePages() {
				gpage, err := gp.NextPage(context.TODO())
				if err != nil {
					return err
				}
				for _, group := range gpage.DeploymentGroups {
					if group == "" {
						continue
					}
					resources = append(resources, terraformutils.NewSimpleResource(
						fmt.Sprintf("%s:%s", application, group),
						fmt.Sprintf("%s_%s", application, group),
						"aws_codedeploy_deployment_group",
						"aws",
						codedeployAllowEmptyValues))
				}
			}
		}
	}

	// Deployment configs. Skip the built-in CodeDeployDefault.* ones — they're
	// AWS-managed and not importable as user resources.
	cp := codedeploy.NewListDeploymentConfigsPaginator(svc, &codedeploy.ListDeploymentConfigsInput{})
	for cp.HasMorePages() {
		cpage, err := cp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, configName := range cpage.DeploymentConfigsList {
			if configName == "" || strings.HasPrefix(configName, "CodeDeployDefault.") {
				continue
			}
			resources = append(resources, terraformutils.NewSimpleResource(
				configName, configName, "aws_codedeploy_deployment_config", "aws", codedeployAllowEmptyValues))
		}
	}

	g.Resources = resources
	return nil
}
