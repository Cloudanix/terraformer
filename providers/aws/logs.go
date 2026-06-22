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
	"strconv"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cloudwatchlogstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

var logsAllowEmptyValues = []string{"tags."}

type LogsGenerator struct {
	AWSService
}

func (g *LogsGenerator) createResources(logGroups *cloudwatchlogs.DescribeLogGroupsOutput) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	for _, logGroup := range logGroups.LogGroups {
		resourceName := StringValue(logGroup.LogGroupName)

		attributes := map[string]string{}

		if logGroup.RetentionInDays != nil {
			attributes["retention_in_days"] = strconv.FormatInt(int64(*logGroup.RetentionInDays), 10)
		}

		if logGroup.KmsKeyId != nil {
			attributes["kms_key_id"] = *logGroup.KmsKeyId
		}

		resources = append(resources, terraformutils.NewResource(
			resourceName,
			resourceName,
			"aws_cloudwatch_log_group",
			"aws",
			attributes,
			logsAllowEmptyValues,
			map[string]interface{}{}))
	}
	return resources
}

// Generate TerraformResources from AWS API
func (g *LogsGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := cloudwatchlogs.NewFromConfig(config)

	ctx := context.TODO()
	var logGroupNames []string
	p := cloudwatchlogs.NewDescribeLogGroupsPaginator(svc, &cloudwatchlogs.DescribeLogGroupsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, lg := range page.LogGroups {
			logGroupNames = append(logGroupNames, StringValue(lg.LogGroupName))
		}
		g.Resources = append(g.Resources, g.createResources(page)...)
	}

	mf := cloudwatchlogs.NewDescribeMetricFiltersPaginator(svc, &cloudwatchlogs.DescribeMetricFiltersInput{})
	for mf.HasMorePages() {
		page, err := mf.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, f := range page.MetricFilters {
			group := StringValue(f.LogGroupName)
			name := StringValue(f.FilterName)
			if group == "" || name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				group+":"+name, group+"_"+name, "aws_cloudwatch_log_metric_filter", "aws", logsAllowEmptyValues))
		}
	}

	for _, group := range logGroupNames {
		if group == "" {
			continue
		}
		out, err := svc.DescribeSubscriptionFilters(ctx, &cloudwatchlogs.DescribeSubscriptionFiltersInput{LogGroupName: &group})
		if err != nil {
			continue
		}
		for _, f := range out.SubscriptionFilters {
			name := StringValue(f.FilterName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				group+"|"+name, group+"_"+name, "aws_cloudwatch_log_subscription_filter", "aws", logsAllowEmptyValues))
		}
		if dp, err := svc.GetDataProtectionPolicy(ctx, &cloudwatchlogs.GetDataProtectionPolicyInput{LogGroupIdentifier: &group}); err == nil &&
			dp.PolicyDocument != nil && StringValue(dp.PolicyDocument) != "" {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				group, group, "aws_cloudwatch_log_data_protection_policy", "aws", logsAllowEmptyValues))
		}
	}

	dests := cloudwatchlogs.NewDescribeDestinationsPaginator(svc, &cloudwatchlogs.DescribeDestinationsInput{})
	for dests.HasMorePages() {
		page, err := dests.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, d := range page.Destinations {
			name := StringValue(d.DestinationName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_cloudwatch_log_destination", "aws", logsAllowEmptyValues))
			if StringValue(d.AccessPolicy) != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					name, name, "aws_cloudwatch_log_destination_policy", "aws", logsAllowEmptyValues))
			}
		}
	}

	if rp, err := svc.DescribeResourcePolicies(ctx, &cloudwatchlogs.DescribeResourcePoliciesInput{}); err == nil {
		for _, p := range rp.ResourcePolicies {
			name := StringValue(p.PolicyName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_cloudwatch_log_resource_policy", "aws", logsAllowEmptyValues))
		}
	}

	for _, pt := range cloudwatchlogstypes.PolicyType("").Values() {
		ap, err := svc.DescribeAccountPolicies(ctx, &cloudwatchlogs.DescribeAccountPoliciesInput{PolicyType: pt})
		if err != nil {
			continue
		}
		for _, p := range ap.AccountPolicies {
			name := StringValue(p.PolicyName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name+":"+string(p.PolicyType), name, "aws_cloudwatch_log_account_policy", "aws", logsAllowEmptyValues))
		}
	}

	if qd, err := svc.DescribeQueryDefinitions(ctx, &cloudwatchlogs.DescribeQueryDefinitionsInput{}); err == nil {
		for _, q := range qd.QueryDefinitions {
			id := StringValue(q.QueryDefinitionId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(q.Name), "aws_cloudwatch_query_definition", "aws", logsAllowEmptyValues))
		}
	}
	return nil
}

// remove retention_in_days if it is 0 (it gets added by the "refresh" stage)
func (g *LogsGenerator) PostConvertHook() error {
	for _, resource := range g.Resources {
		if resource.Item["retention_in_days"] == "0" {
			delete(resource.Item, "retention_in_days")
		}
	}
	return nil
}
