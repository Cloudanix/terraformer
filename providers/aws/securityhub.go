// Copyright 2020 The Terraformer Authors.
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
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
)

var securityhubAllowEmptyValues = []string{"tags."}

type SecurityhubGenerator struct {
	AWSService
}

func (g *SecurityhubGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	client := securityhub.NewFromConfig(config)

	account, err := g.getAccountNumber(config)
	if err != nil {
		return err
	}

	accountDisabled, err := g.addAccount(client, *account)
	if accountDisabled {
		return nil
	}
	if err != nil {
		return err
	}
	err = g.addMembers(client, *account)
	if err != nil {
		return err
	}
	err = g.addStandardsSubscription(client, *account)
	if err != nil {
		return err
	}
	if err := g.addActionTargets(client); err != nil {
		return err
	}
	if err := g.addInsights(client); err != nil {
		return err
	}
	if err := g.addFindingAggregators(client); err != nil {
		return err
	}
	if rules, err := client.ListAutomationRules(context.TODO(), &securityhub.ListAutomationRulesInput{}); err == nil {
		for _, r := range rules.AutomationRulesMetadata {
			arn := StringValue(r.RuleArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(r.RuleName), "aws_securityhub_automation_rule", "aws", securityhubAllowEmptyValues))
		}
	}
	for p := securityhub.NewListConfigurationPoliciesPaginator(client, &securityhub.ListConfigurationPoliciesInput{}); p.HasMorePages(); {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			break
		}
		for _, cp := range page.ConfigurationPolicySummaries {
			arn := StringValue(cp.Arn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(cp.Name), "aws_securityhub_configuration_policy", "aws", securityhubAllowEmptyValues))
		}
	}

	ctx := context.TODO()
	for p := securityhub.NewListEnabledProductsForImportPaginator(client, &securityhub.ListEnabledProductsForImportInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, arn := range page.ProductSubscriptions {
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_securityhub_product_subscription", "aws", securityhubAllowEmptyValues))
		}
	}

	for p := securityhub.NewListOrganizationAdminAccountsPaginator(client, &securityhub.ListOrganizationAdminAccountsInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, a := range page.AdminAccounts {
			id := StringValue(a.AccountId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_securityhub_organization_admin_account", "aws", securityhubAllowEmptyValues))
		}
	}

	if _, err := client.DescribeOrganizationConfiguration(ctx, &securityhub.DescribeOrganizationConfigurationInput{}); err == nil {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			*account, *account, "aws_securityhub_organization_configuration", "aws", securityhubAllowEmptyValues))
	}
	return nil
}

func (g *SecurityhubGenerator) addActionTargets(svc *securityhub.Client) error {
	p := securityhub.NewDescribeActionTargetsPaginator(svc, &securityhub.DescribeActionTargetsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, at := range page.ActionTargets {
			arn := StringValue(at.ActionTargetArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(at.Name), "aws_securityhub_action_target", "aws", securityhubAllowEmptyValues))
		}
	}
	return nil
}

func (g *SecurityhubGenerator) addInsights(svc *securityhub.Client) error {
	p := securityhub.NewGetInsightsPaginator(svc, &securityhub.GetInsightsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, insight := range page.Insights {
			arn := StringValue(insight.InsightArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(insight.Name), "aws_securityhub_insight", "aws", securityhubAllowEmptyValues))
		}
	}
	return nil
}

func (g *SecurityhubGenerator) addFindingAggregators(svc *securityhub.Client) error {
	p := securityhub.NewListFindingAggregatorsPaginator(svc, &securityhub.ListFindingAggregatorsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, fa := range page.FindingAggregators {
			arn := StringValue(fa.FindingAggregatorArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_securityhub_finding_aggregator", "aws", securityhubAllowEmptyValues))
		}
	}
	return nil
}

func (g *SecurityhubGenerator) addAccount(client *securityhub.Client, accountNumber string) (bool, error) {
	_, err := client.GetEnabledStandards(context.TODO(), &securityhub.GetEnabledStandardsInput{})

	if err != nil {
		errorMsg := err.Error()
		if !strings.Contains(errorMsg, "not subscribed to AWS Security Hub") {
			return false, err
		}
		return true, nil
	}
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
		accountNumber,
		accountNumber,
		"aws_securityhub_account",
		"aws",
		securityhubAllowEmptyValues,
	))
	return false, nil
}

func (g *SecurityhubGenerator) addMembers(svc *securityhub.Client, accountNumber string) error {
	p := securityhub.NewListMembersPaginator(svc, &securityhub.ListMembersInput{})

	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, member := range page.Members {
			id := *member.AccountId
			attributes := map[string]string{
				"account_id": id,
			}
			if member.Email != nil {
				attributes["email"] = *member.Email
			}
			g.Resources = append(g.Resources, terraformutils.NewResource(
				id,
				"securityhub_member_"+id,
				"aws_securityhub_member",
				"aws",
				attributes,
				securityhubAllowEmptyValues,
				map[string]interface{}{
					"depends_on": []string{"${aws_securityhub_account.tfer--" + accountNumber + "}"},
				},
			))
		}
	}
	return nil
}

func (g *SecurityhubGenerator) addStandardsSubscription(svc *securityhub.Client, accountNumber string) error {
	p := securityhub.NewGetEnabledStandardsPaginator(svc, &securityhub.GetEnabledStandardsInput{})

	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, standardsSubscription := range page.StandardsSubscriptions {
			id := *standardsSubscription.StandardsSubscriptionArn
			g.Resources = append(g.Resources, terraformutils.NewResource(
				id,
				id,
				"aws_securityhub_standards_subscription",
				"aws",
				map[string]string{
					"standards_arn": id,
				},
				securityhubAllowEmptyValues,
				map[string]interface{}{
					"depends_on": []string{"aws_securityhub_account.tfer--" + accountNumber},
				},
			))
			subArn := id
			for cp := securityhub.NewDescribeStandardsControlsPaginator(svc, &securityhub.DescribeStandardsControlsInput{StandardsSubscriptionArn: &subArn}); cp.HasMorePages(); {
				cpage, err := cp.NextPage(context.TODO())
				if err != nil {
					break
				}
				for _, ctrl := range cpage.Controls {
					arn := StringValue(ctrl.StandardsControlArn)
					if arn == "" {
						continue
					}
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						arn, arn, "aws_securityhub_standards_control", "aws", securityhubAllowEmptyValues))
				}
			}
		}
	}
	return nil
}
