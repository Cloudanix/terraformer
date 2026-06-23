// Copyright 2021 The Terraformer Authors.
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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

var ssmAllowEmptyValues = []string{"tags."}

type SsmGenerator struct {
	AWSService
}

func (g *SsmGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := ssm.NewFromConfig(config)
	p := ssm.NewDescribeParametersPaginator(svc, &ssm.DescribeParametersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, parameter := range page.Parameters {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(parameter.Name),
				StringValue(parameter.Name),
				"aws_ssm_parameter",
				"aws",
				ssmAllowEmptyValues,
			))
		}
	}

	if err := g.addDocuments(svc); err != nil {
		return err
	}
	if err := g.addMaintenanceWindows(svc); err != nil {
		return err
	}
	if err := g.addPatchBaselines(svc); err != nil {
		return err
	}
	if err := g.addAssociations(svc); err != nil {
		return err
	}
	if err := g.addActivationsAndSyncs(svc); err != nil {
		return err
	}

	return nil
}

func (g *SsmGenerator) addActivationsAndSyncs(svc *ssm.Client) error {
	ctx := awsContext()
	for p := ssm.NewDescribeActivationsPaginator(svc, &ssm.DescribeActivationsInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, a := range page.ActivationList {
			id := StringValue(a.ActivationId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_ssm_activation", "aws", ssmAllowEmptyValues))
		}
	}
	for pg := ssm.NewDescribePatchGroupsPaginator(svc, &ssm.DescribePatchGroupsInput{}); pg.HasMorePages(); {
		page, err := pg.NextPage(ctx)
		if err != nil {
			break
		}
		for _, m := range page.Mappings {
			group := StringValue(m.PatchGroup)
			if group == "" || m.BaselineIdentity == nil {
				continue
			}
			baseline := StringValue(m.BaselineIdentity.BaselineId)
			if baseline == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				group+","+baseline, group+"_"+baseline, "aws_ssm_patch_group", "aws", ssmAllowEmptyValues))
		}
	}

	var syncToken *string
	for {
		out, err := svc.ListResourceDataSync(ctx, &ssm.ListResourceDataSyncInput{NextToken: syncToken})
		if err != nil {
			break
		}
		for _, s := range out.ResourceDataSyncItems {
			name := StringValue(s.SyncName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_ssm_resource_data_sync", "aws", ssmAllowEmptyValues))
		}
		if out.NextToken == nil {
			break
		}
		syncToken = out.NextToken
	}
	return nil
}

func (g *SsmGenerator) addDocuments(svc *ssm.Client) error {
	// Self-owned documents only — the account owns thousands of AWS-managed ones.
	p := ssm.NewListDocumentsPaginator(svc, &ssm.ListDocumentsInput{
		Filters: []types.DocumentKeyValuesFilter{{Key: aws.String("Owner"), Values: []string{"Self"}}},
	})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, doc := range page.DocumentIdentifiers {
			name := StringValue(doc.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_ssm_document", "aws", ssmAllowEmptyValues))
		}
	}
	return nil
}

func (g *SsmGenerator) addMaintenanceWindows(svc *ssm.Client) error {
	p := ssm.NewDescribeMaintenanceWindowsPaginator(svc, &ssm.DescribeMaintenanceWindowsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, window := range page.WindowIdentities {
			id := StringValue(window.WindowId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(window.Name), "aws_ssm_maintenance_window", "aws", ssmAllowEmptyValues))
			g.addMaintenanceWindowChildren(svc, id)
		}
	}
	return nil
}

func (g *SsmGenerator) addMaintenanceWindowChildren(svc *ssm.Client, windowID string) {
	ctx := awsContext()
	wid := windowID
	for p := ssm.NewDescribeMaintenanceWindowTargetsPaginator(svc, &ssm.DescribeMaintenanceWindowTargetsInput{WindowId: &wid}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, t := range page.Targets {
			tid := StringValue(t.WindowTargetId)
			if tid == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				wid+"/"+tid, wid+"_"+tid, "aws_ssm_maintenance_window_target", "aws", ssmAllowEmptyValues))
		}
	}
	for p := ssm.NewDescribeMaintenanceWindowTasksPaginator(svc, &ssm.DescribeMaintenanceWindowTasksInput{WindowId: &wid}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, t := range page.Tasks {
			tid := StringValue(t.WindowTaskId)
			if tid == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				wid+"/"+tid, wid+"_"+tid, "aws_ssm_maintenance_window_task", "aws", ssmAllowEmptyValues))
		}
	}
}

func (g *SsmGenerator) addPatchBaselines(svc *ssm.Client) error {
	// Self-owned baselines only — skip the AWS-managed default baselines.
	selfBaselines := map[string]bool{}
	p := ssm.NewDescribePatchBaselinesPaginator(svc, &ssm.DescribePatchBaselinesInput{
		Filters: []types.PatchOrchestratorFilter{{Key: aws.String("OWNER"), Values: []string{"Self"}}},
	})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, baseline := range page.BaselineIdentities {
			id := StringValue(baseline.BaselineId)
			if id == "" {
				continue
			}
			selfBaselines[id] = true
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(baseline.BaselineName), "aws_ssm_patch_baseline", "aws", ssmAllowEmptyValues))
		}
	}
	// A custom default baseline (set per OS) maps to aws_ssm_default_patch_baseline.
	for _, os := range types.OperatingSystem("").Values() {
		out, err := svc.GetDefaultPatchBaseline(awsContext(), &ssm.GetDefaultPatchBaselineInput{OperatingSystem: os})
		if err != nil {
			continue
		}
		id := StringValue(out.BaselineId)
		if selfBaselines[id] {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, string(os), "aws_ssm_default_patch_baseline", "aws", ssmAllowEmptyValues))
		}
	}
	return nil
}

func (g *SsmGenerator) addAssociations(svc *ssm.Client) error {
	p := ssm.NewListAssociationsPaginator(svc, &ssm.ListAssociationsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, association := range page.Associations {
			id := StringValue(association.AssociationId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_ssm_association", "aws", ssmAllowEmptyValues))
		}
	}
	return nil
}
