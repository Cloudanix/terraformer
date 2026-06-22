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

	"github.com/aws/aws-sdk-go-v2/service/comprehend"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type ComprehendGenerator struct {
	AWSService
}

// InitResources enumerates Comprehend document classifiers and entity
// recognizers. Import IDs are the resource ARN.
func (g *ComprehendGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := comprehend.NewFromConfig(config)
	ctx := context.TODO()

	classifiers := comprehend.NewListDocumentClassifiersPaginator(svc, &comprehend.ListDocumentClassifiersInput{})
	for classifiers.HasMorePages() {
		page, err := classifiers.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, c := range page.DocumentClassifierPropertiesList {
			arn := StringValue(c.DocumentClassifierArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_comprehend_document_classifier", "aws", defaultAllowEmptyValues))
		}
	}

	recognizers := comprehend.NewListEntityRecognizersPaginator(svc, &comprehend.ListEntityRecognizersInput{})
	for recognizers.HasMorePages() {
		page, err := recognizers.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, r := range page.EntityRecognizerPropertiesList {
			arn := StringValue(r.EntityRecognizerArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_comprehend_entity_recognizer", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
