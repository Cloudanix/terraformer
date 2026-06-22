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
	"github.com/aws/aws-sdk-go-v2/service/iot"
)

var iotAllowEmptyValues = []string{"tags."}

type IotGenerator struct {
	AWSService
}

func (g *IotGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := iot.NewFromConfig(config)

	if err := g.loadThingTypes(svc); err != nil {
		return err
	}
	if err := g.loadThings(svc); err != nil {
		return err
	}
	if err := g.loadTopicRules(svc); err != nil {
		return err
	}
	if err := g.loadRoleAliases(svc); err != nil {
		return err
	}
	g.loadIotExtras(svc)

	return nil
}

// loadIotExtras enumerates additional top-level IoT resources, each a List*
// paginator returning a name. Import ID is the name.
func (g *IotGenerator) loadIotExtras(svc *iot.Client) {
	ctx := context.TODO()
	add := func(name, tfType string) {
		if name != "" {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, tfType, "aws", iotAllowEmptyValues))
		}
	}
	for p := iot.NewListAuthorizersPaginator(svc, &iot.ListAuthorizersInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range pg.Authorizers {
			add(StringValue(x.AuthorizerName), "aws_iot_authorizer")
		}
	}
	for p := iot.NewListBillingGroupsPaginator(svc, &iot.ListBillingGroupsInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range pg.BillingGroups {
			add(StringValue(x.GroupName), "aws_iot_billing_group")
		}
	}
	for p := iot.NewListPoliciesPaginator(svc, &iot.ListPoliciesInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range pg.Policies {
			add(StringValue(x.PolicyName), "aws_iot_policy")
		}
	}
	for p := iot.NewListThingGroupsPaginator(svc, &iot.ListThingGroupsInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range pg.ThingGroups {
			add(StringValue(x.GroupName), "aws_iot_thing_group")
		}
	}
	for p := iot.NewListProvisioningTemplatesPaginator(svc, &iot.ListProvisioningTemplatesInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range pg.Templates {
			add(StringValue(x.TemplateName), "aws_iot_provisioning_template")
		}
	}
	for p := iot.NewListDomainConfigurationsPaginator(svc, &iot.ListDomainConfigurationsInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range pg.DomainConfigurations {
			add(StringValue(x.DomainConfigurationName), "aws_iot_domain_configuration")
		}
	}
	for p := iot.NewListCertificatesPaginator(svc, &iot.ListCertificatesInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range pg.Certificates {
			add(StringValue(x.CertificateId), "aws_iot_certificate")
		}
	}
	for p := iot.NewListCACertificatesPaginator(svc, &iot.ListCACertificatesInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range pg.Certificates {
			add(StringValue(x.CertificateId), "aws_iot_ca_certificate")
		}
	}
	for p := iot.NewListTopicRuleDestinationsPaginator(svc, &iot.ListTopicRuleDestinationsInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range pg.DestinationSummaries {
			add(StringValue(x.Arn), "aws_iot_topic_rule_destination")
		}
	}
}

func (g *IotGenerator) loadThingTypes(svc *iot.Client) error {
	output, err := svc.ListThingTypes(context.TODO(), &iot.ListThingTypesInput{})
	if err != nil {
		return err
	}
	for _, thingType := range output.ThingTypes {
		g.Resources = append(g.Resources, terraformutils.NewResource(
			*thingType.ThingTypeName,
			*thingType.ThingTypeName,
			"aws_iot_thing_type",
			"aws",
			map[string]string{
				"name": *thingType.ThingTypeName,
			},
			iotAllowEmptyValues,
			map[string]interface{}{},
		))
	}
	return nil
}

func (g *IotGenerator) loadThings(svc *iot.Client) error {
	output, err := svc.ListThings(context.TODO(), &iot.ListThingsInput{})
	if err != nil {
		return err
	}
	for _, thing := range output.Things {
		g.Resources = append(g.Resources, terraformutils.NewResource(
			*thing.ThingName,
			*thing.ThingName,
			"aws_iot_thing",
			"aws",
			map[string]string{
				"name": *thing.ThingName,
			},
			iotAllowEmptyValues,
			map[string]interface{}{},
		))
	}
	return nil
}

func (g *IotGenerator) loadTopicRules(svc *iot.Client) error {
	output, err := svc.ListTopicRules(context.TODO(), &iot.ListTopicRulesInput{})
	if err != nil {
		return err
	}
	for _, rule := range output.Rules {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			*rule.RuleName,
			*rule.RuleName,
			"aws_iot_topic_rule",
			"aws",
			iotAllowEmptyValues))
	}
	return nil
}

func (g *IotGenerator) loadRoleAliases(svc *iot.Client) error {
	output, err := svc.ListRoleAliases(context.TODO(), &iot.ListRoleAliasesInput{})
	if err != nil {
		return err
	}
	for _, roleAlias := range output.RoleAliases {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			roleAlias,
			roleAlias,
			"aws_iot_role_alias",
			"aws",
			iotAllowEmptyValues))
	}
	return nil
}
