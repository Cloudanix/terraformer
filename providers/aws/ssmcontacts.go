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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssmcontacts"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type SSMContactsGenerator struct {
	AWSService
}

// InitResources enumerates SSM Contacts contacts. Import ID is the contact ARN.
func (g *SSMContactsGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := ssmcontacts.NewFromConfig(config)

	ctx := awsContext()
	var contactArns []string
	p := ssmcontacts.NewListContactsPaginator(svc, &ssmcontacts.ListContactsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, contact := range page.Contacts {
			arn := StringValue(contact.ContactArn)
			if arn == "" {
				continue
			}
			contactArns = append(contactArns, arn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(contact.Alias), "aws_ssmcontacts_contact", "aws", defaultAllowEmptyValues))
			// The escalation/engagement plan is a singleton on the contact.
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(contact.Alias), "aws_ssmcontacts_plan", "aws", defaultAllowEmptyValues))
		}
	}

	for _, contactArn := range contactArns {
		ca := contactArn
		for cp := ssmcontacts.NewListContactChannelsPaginator(svc, &ssmcontacts.ListContactChannelsInput{ContactId: aws.String(ca)}); cp.HasMorePages(); {
			page, err := cp.NextPage(ctx)
			if err != nil {
				break
			}
			for _, ch := range page.ContactChannels {
				arn := StringValue(ch.ContactChannelArn)
				if arn == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					arn, arn, "aws_ssmcontacts_contact_channel", "aws", defaultAllowEmptyValues))
			}
		}
	}

	for rp := ssmcontacts.NewListRotationsPaginator(svc, &ssmcontacts.ListRotationsInput{}); rp.HasMorePages(); {
		page, err := rp.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, r := range page.Rotations {
			arn := StringValue(r.RotationArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_ssmcontacts_rotation", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
