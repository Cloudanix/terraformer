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
	"github.com/aws/aws-sdk-go-v2/service/arcregionswitch"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type ARCRegionSwitchGenerator struct {
	AWSService
}

// InitResources enumerates ARC Region switch plans. Import ID is the plan ARN.
func (g *ARCRegionSwitchGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := arcregionswitch.NewFromConfig(config)
	for p := arcregionswitch.NewListPlansPaginator(svc, &arcregionswitch.ListPlansInput{}); p.HasMorePages(); {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, plan := range page.Plans {
			arn := StringValue(plan.Arn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_arcregionswitch_plan", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
