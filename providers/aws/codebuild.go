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
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
)

var codebuildAllowEmptyValues = []string{"tags."}

type CodeBuildGenerator struct {
	AWSService
}

func (g *CodeBuildGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := codebuild.NewFromConfig(config)
	ctx := context.TODO()
	var projectNames []string
	p := codebuild.NewListProjectsPaginator(svc, &codebuild.ListProjectsInput{})
	for p.HasMorePages() {
		page, e := p.NextPage(ctx)
		if e != nil {
			return e
		}
		for _, project := range page.Projects {
			projectNames = append(projectNames, project)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				project,
				project,
				"aws_codebuild_project",
				"aws",
				codebuildAllowEmptyValues))
		}
	}

	// Webhooks are embedded in each project; batch-fetch in chunks of 100.
	for i := 0; i < len(projectNames); i += 100 {
		end := i + 100
		if end > len(projectNames) {
			end = len(projectNames)
		}
		out, err := svc.BatchGetProjects(ctx, &codebuild.BatchGetProjectsInput{Names: projectNames[i:end]})
		if err != nil {
			continue
		}
		for _, proj := range out.Projects {
			if arn := StringValue(proj.Arn); arn != "" {
				arnCopy := arn
				if rp, err := svc.GetResourcePolicy(ctx, &codebuild.GetResourcePolicyInput{ResourceArn: &arnCopy}); err == nil && StringValue(rp.Policy) != "" {
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						arn, arn, "aws_codebuild_resource_policy", "aws", codebuildAllowEmptyValues))
				}
			}
			if proj.Webhook == nil {
				continue
			}
			name := StringValue(proj.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_codebuild_webhook", "aws", codebuildAllowEmptyValues))
		}
	}

	if creds, err := svc.ListSourceCredentials(ctx, &codebuild.ListSourceCredentialsInput{}); err == nil {
		for _, c := range creds.SourceCredentialsInfos {
			arn := StringValue(c.Arn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_codebuild_source_credential", "aws", codebuildAllowEmptyValues))
		}
	}

	rg := codebuild.NewListReportGroupsPaginator(svc, &codebuild.ListReportGroupsInput{})
	for rg.HasMorePages() {
		page, err := rg.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, arn := range page.ReportGroups {
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_codebuild_report_group", "aws", codebuildAllowEmptyValues))
		}
	}

	for fp := codebuild.NewListFleetsPaginator(svc, &codebuild.ListFleetsInput{}); fp.HasMorePages(); {
		page, err := fp.NextPage(ctx)
		if err != nil {
			break
		}
		for _, arn := range page.Fleets {
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_codebuild_fleet", "aws", codebuildAllowEmptyValues))
		}
	}
	return nil
}

func (g *CodeBuildGenerator) PostConvertHook() error {
	for _, r := range g.Resources {
		if r.InstanceInfo.Type != "aws_codebuild_project" {
			continue
		}
		if r.InstanceState.Attributes["concurrent_build_limit"] == "0" {
			delete(r.Item, "concurrent_build_limit")
		}
	}
	return nil
}
