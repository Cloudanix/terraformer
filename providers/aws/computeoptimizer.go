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
	"github.com/aws/aws-sdk-go-v2/service/computeoptimizer"
	"github.com/aws/aws-sdk-go-v2/service/computeoptimizer/types"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type ComputeOptimizerGenerator struct {
	AWSService
}

// InitResources emits the account-level Compute Optimizer enrollment status
// (a singleton, imported by account id) when the account is enrolled.
func (g *ComputeOptimizerGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := computeoptimizer.NewFromConfig(config)
	out, err := svc.GetEnrollmentStatus(awsContext(), &computeoptimizer.GetEnrollmentStatusInput{})
	if err != nil {
		return err
	}
	if out.Status == "Inactive" {
		return nil
	}
	account, err := g.getAccountNumber(config)
	if err != nil {
		return err
	}
	if id := StringValue(account); id != "" {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			id, id, "aws_computeoptimizer_enrollment_status", "aws", defaultAllowEmptyValues))
	}
	return g.loadRecommendationPreferences(svc)
}

// loadRecommendationPreferences emits any set recommendation preferences, one
// per (resource type, scope). Import ID is "resourceType,scopeName,scopeValue"
// (per aws_computeoptimizer_recommendation_preferences). GetRecommendationPreferences
// requires a resource type, so iterate the importable enum values.
func (g *ComputeOptimizerGenerator) loadRecommendationPreferences(svc *computeoptimizer.Client) error {
	for _, rt := range types.ResourceType("").Values() {
		if rt == types.ResourceTypeNotApplicable || rt == "Idle" {
			continue
		}
		var token *string
		for {
			out, err := svc.GetRecommendationPreferences(awsContext(), &computeoptimizer.GetRecommendationPreferencesInput{
				ResourceType: rt,
				NextToken:    token,
			})
			if err != nil {
				return err
			}
			for _, d := range out.RecommendationPreferencesDetails {
				if d.Scope == nil || StringValue(d.Scope.Value) == "" {
					continue
				}
				id := string(d.ResourceType) + "," + string(d.Scope.Name) + "," + StringValue(d.Scope.Value)
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, id, "aws_computeoptimizer_recommendation_preferences", "aws", defaultAllowEmptyValues))
			}
			if out.NextToken == nil || StringValue(out.NextToken) == "" {
				break
			}
			token = out.NextToken
		}
	}
	return nil
}
