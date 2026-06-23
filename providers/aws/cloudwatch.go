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
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchevents"
)

var cloudwatchAllowEmptyValues = []string{"tags."}

type CloudWatchGenerator struct {
	AWSService
}

func (g *CloudWatchGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}

	cloudwatchSvc := cloudwatch.NewFromConfig(config)
	err := g.createMetricAlarms(cloudwatchSvc)
	if err != nil {
		return err
	}
	err = g.createDashboards(cloudwatchSvc)
	if err != nil {
		return err
	}
	if err := g.createMetricStreams(cloudwatchSvc); err != nil {
		return err
	}
	if err := g.createInsightRules(cloudwatchSvc); err != nil {
		return err
	}

	cloudwatcheventsSvc := cloudwatchevents.NewFromConfig(config)
	err = g.createRules(cloudwatcheventsSvc)
	if err != nil {
		return err
	}
	if err := g.createEventBuses(cloudwatcheventsSvc); err != nil {
		return err
	}
	if err := g.createEventConnections(cloudwatcheventsSvc); err != nil {
		return err
	}
	if err := g.createEventAPIDestinations(cloudwatcheventsSvc); err != nil {
		return err
	}
	if archives, err := cloudwatcheventsSvc.ListArchives(awsContext(), &cloudwatchevents.ListArchivesInput{}); err == nil {
		for _, a := range archives.Archives {
			name := StringValue(a.ArchiveName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_cloudwatch_event_archive", "aws", cloudwatchAllowEmptyValues))
		}
	}

	return nil
}

// createInsightRules emits CloudWatch Contributor Insights rules (imported by name).
func (g *CloudWatchGenerator) createInsightRules(svc *cloudwatch.Client) error {
	p := cloudwatch.NewDescribeInsightRulesPaginator(svc, &cloudwatch.DescribeInsightRulesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, r := range page.InsightRules {
			name := StringValue(r.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_cloudwatch_contributor_insight_rule", "aws", cloudwatchAllowEmptyValues))
		}
	}
	return nil
}

func (g *CloudWatchGenerator) createMetricStreams(svc *cloudwatch.Client) error {
	p := cloudwatch.NewListMetricStreamsPaginator(svc, &cloudwatch.ListMetricStreamsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, ms := range page.Entries {
			name := StringValue(ms.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_cloudwatch_metric_stream", "aws", cloudwatchAllowEmptyValues))
		}
	}
	return nil
}

func (g *CloudWatchGenerator) createEventBuses(svc *cloudwatchevents.Client) error {
	var nextToken *string
	for {
		output, err := svc.ListEventBuses(awsContext(), &cloudwatchevents.ListEventBusesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, bus := range output.EventBuses {
			name := StringValue(bus.Name)
			// The "default" event bus is implicit and not managed as a resource.
			if name == "" || name == "default" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_cloudwatch_event_bus", "aws", cloudwatchAllowEmptyValues))
			// Resource-based policy is a separate resource (not inlined on the bus),
			// imported by bus name. Emit only when a policy is attached.
			if StringValue(bus.Policy) != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					name, name, "aws_cloudwatch_event_bus_policy", "aws", cloudwatchAllowEmptyValues))
			}
		}
		nextToken = output.NextToken
		if nextToken == nil {
			return nil
		}
	}
}

func (g *CloudWatchGenerator) createEventConnections(svc *cloudwatchevents.Client) error {
	var nextToken *string
	for {
		output, err := svc.ListConnections(awsContext(), &cloudwatchevents.ListConnectionsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, conn := range output.Connections {
			name := StringValue(conn.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_cloudwatch_event_connection", "aws", cloudwatchAllowEmptyValues))
		}
		nextToken = output.NextToken
		if nextToken == nil {
			return nil
		}
	}
}

func (g *CloudWatchGenerator) createEventAPIDestinations(svc *cloudwatchevents.Client) error {
	var nextToken *string
	for {
		output, err := svc.ListApiDestinations(awsContext(), &cloudwatchevents.ListApiDestinationsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, dest := range output.ApiDestinations {
			name := StringValue(dest.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_cloudwatch_event_api_destination", "aws", cloudwatchAllowEmptyValues))
		}
		nextToken = output.NextToken
		if nextToken == nil {
			return nil
		}
	}
}

func (g *CloudWatchGenerator) createMetricAlarms(cloudwatchSvc *cloudwatch.Client) error {
	var nextToken *string
	for {
		output, err := cloudwatchSvc.DescribeAlarms(awsContext(), &cloudwatch.DescribeAlarmsInput{
			NextToken: nextToken,
		})
		if err != nil {
			return err
		}
		for _, metricAlarm := range output.MetricAlarms {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*metricAlarm.AlarmName,
				*metricAlarm.AlarmName,
				"aws_cloudwatch_metric_alarm",
				"aws",
				cloudwatchAllowEmptyValues))
		}
		for _, compositeAlarm := range output.CompositeAlarms {
			name := StringValue(compositeAlarm.AlarmName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_cloudwatch_composite_alarm", "aws", cloudwatchAllowEmptyValues))
		}
		nextToken = output.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *CloudWatchGenerator) createDashboards(cloudwatchSvc *cloudwatch.Client) error {
	var nextToken *string
	for {
		output, err := cloudwatchSvc.ListDashboards(awsContext(), &cloudwatch.ListDashboardsInput{
			NextToken: nextToken,
		})
		if err != nil {
			return err
		}
		for _, dashboardEntry := range output.DashboardEntries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*dashboardEntry.DashboardName,
				*dashboardEntry.DashboardName,
				"aws_cloudwatch_dashboard",
				"aws",
				cloudwatchAllowEmptyValues))
		}
		nextToken = output.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *CloudWatchGenerator) createRules(cloudwatcheventsSvc *cloudwatchevents.Client) error {
	var listRulesNextToken *string
	for {
		output, err := cloudwatcheventsSvc.ListRules(awsContext(), &cloudwatchevents.ListRulesInput{
			NextToken: listRulesNextToken,
		})
		if err != nil {
			return err
		}
		for _, rule := range output.Rules {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*rule.Name,
				*rule.Name,
				"aws_cloudwatch_event_rule",
				"aws",
				cloudwatchAllowEmptyValues))

			var listTargetsNextToken *string
			for {
				targetResponse, err := cloudwatcheventsSvc.ListTargetsByRule(awsContext(), &cloudwatchevents.ListTargetsByRuleInput{
					Rule:      rule.Name,
					NextToken: listTargetsNextToken,
				})
				if err != nil {
					return err
				}
				for _, target := range targetResponse.Targets {
					targetRef := *rule.Name + "/" + *target.Id
					g.Resources = append(g.Resources, terraformutils.NewResource(
						targetRef,
						targetRef,
						"aws_cloudwatch_event_target",
						"aws",
						map[string]string{
							"rule":      *rule.Name,
							"target_id": *target.Id,
						},
						cloudwatchAllowEmptyValues,
						map[string]interface{}{}))
				}
				listTargetsNextToken = output.NextToken
				if listTargetsNextToken == nil {
					break
				}
			}
		}
		listRulesNextToken = output.NextToken
		if listRulesNextToken == nil {
			break
		}
	}

	return nil
}
