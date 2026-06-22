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

	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type SESv2Generator struct {
	AWSService
}

// InitResources enumerates SES v2 email identities (the existing `ses` service
// covers SES v1 only). The Terraform import ID for aws_sesv2_email_identity is
// the identity name (an email address or a domain).
func (g *SESv2Generator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := sesv2.NewFromConfig(config)

	p := sesv2.NewListEmailIdentitiesPaginator(svc, &sesv2.ListEmailIdentitiesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.EmailIdentities, "aws_sesv2_email_identity",
			defaultAllowEmptyValues,
			func(i types.IdentityInfo) string { return StringValue(i.IdentityName) },
			func(i types.IdentityInfo) string { return StringValue(i.IdentityName) })
		for _, i := range page.EmailIdentities {
			name := StringValue(i.IdentityName)
			if name == "" {
				continue
			}
			// Both attribute sets are singletons on the identity, imported by name.
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_sesv2_email_identity_feedback_attributes", "aws", defaultAllowEmptyValues))
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_sesv2_email_identity_mail_from_attributes", "aws", defaultAllowEmptyValues))
		}
	}

	configSets := sesv2.NewListConfigurationSetsPaginator(svc, &sesv2.ListConfigurationSetsInput{})
	for configSets.HasMorePages() {
		page, err := configSets.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, name := range page.ConfigurationSets {
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_sesv2_configuration_set", "aws", defaultAllowEmptyValues))
			setName := name
			if dest, err := svc.GetConfigurationSetEventDestinations(context.TODO(), &sesv2.GetConfigurationSetEventDestinationsInput{ConfigurationSetName: &setName}); err == nil {
				for _, ed := range dest.EventDestinations {
					edName := StringValue(ed.Name)
					if edName == "" {
						continue
					}
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						setName+"|"+edName, setName+"_"+edName, "aws_sesv2_configuration_set_event_destination", "aws", defaultAllowEmptyValues))
				}
			}
		}
	}

	contactLists := sesv2.NewListContactListsPaginator(svc, &sesv2.ListContactListsInput{})
	for contactLists.HasMorePages() {
		page, err := contactLists.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, cl := range page.ContactLists {
			name := StringValue(cl.ContactListName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_sesv2_contact_list", "aws", defaultAllowEmptyValues))
		}
	}

	ipPools := sesv2.NewListDedicatedIpPoolsPaginator(svc, &sesv2.ListDedicatedIpPoolsInput{})
	for ipPools.HasMorePages() {
		page, err := ipPools.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, name := range page.DedicatedIpPools {
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_sesv2_dedicated_ip_pool", "aws", defaultAllowEmptyValues))
		}
	}

	// Account-level singletons keyed by account ID.
	account, err := g.getAccountNumber(config)
	if err != nil {
		return err
	}
	if accountID := StringValue(account); accountID != "" {
		if out, err := svc.GetAccount(context.TODO(), &sesv2.GetAccountInput{}); err == nil {
			if out.SuppressionAttributes != nil {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					accountID, accountID, "aws_sesv2_account_suppression_attributes", "aws", defaultAllowEmptyValues))
			}
			if out.VdmAttributes != nil {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					accountID, accountID, "aws_sesv2_account_vdm_attributes", "aws", defaultAllowEmptyValues))
			}
		}
	}
	return nil
}
