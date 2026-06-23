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
	"github.com/aws/aws-sdk-go-v2/service/inspector"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type InspectorGenerator struct {
	AWSService
}

// InitResources enumerates Inspector Classic (v1) assessment targets and
// templates. Import ID is the resource ARN.
func (g *InspectorGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := inspector.NewFromConfig(config)
	ctx := awsContext()

	targets := inspector.NewListAssessmentTargetsPaginator(svc, &inspector.ListAssessmentTargetsInput{})
	for targets.HasMorePages() {
		page, err := targets.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, arn := range page.AssessmentTargetArns {
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_inspector_assessment_target", "aws", defaultAllowEmptyValues))
		}
	}

	templates := inspector.NewListAssessmentTemplatesPaginator(svc, &inspector.ListAssessmentTemplatesInput{})
	for templates.HasMorePages() {
		page, err := templates.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, arn := range page.AssessmentTemplateArns {
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_inspector_assessment_template", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
