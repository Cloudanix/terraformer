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

	"github.com/aws/aws-sdk-go-v2/service/sagemaker"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type SageMakerGenerator struct {
	AWSService
}

// InitResources enumerates the common SageMaker resources: domains, notebook
// instances, models, endpoints, endpoint configs, and code repositories.
// Import IDs are the resource's name (domain id for domains).
func (g *SageMakerGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := sagemaker.NewFromConfig(config)
	ctx := context.TODO()

	domains := sagemaker.NewListDomainsPaginator(svc, &sagemaker.ListDomainsInput{})
	for domains.HasMorePages() {
		page, err := domains.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, d := range page.Domains {
			id := StringValue(d.DomainId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(d.DomainName), "aws_sagemaker_domain", "aws", defaultAllowEmptyValues))
		}
	}

	notebooks := sagemaker.NewListNotebookInstancesPaginator(svc, &sagemaker.ListNotebookInstancesInput{})
	for notebooks.HasMorePages() {
		page, err := notebooks.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, n := range page.NotebookInstances {
			name := StringValue(n.NotebookInstanceName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_sagemaker_notebook_instance", "aws", defaultAllowEmptyValues))
		}
	}

	models := sagemaker.NewListModelsPaginator(svc, &sagemaker.ListModelsInput{})
	for models.HasMorePages() {
		page, err := models.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, m := range page.Models {
			name := StringValue(m.ModelName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_sagemaker_model", "aws", defaultAllowEmptyValues))
		}
	}

	endpoints := sagemaker.NewListEndpointsPaginator(svc, &sagemaker.ListEndpointsInput{})
	for endpoints.HasMorePages() {
		page, err := endpoints.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, ep := range page.Endpoints {
			name := StringValue(ep.EndpointName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_sagemaker_endpoint", "aws", defaultAllowEmptyValues))
		}
	}

	endpointConfigs := sagemaker.NewListEndpointConfigsPaginator(svc, &sagemaker.ListEndpointConfigsInput{})
	for endpointConfigs.HasMorePages() {
		page, err := endpointConfigs.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, ec := range page.EndpointConfigs {
			name := StringValue(ec.EndpointConfigName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_sagemaker_endpoint_configuration", "aws", defaultAllowEmptyValues))
		}
	}

	repos := sagemaker.NewListCodeRepositoriesPaginator(svc, &sagemaker.ListCodeRepositoriesInput{})
	for repos.HasMorePages() {
		page, err := repos.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, repo := range page.CodeRepositorySummaryList {
			name := StringValue(repo.CodeRepositoryName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_sagemaker_code_repository", "aws", defaultAllowEmptyValues))
		}
	}

	return nil
}
