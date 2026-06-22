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

	"github.com/aws/aws-sdk-go-v2/service/auditmanager"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager/types"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type AuditManagerGenerator struct {
	AWSService
}

func (g *AuditManagerGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := auditmanager.NewFromConfig(config)
	p := auditmanager.NewListAssessmentsPaginator(svc, &auditmanager.ListAssessmentsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, item := range page.AssessmentMetadata {
			id := StringValue(item.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(item.Name), "aws_auditmanager_assessment", "aws", defaultAllowEmptyValues))
		}
	}

	ctx := context.TODO()
	for cp := auditmanager.NewListControlsPaginator(svc, &auditmanager.ListControlsInput{ControlType: types.ControlTypeCustom}); cp.HasMorePages(); {
		page, err := cp.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, c := range page.ControlMetadataList {
			id := StringValue(c.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(c.Name), "aws_auditmanager_control", "aws", defaultAllowEmptyValues))
		}
	}
	for fp := auditmanager.NewListAssessmentFrameworksPaginator(svc, &auditmanager.ListAssessmentFrameworksInput{FrameworkType: types.FrameworkTypeCustom}); fp.HasMorePages(); {
		page, err := fp.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, f := range page.FrameworkMetadataList {
			id := StringValue(f.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(f.Name), "aws_auditmanager_framework", "aws", defaultAllowEmptyValues))
		}
	}
	account, err := g.getAccountNumber(config)
	if err != nil {
		return err
	}
	if accountID := StringValue(account); accountID != "" {
		if st, err := svc.GetAccountStatus(ctx, &auditmanager.GetAccountStatusInput{}); err == nil && st.Status == types.AccountStatusActive {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				accountID, accountID, "aws_auditmanager_account_registration", "aws", defaultAllowEmptyValues))
		}
		if oa, err := svc.GetOrganizationAdminAccount(ctx, &auditmanager.GetOrganizationAdminAccountInput{}); err == nil && StringValue(oa.AdminAccountId) != "" {
			adminID := StringValue(oa.AdminAccountId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				adminID, adminID, "aws_auditmanager_organization_admin_account_registration", "aws", defaultAllowEmptyValues))
		}
	}

	for rp := auditmanager.NewListAssessmentReportsPaginator(svc, &auditmanager.ListAssessmentReportsInput{}); rp.HasMorePages(); {
		page, err := rp.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, r := range page.AssessmentReports {
			id := StringValue(r.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(r.Name), "aws_auditmanager_assessment_report", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
