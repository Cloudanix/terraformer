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
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

var sesAllowEmptyValues = []string{"tags."}

type SesGenerator struct {
	AWSService
}

func (g *SesGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := ses.NewFromConfig(config)

	if err := g.loadDomainIdentities(svc); err != nil {
		return err
	}
	if err := g.loadMailIdentities(svc); err != nil {
		return err
	}
	if err := g.loadTemplates(svc); err != nil {
		return err
	}
	if err := g.loadConfigurationSets(svc); err != nil {
		return err
	}
	if err := g.loadRuleSets(svc); err != nil {
		return err
	}
	if err := g.loadReceiptExtras(svc); err != nil {
		return err
	}

	return nil
}

func (g *SesGenerator) loadReceiptExtras(svc *ses.Client) error {
	ctx := awsContext()
	if filters, err := svc.ListReceiptFilters(ctx, &ses.ListReceiptFiltersInput{}); err == nil {
		for _, f := range filters.Filters {
			name := StringValue(f.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_ses_receipt_filter", "aws", sesAllowEmptyValues))
		}
	}
	if active, err := svc.DescribeActiveReceiptRuleSet(ctx, &ses.DescribeActiveReceiptRuleSetInput{}); err == nil && active.Metadata != nil {
		name := StringValue(active.Metadata.Name)
		if name != "" {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_ses_active_receipt_rule_set", "aws", sesAllowEmptyValues))
		}
	}
	return nil
}

func (g *SesGenerator) loadDomainIdentities(svc *ses.Client) error {
	p := ses.NewListIdentitiesPaginator(svc, &ses.ListIdentitiesInput{
		IdentityType: "Domain",
	})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, identity := range page.Identities {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				identity,
				identity,
				"aws_ses_domain_identity",
				"aws",
				sesAllowEmptyValues))
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				identity, identity, "aws_ses_domain_dkim", "aws", sesAllowEmptyValues))
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				identity, identity, "aws_ses_domain_identity_verification", "aws", sesAllowEmptyValues))
			if attrs, err := svc.GetIdentityMailFromDomainAttributes(awsContext(), &ses.GetIdentityMailFromDomainAttributesInput{Identities: []string{identity}}); err == nil {
				if a, ok := attrs.MailFromDomainAttributes[identity]; ok && StringValue(a.MailFromDomain) != "" {
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						identity, identity, "aws_ses_domain_mail_from", "aws", sesAllowEmptyValues))
				}
			}
			if na, err := svc.GetIdentityNotificationAttributes(awsContext(), &ses.GetIdentityNotificationAttributesInput{Identities: []string{identity}}); err == nil {
				if a, ok := na.NotificationAttributes[identity]; ok {
					for notifType, topic := range map[string]*string{
						"Bounce": a.BounceTopic, "Complaint": a.ComplaintTopic, "Delivery": a.DeliveryTopic,
					} {
						if StringValue(topic) == "" {
							continue
						}
						g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
							identity+"|"+notifType, identity+"_"+notifType, "aws_ses_identity_notification_topic", "aws", sesAllowEmptyValues))
					}
				}
			}
		}
	}
	return nil
}

func (g *SesGenerator) loadMailIdentities(svc *ses.Client) error {
	p := ses.NewListIdentitiesPaginator(svc, &ses.ListIdentitiesInput{
		IdentityType: "EmailAddress",
	})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, identity := range page.Identities {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				identity,
				identity,
				"aws_ses_email_identity",
				"aws",
				sesAllowEmptyValues))
		}
	}
	return nil
}

func (g *SesGenerator) loadTemplates(svc *ses.Client) error {
	templates, err := svc.ListTemplates(awsContext(), &ses.ListTemplatesInput{})
	if err != nil {
		return err
	}

	for _, templateMetadata := range templates.TemplatesMetadata {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			StringValue(templateMetadata.Name),
			StringValue(templateMetadata.Name),
			"aws_ses_template",
			"aws",
			sesAllowEmptyValues))
	}
	return nil
}

func (g *SesGenerator) loadConfigurationSets(svc *ses.Client) error {
	configurationSets, err := svc.ListConfigurationSets(awsContext(), &ses.ListConfigurationSetsInput{})
	if err != nil {
		return err
	}

	for _, configurationSet := range configurationSets.ConfigurationSets {
		setName := StringValue(configurationSet.Name)
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			setName,
			setName,
			"aws_ses_configuration_set",
			"aws",
			sesAllowEmptyValues))
		if desc, err := svc.DescribeConfigurationSet(awsContext(), &ses.DescribeConfigurationSetInput{
			ConfigurationSetName: configurationSet.Name,
			ConfigurationSetAttributeNames: []types.ConfigurationSetAttribute{
				types.ConfigurationSetAttributeEventDestinations,
			},
		}); err == nil {
			for _, ed := range desc.EventDestinations {
				name := StringValue(ed.Name)
				if name == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					setName+"/"+name, setName+"_"+name, "aws_ses_event_destination", "aws", sesAllowEmptyValues))
			}
		}
	}
	return nil
}

func (g *SesGenerator) loadRuleSets(svc *ses.Client) error {
	ruleSets, err := svc.ListReceiptRuleSets(awsContext(), &ses.ListReceiptRuleSetsInput{})
	if err != nil {
		return err
	}

	for _, ruleSet := range ruleSets.RuleSets {
		ruleSetName := StringValue(ruleSet.Name)
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			ruleSetName,
			ruleSetName,
			"aws_ses_receipt_rule_set",
			"aws",
			sesAllowEmptyValues))
		rules, err := svc.DescribeReceiptRuleSet(awsContext(), &ses.DescribeReceiptRuleSetInput{
			RuleSetName: ruleSet.Name,
		})
		if err != nil {
			return err
		}
		for _, rule := range rules.Rules {
			ruleID := ruleSetName + ":" + *rule.Name
			g.Resources = append(g.Resources, terraformutils.NewResource(
				*rule.Name,
				ruleID,
				"aws_ses_receipt_rule",
				"aws",
				map[string]string{
					"name":          *rule.Name,
					"rule_set_name": ruleSetName,
				},
				sesAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}
