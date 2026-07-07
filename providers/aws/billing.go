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
	"github.com/aws/aws-sdk-go-v2/service/billing"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type BillingGenerator struct {
	AWSService
}

func (g *BillingGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := billing.NewFromConfig(config)
	for p := billing.NewListBillingViewsPaginator(svc, &billing.ListBillingViewsInput{}); p.HasMorePages(); {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, v := range page.BillingViews {
			arn := StringValue(v.Arn)
			if arn == "" {
				continue
			}
			// aws_billing_view Read fetches by the arn attribute, not the ID.
			// Terraformer does a plain Refresh, so seed arn or GetBillingView
			// fails with "Value at 'arn' failed to satisfy constraint" (null arn).
			g.Resources = append(g.Resources, terraformutils.NewResource(
				arn, arn, "aws_billing_view", "aws",
				map[string]string{"arn": arn},
				defaultAllowEmptyValues, map[string]interface{}{}))
		}
	}
	return nil
}
