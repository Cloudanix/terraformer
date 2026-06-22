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
			arn := StringValue(gr.Arn)
			version := StringValue(gr.Version)
			if arn != "" && version != "" && version != "DRAFT" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					arn+","+version, id+"_"+version, "aws_bedrock_guardrail_version", "aws", defaultAllowEmptyValues))
			}
		}
	}

	// Account/region-level model invocation logging config; import ID is the region.
	if region := config.Region; region != "" {
		if out, err := svc.GetModelInvocationLoggingConfiguration(ctx, &bedrock.GetModelInvocationLoggingConfigurationInput{}); err == nil && out.LoggingConfig != nil {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				region, region, "aws_bedrock_model_invocation_logging_configuration", "aws", defaultAllowEmptyValues))
		}
	}

	for pmt := bedrock.NewListProvisionedModelThroughputsPaginator(svc, &bedrock.ListProvisionedModelThroughputsInput{}); pmt.HasMorePages(); {
		page, err := pmt.NextPage(ctx)
		if err != nil {
			break
		}
		for _, m := range page.ProvisionedModelSummaries {
			arn := StringValue(m.ProvisionedModelArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(m.ProvisionedModelName), "aws_bedrock_provisioned_model_throughput", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
