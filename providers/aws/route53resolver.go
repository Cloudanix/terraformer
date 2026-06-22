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

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/route53resolver"
	"github.com/aws/aws-sdk-go-v2/service/route53resolver/types"
)

type Route53ResolverGenerator struct {
	AWSService
}

// InitResources enumerates Route 53 Resolver resources (endpoints, rules, DNS
// Firewall, query logging, DNSSEC). Every resource's Terraform import ID is its
// Resolver-assigned Id, which is also used as the resource name for uniqueness.
func (g *Route53ResolverGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := route53resolver.NewFromConfig(config)
	ctx := context.TODO()

	endpoints := route53resolver.NewListResolverEndpointsPaginator(svc, &route53resolver.ListResolverEndpointsInput{})
	for endpoints.HasMorePages() {
		page, err := endpoints.NextPage(ctx)
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.ResolverEndpoints, "aws_route53_resolver_endpoint",
			defaultAllowEmptyValues,
			func(r types.ResolverEndpoint) string { return StringValue(r.Id) },
			func(r types.ResolverEndpoint) string { return StringValue(r.Id) })
	}

	rules := route53resolver.NewListResolverRulesPaginator(svc, &route53resolver.ListResolverRulesInput{})
	for rules.HasMorePages() {
		page, err := rules.NextPage(ctx)
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.ResolverRules, "aws_route53_resolver_rule",
			defaultAllowEmptyValues,
			func(r types.ResolverRule) string { return StringValue(r.Id) },
			func(r types.ResolverRule) string { return StringValue(r.Id) })
	}

	ruleAssocs := route53resolver.NewListResolverRuleAssociationsPaginator(svc, &route53resolver.ListResolverRuleAssociationsInput{})
	for ruleAssocs.HasMorePages() {
		page, err := ruleAssocs.NextPage(ctx)
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.ResolverRuleAssociations, "aws_route53_resolver_rule_association",
			defaultAllowEmptyValues,
			func(r types.ResolverRuleAssociation) string { return StringValue(r.Id) },
			func(r types.ResolverRuleAssociation) string { return StringValue(r.Id) })
	}

	queryLogConfigs := route53resolver.NewListResolverQueryLogConfigsPaginator(svc, &route53resolver.ListResolverQueryLogConfigsInput{})
	for queryLogConfigs.HasMorePages() {
		page, err := queryLogConfigs.NextPage(ctx)
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.ResolverQueryLogConfigs, "aws_route53_resolver_query_log_config",
			defaultAllowEmptyValues,
			func(r types.ResolverQueryLogConfig) string { return StringValue(r.Id) },
			func(r types.ResolverQueryLogConfig) string { return StringValue(r.Id) })
	}

	queryLogAssocs := route53resolver.NewListResolverQueryLogConfigAssociationsPaginator(svc, &route53resolver.ListResolverQueryLogConfigAssociationsInput{})
	for queryLogAssocs.HasMorePages() {
		page, err := queryLogAssocs.NextPage(ctx)
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.ResolverQueryLogConfigAssociations, "aws_route53_resolver_query_log_config_association",
			defaultAllowEmptyValues,
			func(r types.ResolverQueryLogConfigAssociation) string { return StringValue(r.Id) },
			func(r types.ResolverQueryLogConfigAssociation) string { return StringValue(r.Id) })
	}

	dnssecConfigs := route53resolver.NewListResolverDnssecConfigsPaginator(svc, &route53resolver.ListResolverDnssecConfigsInput{})
	for dnssecConfigs.HasMorePages() {
		page, err := dnssecConfigs.NextPage(ctx)
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.ResolverDnssecConfigs, "aws_route53_resolver_dnssec_config",
			defaultAllowEmptyValues,
			func(r types.ResolverDnssecConfig) string { return StringValue(r.Id) },
			func(r types.ResolverDnssecConfig) string { return StringValue(r.Id) })
	}

	var ruleGroupIDs []string
	firewallRuleGroups := route53resolver.NewListFirewallRuleGroupsPaginator(svc, &route53resolver.ListFirewallRuleGroupsInput{})
	for firewallRuleGroups.HasMorePages() {
		page, err := firewallRuleGroups.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, rg := range page.FirewallRuleGroups {
			ruleGroupIDs = append(ruleGroupIDs, StringValue(rg.Id))
		}
		g.Resources = appendSimpleResources(g.Resources, page.FirewallRuleGroups, "aws_route53_resolver_firewall_rule_group",
			defaultAllowEmptyValues,
			func(r types.FirewallRuleGroupMetadata) string { return StringValue(r.Id) },
			func(r types.FirewallRuleGroupMetadata) string { return StringValue(r.Id) })
	}

	for _, groupID := range ruleGroupIDs {
		if groupID == "" {
			continue
		}
		rules := route53resolver.NewListFirewallRulesPaginator(svc, &route53resolver.ListFirewallRulesInput{FirewallRuleGroupId: &groupID})
		for rules.HasMorePages() {
			page, err := rules.NextPage(ctx)
			if err != nil {
				break
			}
			for _, r := range page.FirewallRules {
				domainListID := StringValue(r.FirewallDomainListId)
				if domainListID == "" {
					continue
				}
				id := groupID + ":" + domainListID
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, id, "aws_route53_resolver_firewall_rule", "aws", defaultAllowEmptyValues))
			}
		}
	}

	firewallConfigs := route53resolver.NewListFirewallConfigsPaginator(svc, &route53resolver.ListFirewallConfigsInput{})
	for firewallConfigs.HasMorePages() {
		page, err := firewallConfigs.NextPage(ctx)
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.FirewallConfigs, "aws_route53_resolver_firewall_config",
			defaultAllowEmptyValues,
			func(r types.FirewallConfig) string { return StringValue(r.Id) },
			func(r types.FirewallConfig) string { return StringValue(r.Id) })
	}

	resolverConfigs := route53resolver.NewListResolverConfigsPaginator(svc, &route53resolver.ListResolverConfigsInput{})
	for resolverConfigs.HasMorePages() {
		page, err := resolverConfigs.NextPage(ctx)
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.ResolverConfigs, "aws_route53_resolver_config",
			defaultAllowEmptyValues,
			func(r types.ResolverConfig) string { return StringValue(r.Id) },
			func(r types.ResolverConfig) string { return StringValue(r.Id) })
	}

	firewallDomainLists := route53resolver.NewListFirewallDomainListsPaginator(svc, &route53resolver.ListFirewallDomainListsInput{})
	for firewallDomainLists.HasMorePages() {
		page, err := firewallDomainLists.NextPage(ctx)
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.FirewallDomainLists, "aws_route53_resolver_firewall_domain_list",
			defaultAllowEmptyValues,
			func(r types.FirewallDomainListMetadata) string { return StringValue(r.Id) },
			func(r types.FirewallDomainListMetadata) string { return StringValue(r.Id) })
	}

	firewallRuleGroupAssocs := route53resolver.NewListFirewallRuleGroupAssociationsPaginator(svc, &route53resolver.ListFirewallRuleGroupAssociationsInput{})
	for firewallRuleGroupAssocs.HasMorePages() {
		page, err := firewallRuleGroupAssocs.NextPage(ctx)
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.FirewallRuleGroupAssociations, "aws_route53_resolver_firewall_rule_group_association",
			defaultAllowEmptyValues,
			func(r types.FirewallRuleGroupAssociation) string { return StringValue(r.Id) },
			func(r types.FirewallRuleGroupAssociation) string { return StringValue(r.Id) })
	}

	return nil
}
