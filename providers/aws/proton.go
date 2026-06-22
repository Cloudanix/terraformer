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

	"github.com/aws/aws-sdk-go-v2/service/proton"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type ProtonGenerator struct {
	AWSService
}

// InitResources enumerates Proton environments. Import ID is the environment name.
func (g *ProtonGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := proton.NewFromConfig(config)

	p := proton.NewListEnvironmentsPaginator(svc, &proton.ListEnvironmentsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, env := range page.Environments {
			name := StringValue(env.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_proton_environment", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
