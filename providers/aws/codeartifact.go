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

	"github.com/aws/aws-sdk-go-v2/service/codeartifact"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type CodeArtifactGenerator struct {
	AWSService
}

// InitResources enumerates CodeArtifact domains. Import ID is the domain name.
func (g *CodeArtifactGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := codeartifact.NewFromConfig(config)

	p := codeartifact.NewListDomainsPaginator(svc, &codeartifact.ListDomainsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, domain := range page.Domains {
			name := StringValue(domain.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_codeartifact_domain", "aws", defaultAllowEmptyValues))
		}
	}

	for rp := codeartifact.NewListRepositoriesPaginator(svc, &codeartifact.ListRepositoriesInput{}); rp.HasMorePages(); {
		page, err := rp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, repo := range page.Repositories {
			arn := StringValue(repo.Arn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(repo.Name), "aws_codeartifact_repository", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
