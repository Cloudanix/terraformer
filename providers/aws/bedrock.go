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

	"github.com/aws/aws-sdk-go-v2/service/bedrock"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type BedrockGenerator struct {
	AWSService
}

// InitResources enumerates Bedrock custom models (import by ARN) and guardrails
// (import by guardrail id).
func (g *BedrockGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := bedrock.NewFromConfig(config)
	ctx := context.TODO()

	models := bedrock.NewListCustomModelsPaginator(svc, &bedrock.ListCustomModelsInput{})
	for models.HasMorePages() {
		page, err := models.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, m := range page.ModelSummaries {
			arn := StringValue(m.ModelArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(m.ModelName), "aws_bedrock_custom_model", "aws", defaultAllowEmptyValues))
		}
	}

	guardrails := bedrock.NewListGuardrailsPaginator(svc, &bedrock.ListGuardrailsInput{})
	for guardrails.HasMorePages() {
		page, err := guardrails.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, gr := range page.Guardrails {
			id := StringValue(gr.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(gr.Name), "aws_bedrock_guardrail", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
