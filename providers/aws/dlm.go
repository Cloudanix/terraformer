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
	"github.com/aws/aws-sdk-go-v2/service/dlm"
	"github.com/aws/aws-sdk-go-v2/service/dlm/types"
)

type DlmGenerator struct {
	AWSService
}

// InitResources enumerates Data Lifecycle Manager policies. DLM's
// GetLifecyclePolicies returns every policy in a single call (no paginator).
// The Terraform import ID for aws_dlm_lifecycle_policy is the policy ID.
func (g *DlmGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := dlm.NewFromConfig(config)

	output, err := svc.GetLifecyclePolicies(awsContext(), &dlm.GetLifecyclePoliciesInput{})
	if err != nil {
		return err
	}

	g.Resources = appendSimpleResources(g.Resources, output.Policies, "aws_dlm_lifecycle_policy",
		defaultAllowEmptyValues,
		func(p types.LifecyclePolicySummary) string { return StringValue(p.PolicyId) },
		func(p types.LifecyclePolicySummary) string { return StringValue(p.PolicyId) })
	return nil
}
