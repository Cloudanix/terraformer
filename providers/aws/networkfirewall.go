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

	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type NetworkFirewallGenerator struct {
	AWSService
}

// InitResources enumerates Network Firewall firewalls, policies, and rule
// groups. Import IDs are the resource ARN.
func (g *NetworkFirewallGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := networkfirewall.NewFromConfig(config)
	ctx := context.TODO()

	firewalls := networkfirewall.NewListFirewallsPaginator(svc, &networkfirewall.ListFirewallsInput{})
	for firewalls.HasMorePages() {
		page, err := firewalls.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, fw := range page.Firewalls {
			arn := StringValue(fw.FirewallArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_networkfirewall_firewall", "aws", defaultAllowEmptyValues))
			if lc, err := svc.DescribeLoggingConfiguration(ctx, &networkfirewall.DescribeLoggingConfigurationInput{FirewallArn: fw.FirewallArn}); err == nil &&
				lc.LoggingConfiguration != nil && len(lc.LoggingConfiguration.LogDestinationConfigs) > 0 {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					arn, arn, "aws_networkfirewall_logging_configuration", "aws", defaultAllowEmptyValues))
			}
		}
	}

	tlsConfigs := networkfirewall.NewListTLSInspectionConfigurationsPaginator(svc, &networkfirewall.ListTLSInspectionConfigurationsInput{})
	for tlsConfigs.HasMorePages() {
		page, err := tlsConfigs.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, tc := range page.TLSInspectionConfigurations {
			arn := StringValue(tc.Arn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_networkfirewall_tls_inspection_configuration", "aws", defaultAllowEmptyValues))
		}
	}

	policies := networkfirewall.NewListFirewallPoliciesPaginator(svc, &networkfirewall.ListFirewallPoliciesInput{})
	for policies.HasMorePages() {
		page, err := policies.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, fp := range page.FirewallPolicies {
			arn := StringValue(fp.Arn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_networkfirewall_firewall_policy", "aws", defaultAllowEmptyValues))
			arnCopy := arn
			if rp, err := svc.DescribeResourcePolicy(ctx, &networkfirewall.DescribeResourcePolicyInput{ResourceArn: &arnCopy}); err == nil && StringValue(rp.Policy) != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					arn, arn, "aws_networkfirewall_resource_policy", "aws", defaultAllowEmptyValues))
			}
		}
	}

	ruleGroups := networkfirewall.NewListRuleGroupsPaginator(svc, &networkfirewall.ListRuleGroupsInput{})
	for ruleGroups.HasMorePages() {
		page, err := ruleGroups.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, rg := range page.RuleGroups {
			arn := StringValue(rg.Arn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_networkfirewall_rule_group", "aws", defaultAllowEmptyValues))
			arnCopy := arn
			if rp, err := svc.DescribeResourcePolicy(ctx, &networkfirewall.DescribeResourcePolicyInput{ResourceArn: &arnCopy}); err == nil && StringValue(rp.Policy) != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					arn, arn, "aws_networkfirewall_resource_policy", "aws", defaultAllowEmptyValues))
			}
		}
	}
	return nil
}
