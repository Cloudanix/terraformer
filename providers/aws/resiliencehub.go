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
	"github.com/aws/aws-sdk-go-v2/service/resiliencehub"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type ResilienceHubGenerator struct {
	AWSService
}

func (g *ResilienceHubGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := resiliencehub.NewFromConfig(config)
	p := resiliencehub.NewListAppsPaginator(svc, &resiliencehub.ListAppsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, item := range page.AppSummaries {
			id := StringValue(item.AppArn)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(item.Name), "aws_resiliencehub_app", "aws", defaultAllowEmptyValues))
		}
	}
	for pp := resiliencehub.NewListResiliencyPoliciesPaginator(svc, &resiliencehub.ListResiliencyPoliciesInput{}); pp.HasMorePages(); {
		page, err := pp.NextPage(awsContext())
		if err != nil {
			break
		}
		for _, pol := range page.ResiliencyPolicies {
			arn := StringValue(pol.PolicyArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(pol.PolicyName), "aws_resiliencehub_resiliency_policy", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
