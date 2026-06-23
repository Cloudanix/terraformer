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

	"github.com/aws/aws-sdk-go-v2/service/savingsplans"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type SavingsPlansGenerator struct {
	AWSService
}

func (g *SavingsPlansGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := savingsplans.NewFromConfig(config)
	out, err := svc.DescribeSavingsPlans(context.TODO(), &savingsplans.DescribeSavingsPlansInput{})
	if err != nil {
		return err
	}
	for _, sp := range out.SavingsPlans {
		id := StringValue(sp.SavingsPlanId)
		if id == "" {
			continue
		}
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			id, id, "aws_savingsplans_savings_plan", "aws", defaultAllowEmptyValues))
	}
	return nil
}
