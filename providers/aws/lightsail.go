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

	"github.com/aws/aws-sdk-go-v2/service/lightsail"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type LightsailGenerator struct {
	AWSService
}

// InitResources enumerates Lightsail instances. Import ID is the instance name.
func (g *LightsailGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := lightsail.NewFromConfig(config)

	var token *string
	for {
		out, err := svc.GetInstances(context.TODO(), &lightsail.GetInstancesInput{PageToken: token})
		if err != nil {
			return err
		}
		for _, instance := range out.Instances {
			name := StringValue(instance.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_lightsail_instance", "aws", defaultAllowEmptyValues))
		}
		if out.NextPageToken == nil {
			return nil
		}
		token = out.NextPageToken
	}
}
