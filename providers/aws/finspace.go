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

	"github.com/aws/aws-sdk-go-v2/service/finspace"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type FinspaceGenerator struct {
	AWSService
}

// InitResources enumerates FinSpace Kx environments. Import ID is the environment id.
func (g *FinspaceGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := finspace.NewFromConfig(config)

	p := finspace.NewListKxEnvironmentsPaginator(svc, &finspace.ListKxEnvironmentsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, env := range page.Environments {
			id := StringValue(env.EnvironmentId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(env.Name), "aws_finspace_kx_environment", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
