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

	"github.com/aws/aws-sdk-go-v2/service/imagebuilder"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type ImageBuilderGenerator struct {
	AWSService
}

// InitResources enumerates EC2 Image Builder image pipelines. Import ID is the ARN.
func (g *ImageBuilderGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := imagebuilder.NewFromConfig(config)

	ctx := context.TODO()
	add := func(arn, tfType string) {
		if arn != "" {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, tfType, "aws", defaultAllowEmptyValues))
		}
	}

	p := imagebuilder.NewListImagePipelinesPaginator(svc, &imagebuilder.ListImagePipelinesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, pipeline := range page.ImagePipelineList {
			arn := StringValue(pipeline.Arn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(pipeline.Name), "aws_imagebuilder_image_pipeline", "aws", defaultAllowEmptyValues))
		}
	}

	for c := imagebuilder.NewListComponentsPaginator(svc, &imagebuilder.ListComponentsInput{}); c.HasMorePages(); {
		page, err := c.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range page.ComponentVersionList {
			add(StringValue(x.Arn), "aws_imagebuilder_component")
		}
	}
	for c := imagebuilder.NewListContainerRecipesPaginator(svc, &imagebuilder.ListContainerRecipesInput{}); c.HasMorePages(); {
		page, err := c.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range page.ContainerRecipeSummaryList {
			add(StringValue(x.Arn), "aws_imagebuilder_container_recipe")
		}
	}
	for c := imagebuilder.NewListDistributionConfigurationsPaginator(svc, &imagebuilder.ListDistributionConfigurationsInput{}); c.HasMorePages(); {
		page, err := c.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range page.DistributionConfigurationSummaryList {
			add(StringValue(x.Arn), "aws_imagebuilder_distribution_configuration")
		}
	}
	for c := imagebuilder.NewListImagesPaginator(svc, &imagebuilder.ListImagesInput{}); c.HasMorePages(); {
		page, err := c.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range page.ImageVersionList {
			add(StringValue(x.Arn), "aws_imagebuilder_image")
		}
	}
	for c := imagebuilder.NewListImageRecipesPaginator(svc, &imagebuilder.ListImageRecipesInput{}); c.HasMorePages(); {
		page, err := c.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range page.ImageRecipeSummaryList {
			add(StringValue(x.Arn), "aws_imagebuilder_image_recipe")
		}
	}
	for c := imagebuilder.NewListInfrastructureConfigurationsPaginator(svc, &imagebuilder.ListInfrastructureConfigurationsInput{}); c.HasMorePages(); {
		page, err := c.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range page.InfrastructureConfigurationSummaryList {
			add(StringValue(x.Arn), "aws_imagebuilder_infrastructure_configuration")
		}
	}
	for c := imagebuilder.NewListWorkflowsPaginator(svc, &imagebuilder.ListWorkflowsInput{}); c.HasMorePages(); {
		page, err := c.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range page.WorkflowVersionList {
			add(StringValue(x.Arn), "aws_imagebuilder_workflow")
		}
	}
	return nil
}
