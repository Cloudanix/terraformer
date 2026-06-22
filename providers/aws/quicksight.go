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
	"github.com/aws/aws-sdk-go-v2/service/quicksight"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type QuickSightGenerator struct {
	AWSService
}

// InitResources enumerates QuickSight data sets and data sources for the current
// account. All QuickSight APIs are scoped by AwsAccountId. Import IDs are
// "<account-id>,<resource-id>".
func (g *QuickSightGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	account, err := g.getAccountNumber(config)
	if err != nil {
		return err
	}
	accountID := StringValue(account)
	svc := quicksight.NewFromConfig(config)
	ctx := context.TODO()

	dataSets := quicksight.NewListDataSetsPaginator(svc, &quicksight.ListDataSetsInput{AwsAccountId: aws.String(accountID)})
	for dataSets.HasMorePages() {
		page, err := dataSets.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, ds := range page.DataSetSummaries {
			id := StringValue(ds.DataSetId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				accountID+","+id, id, "aws_quicksight_data_set", "aws", defaultAllowEmptyValues))
			if sched, err := svc.ListRefreshSchedules(ctx, &quicksight.ListRefreshSchedulesInput{AwsAccountId: aws.String(accountID), DataSetId: aws.String(id)}); err == nil {
				for _, s := range sched.RefreshSchedules {
					sid := StringValue(s.ScheduleId)
					if sid == "" {
						continue
					}
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						accountID+","+id+","+sid, id+"_"+sid, "aws_quicksight_refresh_schedule", "aws", defaultAllowEmptyValues))
				}
			}
		}
	}

	dataSources := quicksight.NewListDataSourcesPaginator(svc, &quicksight.ListDataSourcesInput{AwsAccountId: aws.String(accountID)})
	for dataSources.HasMorePages() {
		page, err := dataSources.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, src := range page.DataSources {
			id := StringValue(src.DataSourceId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				accountID+","+id, id, "aws_quicksight_data_source", "aws", defaultAllowEmptyValues))
		}
	}

	add := func(id, tfType string) {
		if id != "" {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				accountID+","+id, id, tfType, "aws", defaultAllowEmptyValues))
		}
	}
	for p := quicksight.NewListAnalysesPaginator(svc, &quicksight.ListAnalysesInput{AwsAccountId: aws.String(accountID)}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range page.AnalysisSummaryList {
			add(StringValue(x.AnalysisId), "aws_quicksight_analysis")
		}
	}
	for p := quicksight.NewListDashboardsPaginator(svc, &quicksight.ListDashboardsInput{AwsAccountId: aws.String(accountID)}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range page.DashboardSummaryList {
			add(StringValue(x.DashboardId), "aws_quicksight_dashboard")
		}
	}
	for p := quicksight.NewListTemplatesPaginator(svc, &quicksight.ListTemplatesInput{AwsAccountId: aws.String(accountID)}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range page.TemplateSummaryList {
			templateID := StringValue(x.TemplateId)
			add(templateID, "aws_quicksight_template")
			if templateID == "" {
				continue
			}
			for ap := quicksight.NewListTemplateAliasesPaginator(svc, &quicksight.ListTemplateAliasesInput{AwsAccountId: aws.String(accountID), TemplateId: aws.String(templateID)}); ap.HasMorePages(); {
				apage, err := ap.NextPage(ctx)
				if err != nil {
					break
				}
				for _, a := range apage.TemplateAliasList {
					alias := StringValue(a.AliasName)
					if alias == "" {
						continue
					}
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						accountID+","+templateID+","+alias, templateID+"_"+alias, "aws_quicksight_template_alias", "aws", defaultAllowEmptyValues))
				}
			}
		}
	}
	for p := quicksight.NewListThemesPaginator(svc, &quicksight.ListThemesInput{AwsAccountId: aws.String(accountID)}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range page.ThemeSummaryList {
			add(StringValue(x.ThemeId), "aws_quicksight_theme")
		}
	}
	for p := quicksight.NewListFoldersPaginator(svc, &quicksight.ListFoldersInput{AwsAccountId: aws.String(accountID)}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range page.FolderSummaryList {
			add(StringValue(x.FolderId), "aws_quicksight_folder")
		}
	}
	for p := quicksight.NewListVPCConnectionsPaginator(svc, &quicksight.ListVPCConnectionsInput{AwsAccountId: aws.String(accountID)}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range page.VPCConnectionSummaries {
			add(StringValue(x.VPCConnectionId), "aws_quicksight_vpc_connection")
		}
	}

	var namespaces []string
	for p := quicksight.NewListNamespacesPaginator(svc, &quicksight.ListNamespacesInput{AwsAccountId: aws.String(accountID)}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range page.Namespaces {
			name := StringValue(x.Name)
			if name == "" {
				continue
			}
			namespaces = append(namespaces, name)
			add(name, "aws_quicksight_namespace")
		}
	}
	for _, ns := range namespaces {
		namespace := ns
		for gp := quicksight.NewListGroupsPaginator(svc, &quicksight.ListGroupsInput{AwsAccountId: aws.String(accountID), Namespace: &namespace}); gp.HasMorePages(); {
			page, err := gp.NextPage(ctx)
			if err != nil {
				break
			}
			for _, grp := range page.GroupList {
				name := StringValue(grp.GroupName)
				if name == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					accountID+"/"+namespace+"/"+name, name, "aws_quicksight_group", "aws", defaultAllowEmptyValues))
			}
		}
		for up := quicksight.NewListUsersPaginator(svc, &quicksight.ListUsersInput{AwsAccountId: aws.String(accountID), Namespace: &namespace}); up.HasMorePages(); {
			page, err := up.NextPage(ctx)
			if err != nil {
				break
			}
			for _, u := range page.UserList {
				name := StringValue(u.UserName)
				if name == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					accountID+"/"+namespace+"/"+name, name, "aws_quicksight_user", "aws", defaultAllowEmptyValues))
			}
		}
		for ap := quicksight.NewListIAMPolicyAssignmentsPaginator(svc, &quicksight.ListIAMPolicyAssignmentsInput{AwsAccountId: aws.String(accountID), Namespace: &namespace}); ap.HasMorePages(); {
			page, err := ap.NextPage(ctx)
			if err != nil {
				break
			}
			for _, a := range page.IAMPolicyAssignments {
				name := StringValue(a.AssignmentName)
				if name == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					accountID+","+namespace+","+name, namespace+"_"+name, "aws_quicksight_iam_policy_assignment", "aws", defaultAllowEmptyValues))
			}
		}
	}
	return nil
}
