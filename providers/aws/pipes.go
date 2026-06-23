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
	"github.com/aws/aws-sdk-go-v2/service/pipes"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type PipesGenerator struct {
	AWSService
}

// InitResources enumerates EventBridge Pipes. Import ID is the pipe name.
func (g *PipesGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := pipes.NewFromConfig(config)

	p := pipes.NewListPipesPaginator(svc, &pipes.ListPipesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, pipe := range page.Pipes {
			name := StringValue(pipe.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_pipes_pipe", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
