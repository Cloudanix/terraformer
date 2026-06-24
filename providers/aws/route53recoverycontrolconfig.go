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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53recoverycontrolconfig"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type Route53RecoveryControlConfigGenerator struct {
	AWSService
}

// InitResources enumerates Route 53 Application Recovery Controller clusters.
// Import ID is the cluster ARN.
func (g *Route53RecoveryControlConfigGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := route53recoverycontrolconfig.NewFromConfig(config)
	var clusterArns []string
	p := route53recoverycontrolconfig.NewListClustersPaginator(svc, &route53recoverycontrolconfig.ListClustersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, c := range page.Clusters {
			arn := StringValue(c.ClusterArn)
			if arn == "" {
				continue
			}
			clusterArns = append(clusterArns, arn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(c.Name), "aws_route53recoverycontrolconfig_cluster", "aws", defaultAllowEmptyValues))
		}
	}

	ctx := awsContext()
	for _, clusterArn := range clusterArns {
		ca := clusterArn
		for cp := route53recoverycontrolconfig.NewListControlPanelsPaginator(svc, &route53recoverycontrolconfig.ListControlPanelsInput{ClusterArn: &ca}); cp.HasMorePages(); {
			page, err := cp.NextPage(ctx)
			if err != nil {
				break
			}
			for _, panel := range page.ControlPanels {
				panelArn := StringValue(panel.ControlPanelArn)
				if panelArn == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					panelArn, StringValue(panel.Name), "aws_route53recoverycontrolconfig_control_panel", "aws", defaultAllowEmptyValues))
				for rp := route53recoverycontrolconfig.NewListRoutingControlsPaginator(svc, &route53recoverycontrolconfig.ListRoutingControlsInput{ControlPanelArn: aws.String(panelArn)}); rp.HasMorePages(); {
					rpage, err := rp.NextPage(ctx)
					if err != nil {
						break
					}
					for _, rc := range rpage.RoutingControls {
						a := StringValue(rc.RoutingControlArn)
						if a != "" {
							g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
								a, StringValue(rc.Name), "aws_route53recoverycontrolconfig_routing_control", "aws", defaultAllowEmptyValues))
						}
					}
				}
				for sp := route53recoverycontrolconfig.NewListSafetyRulesPaginator(svc, &route53recoverycontrolconfig.ListSafetyRulesInput{ControlPanelArn: aws.String(panelArn)}); sp.HasMorePages(); {
					spage, err := sp.NextPage(ctx)
					if err != nil {
						break
					}
					for _, rule := range spage.SafetyRules {
						var a string
						switch {
						case rule.ASSERTION != nil:
							a = StringValue(rule.ASSERTION.SafetyRuleArn)
						case rule.GATING != nil:
							a = StringValue(rule.GATING.SafetyRuleArn)
						}
						if a != "" {
							g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
								a, a, "aws_route53recoverycontrolconfig_safety_rule", "aws", defaultAllowEmptyValues))
						}
					}
				}
			}
		}
	}
	return nil
}
