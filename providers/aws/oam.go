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
	"github.com/aws/aws-sdk-go-v2/service/oam"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type OAMGenerator struct {
	AWSService
}

// InitResources enumerates CloudWatch Observability Access Manager sinks and
// links. Import IDs are the resource ARN.
func (g *OAMGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := oam.NewFromConfig(config)
	ctx := awsContext()

	sinks := oam.NewListSinksPaginator(svc, &oam.ListSinksInput{})
	for sinks.HasMorePages() {
		page, err := sinks.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, sink := range page.Items {
			arn := StringValue(sink.Arn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_oam_sink", "aws", defaultAllowEmptyValues))
			if pol, err := svc.GetSinkPolicy(ctx, &oam.GetSinkPolicyInput{SinkIdentifier: sink.Arn}); err == nil && StringValue(pol.Policy) != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					arn, arn, "aws_oam_sink_policy", "aws", defaultAllowEmptyValues))
			}
		}
	}

	links := oam.NewListLinksPaginator(svc, &oam.ListLinksInput{})
	for links.HasMorePages() {
		page, err := links.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, link := range page.Items {
			arn := StringValue(link.Arn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_oam_link", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
