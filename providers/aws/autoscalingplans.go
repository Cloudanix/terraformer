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
	"github.com/aws/aws-sdk-go-v2/service/autoscalingplans"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type AutoScalingPlansGenerator struct {
	AWSService
}

// InitResources enumerates Auto Scaling scaling plans. Import ID is the plan name.
func (g *AutoScalingPlansGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := autoscalingplans.NewFromConfig(config)

	var token *string
	for {
		out, err := svc.DescribeScalingPlans(awsContext(), &autoscalingplans.DescribeScalingPlansInput{NextToken: token})
		if err != nil {
			return err
		}
		for _, plan := range out.ScalingPlans {
			name := StringValue(plan.ScalingPlanName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_autoscalingplans_scaling_plan", "aws", defaultAllowEmptyValues))
		}
		if out.NextToken == nil {
			return nil
		}
		token = out.NextToken
	}
}
