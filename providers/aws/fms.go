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

	"github.com/aws/aws-sdk-go-v2/service/fms"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type FmsGenerator struct {
	AWSService
}

// InitResources enumerates Firewall Manager policies. Import ID is the policy id.
func (g *FmsGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := fms.NewFromConfig(config)

	p := fms.NewListPoliciesPaginator(svc, &fms.ListPoliciesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, policy := range page.PolicyList {
			id := StringValue(policy.PolicyId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_fms_policy", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
