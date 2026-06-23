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
	"github.com/aws/aws-sdk-go-v2/service/inspector2"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type Inspector2Generator struct {
	AWSService
}

// InitResources enumerates Inspector2 delegated admin accounts, member
// associations, and filters. Import IDs: account id for admin/member, filter ARN.
func (g *Inspector2Generator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := inspector2.NewFromConfig(config)
	ctx := awsContext()

	admins := inspector2.NewListDelegatedAdminAccountsPaginator(svc, &inspector2.ListDelegatedAdminAccountsInput{})
	for admins.HasMorePages() {
		page, err := admins.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, admin := range page.DelegatedAdminAccounts {
			id := StringValue(admin.AccountId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_inspector2_delegated_admin_account", "aws", defaultAllowEmptyValues))
		}
	}

	members := inspector2.NewListMembersPaginator(svc, &inspector2.ListMembersInput{})
	for members.HasMorePages() {
		page, err := members.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, member := range page.Members {
			id := StringValue(member.AccountId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_inspector2_member_association", "aws", defaultAllowEmptyValues))
		}
	}

	filters := inspector2.NewListFiltersPaginator(svc, &inspector2.ListFiltersInput{})
	for filters.HasMorePages() {
		page, err := filters.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, filter := range page.Filters {
			arn := StringValue(filter.Arn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_inspector2_filter", "aws", defaultAllowEmptyValues))
		}
	}

	account, err := g.getAccountNumber(config)
	if err != nil {
		return err
	}
	accountID := StringValue(account)
	if accountID != "" {
		if status, err := svc.BatchGetAccountStatus(ctx, &inspector2.BatchGetAccountStatusInput{AccountIds: []string{accountID}}); err == nil && len(status.Accounts) > 0 {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				accountID, accountID, "aws_inspector2_enabler", "aws", defaultAllowEmptyValues))
		}
		if _, err := svc.DescribeOrganizationConfiguration(ctx, &inspector2.DescribeOrganizationConfigurationInput{}); err == nil {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				accountID, accountID, "aws_inspector2_organization_configuration", "aws", defaultAllowEmptyValues))
		}
	}

	return nil
}
