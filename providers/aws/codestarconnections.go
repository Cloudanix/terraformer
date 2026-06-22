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

	"github.com/aws/aws-sdk-go-v2/service/codestarconnections"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type CodeStarConnectionsGenerator struct {
	AWSService
}

// InitResources enumerates CodeStar (CodeConnections) connections. Import ID is
// the connection ARN.
func (g *CodeStarConnectionsGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := codestarconnections.NewFromConfig(config)

	p := codestarconnections.NewListConnectionsPaginator(svc, &codestarconnections.ListConnectionsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, conn := range page.Connections {
			arn := StringValue(conn.ConnectionArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(conn.ConnectionName), "aws_codestarconnections_connection", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
