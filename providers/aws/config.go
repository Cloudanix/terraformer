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

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
)

var configAllowEmptyValues = []string{"tags."}

type ConfigGenerator struct {
	AWSService
}

func (g *ConfigGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	client := configservice.NewFromConfig(config)

	configurationRecorderRefs, err := g.addConfigurationRecorders(client)
	if err != nil {
		return err
	}
	err = g.addConfigRules(client, configurationRecorderRefs)
	if err != nil {
		return err
	}
	err = g.addDeliveryChannels(client, configurationRecorderRefs)
	if err != nil {
		return err
	}
	return g.addConfigExtras(client)
}

func (g *ConfigGenerator) addConfigExtras(svc *configservice.Client) error {
	ctx := context.TODO()
	for p := configservice.NewDescribeConfigurationAggregatorsPaginator(svc, &configservice.DescribeConfigurationAggregatorsInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, a := range page.ConfigurationAggregators {
			name := StringValue(a.ConfigurationAggregatorName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_config_configuration_aggregator", "aws", configAllowEmptyValues))
		}
	}
	for p := configservice.NewDescribeConformancePacksPaginator(svc, &configservice.DescribeConformancePacksInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, cp := range page.ConformancePackDetails {
			name := StringValue(cp.ConformancePackName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_config_conformance_pack", "aws", configAllowEmptyValues))
		}
	}
	for p := configservice.NewDescribeRetentionConfigurationsPaginator(svc, &configservice.DescribeRetentionConfigurationsInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, rc := range page.RetentionConfigurations {
			name := StringValue(rc.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_config_retention_configuration", "aws", configAllowEmptyValues))
		}
	}
	for p := configservice.NewDescribeAggregationAuthorizationsPaginator(svc, &configservice.DescribeAggregationAuthorizationsInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, a := range page.AggregationAuthorizations {
			acct := StringValue(a.AuthorizedAccountId)
			region := StringValue(a.AuthorizedAwsRegion)
			if acct == "" || region == "" {
				continue
			}
			id := acct + "/" + region
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, acct+"_"+region, "aws_config_aggregate_authorization", "aws", configAllowEmptyValues))
		}
	}
	for p := configservice.NewDescribeOrganizationConformancePacksPaginator(svc, &configservice.DescribeOrganizationConformancePacksInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, ocp := range page.OrganizationConformancePacks {
			name := StringValue(ocp.OrganizationConformancePackName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_config_organization_conformance_pack", "aws", configAllowEmptyValues))
		}
	}
	for p := configservice.NewDescribeOrganizationConfigRulesPaginator(svc, &configservice.DescribeOrganizationConfigRulesInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, rule := range page.OrganizationConfigRules {
			name := StringValue(rule.OrganizationConfigRuleName)
			if name == "" {
				continue
			}
			tfType := ""
			switch {
			case rule.OrganizationCustomPolicyRuleMetadata != nil:
				tfType = "aws_config_organization_custom_policy_rule"
			case rule.OrganizationCustomRuleMetadata != nil:
				tfType = "aws_config_organization_custom_rule"
			case rule.OrganizationManagedRuleMetadata != nil:
				tfType = "aws_config_organization_managed_rule"
			default:
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, tfType, "aws", configAllowEmptyValues))
		}
	}
	return nil
}

func (g *ConfigGenerator) addConfigurationRecorders(svc *configservice.Client) ([]string, error) {
	configurationRecorders, err := svc.DescribeConfigurationRecorders(context.TODO(),
		&configservice.DescribeConfigurationRecordersInput{})

	if err != nil {
		return nil, err
	}
	var configurationRecorderRefs []string
	for _, configurationRecorder := range configurationRecorders.ConfigurationRecorders {
		name := *configurationRecorder.Name
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			name,
			name,
			"aws_config_configuration_recorder",
			"aws",
			configAllowEmptyValues,
		))
		configurationRecorderRefs = append(configurationRecorderRefs,
			"aws_config_configuration_recorder.tfer--"+name)
	}
	return configurationRecorderRefs, nil
}

func (g *ConfigGenerator) addConfigRules(svc *configservice.Client, configurationRecorderRefs []string) error {
	var nextToken *string

	for {
		configRules, err := svc.DescribeConfigRules(
			context.TODO(),
			&configservice.DescribeConfigRulesInput{
				NextToken: nextToken,
			})

		if err != nil {
			return err
		}
		for _, configRule := range configRules.ConfigRules {
			name := *configRule.ConfigRuleName
			g.Resources = append(g.Resources, terraformutils.NewResource(
				name,
				name,
				"aws_config_config_rule",
				"aws",
				map[string]string{},
				configAllowEmptyValues,
				map[string]interface{}{
					"depends_on": configurationRecorderRefs,
				},
			))
		}
		nextToken = configRules.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *ConfigGenerator) addDeliveryChannels(svc *configservice.Client, configurationRecorderRefs []string) error {
	deliveryChannels, err := svc.DescribeDeliveryChannels(context.TODO(),
		&configservice.DescribeDeliveryChannelsInput{})

	if err != nil {
		return err
	}
	for _, deliveryChannel := range deliveryChannels.DeliveryChannels {
		name := *deliveryChannel.Name
		g.Resources = append(g.Resources, terraformutils.NewResource(
			name,
			name,
			"aws_config_delivery_channel",
			"aws",
			map[string]string{},
			configAllowEmptyValues,
			map[string]interface{}{
				"depends_on": configurationRecorderRefs,
			},
		))
	}
	return nil
}
