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

	"github.com/aws/aws-sdk-go-v2/service/costoptimizationhub"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type CostOptimizationHubGenerator struct {
	AWSService
}

// InitResources emits the account-level Cost Optimization Hub enrollment status
// and preferences (singletons, imported by account id) when enrolled.
func (g *CostOptimizationHubGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := costoptimizationhub.NewFromConfig(config)
	ctx := context.TODO()

	account, err := g.getAccountNumber(config)
	if err != nil {
		return err
	}
	accountID := StringValue(account)
	if accountID == "" {
		return nil
	}

	enrolled := false
	if out, err := svc.ListEnrollmentStatuses(ctx, &costoptimizationhub.ListEnrollmentStatusesInput{}); err == nil {
		for _, s := range out.Items {
			if string(s.Status) == "Active" {
				enrolled = true
			}
		}
	}
	if !enrolled {
		return nil
	}
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
		accountID, accountID, "aws_costoptimizationhub_enrollment_status", "aws", defaultAllowEmptyValues))
	if _, err := svc.GetPreferences(ctx, &costoptimizationhub.GetPreferencesInput{}); err == nil {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			accountID, accountID, "aws_costoptimizationhub_preferences", "aws", defaultAllowEmptyValues))
	}
	return nil
}
