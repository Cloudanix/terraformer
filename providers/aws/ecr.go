// Copyright 2020 The Terraformer Authors.
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

	"github.com/aws/aws-sdk-go-v2/service/ecr"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type EcrGenerator struct {
	AWSService
}

func (g *EcrGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}

	svc := ecr.NewFromConfig(config)

	p := ecr.NewDescribeRepositoriesPaginator(svc, &ecr.DescribeRepositoriesInput{})
	for p.HasMorePages() {
		page, e := p.NextPage(context.TODO())
		if e != nil {
			return e
		}
		for _, repository := range page.Repositories {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*repository.RepositoryName,
				*repository.RepositoryName,
				"aws_ecr_repository",
				"aws",
				defaultAllowEmptyValues))

			_, err := svc.GetRepositoryPolicy(context.TODO(), &ecr.GetRepositoryPolicyInput{
				RepositoryName: repository.RepositoryName,
				RegistryId:     repository.RegistryId,
			})
			if err == nil {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					*repository.RepositoryName,
					*repository.RepositoryName,
					"aws_ecr_repository_policy",
					"aws",
					defaultAllowEmptyValues))
			}

			_, err = svc.GetLifecyclePolicy(context.TODO(), &ecr.GetLifecyclePolicyInput{
				RepositoryName: repository.RepositoryName,
				RegistryId:     repository.RegistryId,
			})
			if err == nil {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					*repository.RepositoryName,
					*repository.RepositoryName,
					"aws_ecr_lifecycle_policy",
					"aws",
					defaultAllowEmptyValues))
			}
		}
	}

	rules := ecr.NewDescribePullThroughCacheRulesPaginator(svc, &ecr.DescribePullThroughCacheRulesInput{})
	for rules.HasMorePages() {
		page, err := rules.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, r := range page.PullThroughCacheRules {
			prefix := StringValue(r.EcrRepositoryPrefix)
			if prefix == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				prefix, prefix, "aws_ecr_pull_through_cache_rule", "aws", defaultAllowEmptyValues))
		}
	}

	// Account-level singletons keyed by registry (account) ID.
	registryID, err := g.getAccountNumber(config)
	if err != nil {
		return err
	}
	account := StringValue(registryID)
	if account != "" {
		if _, err := svc.GetRegistryScanningConfiguration(context.TODO(), &ecr.GetRegistryScanningConfigurationInput{}); err == nil {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				account, account, "aws_ecr_registry_scanning_configuration", "aws", defaultAllowEmptyValues))
		}
		if out, err := svc.DescribeRegistry(context.TODO(), &ecr.DescribeRegistryInput{}); err == nil &&
			out.ReplicationConfiguration != nil && len(out.ReplicationConfiguration.Rules) > 0 {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				account, account, "aws_ecr_replication_configuration", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}

func (g *EcrGenerator) PostConvertHook() error {
	g.wrapPolicyAttribute(g.Resources, "policy", "aws_ecr_repository_policy", "aws_ecr_lifecycle_policy")
	return nil
}
