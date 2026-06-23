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

	"github.com/aws/aws-sdk-go-v2/service/computeoptimizer"

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
	out, err := svc.GetEnrollmentStatus(context.TODO(), &computeoptimizer.GetEnrollmentStatusInput{})
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
	return nil
}
