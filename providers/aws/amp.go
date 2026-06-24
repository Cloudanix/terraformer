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
	"github.com/aws/aws-sdk-go-v2/service/amp"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type PrometheusGenerator struct {
	AWSService
}

func (g *PrometheusGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := amp.NewFromConfig(config)
	ctx := awsContext()
	var workspaceIDs []string
	p := amp.NewListWorkspacesPaginator(svc, &amp.ListWorkspacesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, item := range page.Workspaces {
			id := StringValue(item.WorkspaceId)
			if id == "" {
				continue
			}
			workspaceIDs = append(workspaceIDs, id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_prometheus_workspace", "aws", defaultAllowEmptyValues))
		}
	}

	for _, wsID := range workspaceIDs {
		ws := wsID
		if _, err := svc.DescribeAlertManagerDefinition(ctx, &amp.DescribeAlertManagerDefinitionInput{WorkspaceId: &ws}); err == nil {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				ws, ws, "aws_prometheus_alert_manager_definition", "aws", defaultAllowEmptyValues))
		}
		for rp := amp.NewListRuleGroupsNamespacesPaginator(svc, &amp.ListRuleGroupsNamespacesInput{WorkspaceId: &ws}); rp.HasMorePages(); {
			page, err := rp.NextPage(ctx)
			if err != nil {
				break
			}
			for _, ns := range page.RuleGroupsNamespaces {
				arn := StringValue(ns.Arn)
				if arn == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					arn, StringValue(ns.Name), "aws_prometheus_rule_group_namespace", "aws", defaultAllowEmptyValues))
			}
		}
	}

	for sp := amp.NewListScrapersPaginator(svc, &amp.ListScrapersInput{}); sp.HasMorePages(); {
		page, err := sp.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, s := range page.Scrapers {
			id := StringValue(s.ScraperId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_prometheus_scraper", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
