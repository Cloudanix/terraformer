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
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/aws/aws-sdk-go-v2/service/backup/types"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type BackupGenerator struct {
	AWSService
}

// InitResources enumerates AWS Backup resources. Import IDs (per the
// terraform-provider-aws docs):
//   - aws_backup_vault          → vault name
//   - aws_backup_plan           → plan ID
//   - aws_backup_selection      → "<plan-id>|<selection-id>"
//   - aws_backup_framework      → framework name
//   - aws_backup_report_plan    → report plan name
//   - aws_backup_restore_testing_plan → restore testing plan name
func (g *BackupGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := backup.NewFromConfig(config)
	ctx := context.TODO()

	var vaultNames []string
	vaults := backup.NewListBackupVaultsPaginator(svc, &backup.ListBackupVaultsInput{})
	for vaults.HasMorePages() {
		page, err := vaults.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, v := range page.BackupVaultList {
			name := StringValue(v.BackupVaultName)
			vaultNames = append(vaultNames, name)
			tfType := "aws_backup_vault"
			if v.VaultType == types.VaultTypeLogicallyAirGappedBackupVault {
				tfType = "aws_backup_logically_air_gapped_vault"
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, tfType, "aws", defaultAllowEmptyValues))
			if v.Locked != nil && *v.Locked {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					name, name, "aws_backup_vault_lock_configuration", "aws", defaultAllowEmptyValues))
			}
		}
	}

	for _, vault := range vaultNames {
		if vault == "" {
			continue
		}
		if _, err := svc.GetBackupVaultNotifications(ctx, &backup.GetBackupVaultNotificationsInput{BackupVaultName: aws.String(vault)}); err == nil {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				vault, vault, "aws_backup_vault_notifications", "aws", defaultAllowEmptyValues))
		}
		if _, err := svc.GetBackupVaultAccessPolicy(ctx, &backup.GetBackupVaultAccessPolicyInput{BackupVaultName: aws.String(vault)}); err == nil {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				vault, vault, "aws_backup_vault_policy", "aws", defaultAllowEmptyValues))
		}
	}

	// Plans, collecting IDs so selections (nested under each plan) can be listed.
	var planIDs []string
	plans := backup.NewListBackupPlansPaginator(svc, &backup.ListBackupPlansInput{})
	for plans.HasMorePages() {
		page, err := plans.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, plan := range page.BackupPlansList {
			planID := StringValue(plan.BackupPlanId)
			if planID == "" {
				continue
			}
			planIDs = append(planIDs, planID)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				planID, StringValue(plan.BackupPlanName), "aws_backup_plan", "aws", defaultAllowEmptyValues))
		}
	}

	for _, planID := range planIDs {
		selections := backup.NewListBackupSelectionsPaginator(svc, &backup.ListBackupSelectionsInput{BackupPlanId: aws.String(planID)})
		for selections.HasMorePages() {
			page, err := selections.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, sel := range page.BackupSelectionsList {
				selID := StringValue(sel.SelectionId)
				if selID == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					fmt.Sprintf("%s|%s", planID, selID),
					StringValue(sel.SelectionName),
					"aws_backup_selection", "aws", defaultAllowEmptyValues))
			}
		}
	}

	frameworks := backup.NewListFrameworksPaginator(svc, &backup.ListFrameworksInput{})
	for frameworks.HasMorePages() {
		page, err := frameworks.NextPage(ctx)
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.Frameworks, "aws_backup_framework",
			defaultAllowEmptyValues,
			func(f types.Framework) string { return StringValue(f.FrameworkName) },
			func(f types.Framework) string { return StringValue(f.FrameworkName) })
	}

	reportPlans := backup.NewListReportPlansPaginator(svc, &backup.ListReportPlansInput{})
	for reportPlans.HasMorePages() {
		page, err := reportPlans.NextPage(ctx)
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.ReportPlans, "aws_backup_report_plan",
			defaultAllowEmptyValues,
			func(r types.ReportPlan) string { return StringValue(r.ReportPlanName) },
			func(r types.ReportPlan) string { return StringValue(r.ReportPlanName) })
	}

	var restoreTestingPlanNames []string
	restoreTestingPlans := backup.NewListRestoreTestingPlansPaginator(svc, &backup.ListRestoreTestingPlansInput{})
	for restoreTestingPlans.HasMorePages() {
		page, err := restoreTestingPlans.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, p := range page.RestoreTestingPlans {
			restoreTestingPlanNames = append(restoreTestingPlanNames, StringValue(p.RestoreTestingPlanName))
		}
		g.Resources = appendSimpleResources(g.Resources, page.RestoreTestingPlans, "aws_backup_restore_testing_plan",
			defaultAllowEmptyValues,
			func(p types.RestoreTestingPlanForList) string { return StringValue(p.RestoreTestingPlanName) },
			func(p types.RestoreTestingPlanForList) string { return StringValue(p.RestoreTestingPlanName) })
	}

	for _, planName := range restoreTestingPlanNames {
		if planName == "" {
			continue
		}
		selections := backup.NewListRestoreTestingSelectionsPaginator(svc, &backup.ListRestoreTestingSelectionsInput{RestoreTestingPlanName: aws.String(planName)})
		for selections.HasMorePages() {
			page, err := selections.NextPage(ctx)
			if err != nil {
				break
			}
			for _, sel := range page.RestoreTestingSelections {
				selName := StringValue(sel.RestoreTestingSelectionName)
				if selName == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					fmt.Sprintf("%s:%s", selName, planName),
					fmt.Sprintf("%s_%s", planName, selName),
					"aws_backup_restore_testing_selection", "aws", defaultAllowEmptyValues))
			}
		}
	}

	// Account-level singletons keyed by account ID.
	account, err := g.getAccountNumber(config)
	if err != nil {
		return err
	}
	accountID := StringValue(account)
	if accountID != "" {
		if _, err := svc.DescribeGlobalSettings(ctx, &backup.DescribeGlobalSettingsInput{}); err == nil {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				accountID, accountID, "aws_backup_global_settings", "aws", defaultAllowEmptyValues))
		}
		if _, err := svc.DescribeRegionSettings(ctx, &backup.DescribeRegionSettingsInput{}); err == nil {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				accountID, accountID, "aws_backup_region_settings", "aws", defaultAllowEmptyValues))
		}
	}

	return nil
}
