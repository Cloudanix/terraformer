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

	"github.com/aws/aws-sdk-go-v2/service/costexplorer"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type CostExplorerGenerator struct {
	AWSService
}

// InitResources enumerates Cost Explorer anomaly monitors and cost categories.
// Import IDs are the resource ARN.
func (g *CostExplorerGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := costexplorer.NewFromConfig(config)
	ctx := context.TODO()

	monitors := costexplorer.NewGetAnomalyMonitorsPaginator(svc, &costexplorer.GetAnomalyMonitorsInput{})
	for monitors.HasMorePages() {
		page, err := monitors.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, m := range page.AnomalyMonitors {
			arn := StringValue(m.MonitorArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(m.MonitorName), "aws_ce_anomaly_monitor", "aws", defaultAllowEmptyValues))
		}
	}

	subs := costexplorer.NewGetAnomalySubscriptionsPaginator(svc, &costexplorer.GetAnomalySubscriptionsInput{})
	for subs.HasMorePages() {
		page, err := subs.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, s := range page.AnomalySubscriptions {
			arn := StringValue(s.SubscriptionArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(s.SubscriptionName), "aws_ce_anomaly_subscription", "aws", defaultAllowEmptyValues))
		}
	}

	categories := costexplorer.NewListCostCategoryDefinitionsPaginator(svc, &costexplorer.ListCostCategoryDefinitionsInput{})
	for categories.HasMorePages() {
		page, err := categories.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, c := range page.CostCategoryReferences {
			arn := StringValue(c.CostCategoryArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(c.Name), "aws_ce_cost_category", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
