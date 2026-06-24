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
		page, err := p.NextPage(awsContext())
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
			domainName := name
			if _, err := svc.GetDomainPermissionsPolicy(awsContext(), &codeartifact.GetDomainPermissionsPolicyInput{Domain: &domainName}); err == nil {
				if arn := StringValue(domain.Arn); arn != "" {
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						arn, name, "aws_codeartifact_domain_permissions_policy", "aws", defaultAllowEmptyValues))
				}
			}
		}
	}

	for rp := codeartifact.NewListRepositoriesPaginator(svc, &codeartifact.ListRepositoriesInput{}); rp.HasMorePages(); {
		page, err := rp.NextPage(awsContext())
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
			repoDomain, repoName := StringValue(repo.DomainName), StringValue(repo.Name)
			if repoDomain != "" && repoName != "" {
				if _, err := svc.GetRepositoryPermissionsPolicy(awsContext(), &codeartifact.GetRepositoryPermissionsPolicyInput{Domain: &repoDomain, Repository: &repoName}); err == nil {
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						arn, repoName, "aws_codeartifact_repository_permissions_policy", "aws", defaultAllowEmptyValues))
				}
			}
		}
	}
	return nil
}
