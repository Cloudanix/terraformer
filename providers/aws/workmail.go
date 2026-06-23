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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workmail"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type WorkMailGenerator struct {
	AWSService
}

// InitResources enumerates WorkMail organizations and their groups/users.
// Import IDs: organization id; "<org_id>/<group_id>"; "<org_id>/<user_id>".
func (g *WorkMailGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := workmail.NewFromConfig(config)
	ctx := context.TODO()
	var orgIDs []string
	for p := workmail.NewListOrganizationsPaginator(svc, &workmail.ListOrganizationsInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, o := range page.OrganizationSummaries {
			id := StringValue(o.OrganizationId)
			if id == "" {
				continue
			}
			orgIDs = append(orgIDs, id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_workmail_organization", "aws", defaultAllowEmptyValues))
		}
	}
	for _, orgID := range orgIDs {
		oid := orgID
		for p := workmail.NewListGroupsPaginator(svc, &workmail.ListGroupsInput{OrganizationId: aws.String(oid)}); p.HasMorePages(); {
			page, err := p.NextPage(ctx)
			if err != nil {
				break
			}
			for _, grp := range page.Groups {
				gid := StringValue(grp.Id)
				if gid == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					oid+"/"+gid, oid+"_"+gid, "aws_workmail_group", "aws", defaultAllowEmptyValues))
			}
		}
		for p := workmail.NewListUsersPaginator(svc, &workmail.ListUsersInput{OrganizationId: aws.String(oid)}); p.HasMorePages(); {
			page, err := p.NextPage(ctx)
			if err != nil {
				break
			}
			for _, u := range page.Users {
				uid := StringValue(u.Id)
				if uid == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					oid+"/"+uid, oid+"_"+uid, "aws_workmail_user", "aws", defaultAllowEmptyValues))
			}
		}
		for p := workmail.NewListMailDomainsPaginator(svc, &workmail.ListMailDomainsInput{OrganizationId: aws.String(oid)}); p.HasMorePages(); {
			page, err := p.NextPage(ctx)
			if err != nil {
				break
			}
			for _, d := range page.MailDomains {
				domain := StringValue(d.DomainName)
				if domain == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					oid+","+domain, oid+"_"+domain, "aws_workmail_domain", "aws", defaultAllowEmptyValues))
			}
		}
	}
	return nil
}
