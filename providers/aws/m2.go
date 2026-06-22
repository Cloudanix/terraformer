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

	"github.com/aws/aws-sdk-go-v2/service/m2"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type M2Generator struct {
	AWSService
}

// InitResources enumerates Mainframe Modernization (M2) applications. Import ID
// is the application id.
func (g *M2Generator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := m2.NewFromConfig(config)

	ctx := context.TODO()
	p := m2.NewListApplicationsPaginator(svc, &m2.ListApplicationsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, app := range page.Applications {
			id := StringValue(app.ApplicationId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(app.Name), "aws_m2_application", "aws", defaultAllowEmptyValues))
			appID := id
			for dp := m2.NewListDeploymentsPaginator(svc, &m2.ListDeploymentsInput{ApplicationId: &appID}); dp.HasMorePages(); {
				dpage, err := dp.NextPage(ctx)
				if err != nil {
					break
				}
				for _, d := range dpage.Deployments {
					did := StringValue(d.DeploymentId)
					if did == "" {
						continue
					}
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						did+","+appID, appID+"_"+did, "aws_m2_deployment", "aws", defaultAllowEmptyValues))
				}
			}
		}
	}

	for ep := m2.NewListEnvironmentsPaginator(svc, &m2.ListEnvironmentsInput{}); ep.HasMorePages(); {
		page, err := ep.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, e := range page.Environments {
			id := StringValue(e.EnvironmentId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(e.Name), "aws_m2_environment", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
