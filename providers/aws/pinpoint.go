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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pinpoint"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type PinpointGenerator struct {
	AWSService
}

// InitResources enumerates Pinpoint apps. GetApps is token-paginated manually
// (no v2 paginator). Import ID is the application id.
func (g *PinpointGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := pinpoint.NewFromConfig(config)

	var token *string
	for {
		out, err := svc.GetApps(context.TODO(), &pinpoint.GetAppsInput{Token: token})
		if err != nil {
			return err
		}
		if out.ApplicationsResponse == nil {
			return nil
		}
		for _, app := range out.ApplicationsResponse.Item {
			id := StringValue(app.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(app.Name), "aws_pinpoint_app", "aws", defaultAllowEmptyValues))
		}
		if out.ApplicationsResponse.NextToken == nil {
			return nil
		}
		token = aws.String(StringValue(out.ApplicationsResponse.NextToken))
	}
}
