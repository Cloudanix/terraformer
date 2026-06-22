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

	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type OpenSearchServerlessGenerator struct {
	AWSService
}

// InitResources enumerates OpenSearch Serverless collections. Import ID is the
// collection id.
func (g *OpenSearchServerlessGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := opensearchserverless.NewFromConfig(config)

	p := opensearchserverless.NewListCollectionsPaginator(svc, &opensearchserverless.ListCollectionsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, c := range page.CollectionSummaries {
			id := StringValue(c.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(c.Name), "aws_opensearchserverless_collection", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
