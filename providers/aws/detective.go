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

	"github.com/aws/aws-sdk-go-v2/service/detective"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type DetectiveGenerator struct {
	AWSService
}

// InitResources enumerates Detective behavior graphs. Import ID is the graph ARN.
func (g *DetectiveGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := detective.NewFromConfig(config)

	ctx := context.TODO()
	var graphArns []string
	p := detective.NewListGraphsPaginator(svc, &detective.ListGraphsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, graph := range page.GraphList {
			arn := StringValue(graph.Arn)
			if arn == "" {
				continue
			}
			graphArns = append(graphArns, arn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_detective_graph", "aws", defaultAllowEmptyValues))
		}
	}

	for _, graphArn := range graphArns {
		ga := graphArn
		for mp := detective.NewListMembersPaginator(svc, &detective.ListMembersInput{GraphArn: &ga}); mp.HasMorePages(); {
			page, err := mp.NextPage(ctx)
			if err != nil {
				break
			}
			for _, m := range page.MemberDetails {
				acct := StringValue(m.AccountId)
				if acct == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					ga+"/"+acct, acct, "aws_detective_member", "aws", defaultAllowEmptyValues))
			}
		}
		if _, err := svc.DescribeOrganizationConfiguration(ctx, &detective.DescribeOrganizationConfigurationInput{GraphArn: &ga}); err == nil {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				ga, ga, "aws_detective_organization_configuration", "aws", defaultAllowEmptyValues))
		}
	}

	if admins, err := svc.ListOrganizationAdminAccounts(ctx, &detective.ListOrganizationAdminAccountsInput{}); err == nil {
		for _, a := range admins.Administrators {
			id := StringValue(a.AccountId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_detective_organization_admin_account", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
