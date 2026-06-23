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
	"github.com/aws/aws-sdk-go-v2/service/rekognition"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type RekognitionGenerator struct {
	AWSService
}

// InitResources enumerates Rekognition collections and stream processors.
// Import IDs are the collection id / stream processor name.
func (g *RekognitionGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := rekognition.NewFromConfig(config)
	ctx := awsContext()

	collections := rekognition.NewListCollectionsPaginator(svc, &rekognition.ListCollectionsInput{})
	for collections.HasMorePages() {
		page, err := collections.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, id := range page.CollectionIds {
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_rekognition_collection", "aws", defaultAllowEmptyValues))
		}
	}

	processors := rekognition.NewListStreamProcessorsPaginator(svc, &rekognition.ListStreamProcessorsInput{})
	for processors.HasMorePages() {
		page, err := processors.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, sp := range page.StreamProcessors {
			name := StringValue(sp.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_rekognition_stream_processor", "aws", defaultAllowEmptyValues))
		}
	}

	for projects := rekognition.NewDescribeProjectsPaginator(svc, &rekognition.DescribeProjectsInput{}); projects.HasMorePages(); {
		page, err := projects.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, p := range page.ProjectDescriptions {
			arn := StringValue(p.ProjectArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_rekognition_project", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
