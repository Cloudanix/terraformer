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
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling/types"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type AppAutoScalingGenerator struct {
	AWSService
}

// InitResources enumerates Application Auto Scaling targets, policies, and
// scheduled actions. The describe APIs are scoped to a single ServiceNamespace
// (dynamodb, ecs, rds, …), so every namespace is queried in turn. Import IDs are
// the provider's composite forms:
//   - target            → "<namespace>/<resource-id>/<scalable-dimension>"
//   - policy            → "<namespace>/<resource-id>/<scalable-dimension>/<policy-name>"
//   - scheduled_action  → "<namespace>/<resource-id>/<scalable-dimension>/<action-name>"
func (g *AppAutoScalingGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := applicationautoscaling.NewFromConfig(config)
	ctx := context.TODO()

	for _, namespace := range types.ServiceNamespace("").Values() {
		targets := applicationautoscaling.NewDescribeScalableTargetsPaginator(svc, &applicationautoscaling.DescribeScalableTargetsInput{ServiceNamespace: namespace})
		for targets.HasMorePages() {
			page, err := targets.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, t := range page.ScalableTargets {
				resourceID := StringValue(t.ResourceId)
				if resourceID == "" {
					continue
				}
				id := fmt.Sprintf("%s/%s/%s", namespace, resourceID, t.ScalableDimension)
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, resourceID, "aws_appautoscaling_target", "aws", defaultAllowEmptyValues))
			}
		}

		policies := applicationautoscaling.NewDescribeScalingPoliciesPaginator(svc, &applicationautoscaling.DescribeScalingPoliciesInput{ServiceNamespace: namespace})
		for policies.HasMorePages() {
			page, err := policies.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, p := range page.ScalingPolicies {
				resourceID := StringValue(p.ResourceId)
				policyName := StringValue(p.PolicyName)
				if resourceID == "" || policyName == "" {
					continue
				}
				id := fmt.Sprintf("%s/%s/%s/%s", namespace, resourceID, p.ScalableDimension, policyName)
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, policyName, "aws_appautoscaling_policy", "aws", defaultAllowEmptyValues))
			}
		}

		actions := applicationautoscaling.NewDescribeScheduledActionsPaginator(svc, &applicationautoscaling.DescribeScheduledActionsInput{ServiceNamespace: namespace})
		for actions.HasMorePages() {
			page, err := actions.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, a := range page.ScheduledActions {
				resourceID := StringValue(a.ResourceId)
				actionName := StringValue(a.ScheduledActionName)
				if resourceID == "" || actionName == "" {
					continue
				}
				id := fmt.Sprintf("%s/%s/%s/%s", namespace, resourceID, a.ScalableDimension, actionName)
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, actionName, "aws_appautoscaling_scheduled_action", "aws", defaultAllowEmptyValues))
			}
		}
	}

	return nil
}
