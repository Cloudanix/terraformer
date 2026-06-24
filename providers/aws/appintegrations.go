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
	"github.com/aws/aws-sdk-go-v2/service/appintegrations"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type AppIntegrationsGenerator struct {
	AWSService
}

// InitResources enumerates AppIntegrations applications. Import ID is the ARN.
func (g *AppIntegrationsGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := appintegrations.NewFromConfig(config)
	p := appintegrations.NewListApplicationsPaginator(svc, &appintegrations.ListApplicationsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, app := range page.Applications {
			arn := StringValue(app.Arn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(app.Name), "aws_appintegrations_application", "aws", defaultAllowEmptyValues))
		}
	}

	ctx := awsContext()
	for dp := appintegrations.NewListDataIntegrationsPaginator(svc, &appintegrations.ListDataIntegrationsInput{}); dp.HasMorePages(); {
		page, err := dp.NextPage(ctx)
		if err != nil {
			break
		}
		for _, d := range page.DataIntegrations {
			name := StringValue(d.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_appintegrations_data_integration", "aws", defaultAllowEmptyValues))
		}
	}
	for ep := appintegrations.NewListEventIntegrationsPaginator(svc, &appintegrations.ListEventIntegrationsInput{}); ep.HasMorePages(); {
		page, err := ep.NextPage(ctx)
		if err != nil {
			break
		}
		for _, e := range page.EventIntegrations {
			name := StringValue(e.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_appintegrations_event_integration", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
