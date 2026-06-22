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

	"github.com/aws/aws-sdk-go-v2/service/route53recoveryreadiness"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type Route53RecoveryReadinessGenerator struct {
	AWSService
}

// InitResources enumerates Route 53 Recovery Readiness cells. Import ID is the
// cell name.
func (g *Route53RecoveryReadinessGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := route53recoveryreadiness.NewFromConfig(config)
	p := route53recoveryreadiness.NewListCellsPaginator(svc, &route53recoveryreadiness.ListCellsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, c := range page.Cells {
			name := StringValue(c.CellName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_route53recoveryreadiness_cell", "aws", defaultAllowEmptyValues))
		}
	}

	ctx := context.TODO()
	for rc := route53recoveryreadiness.NewListReadinessChecksPaginator(svc, &route53recoveryreadiness.ListReadinessChecksInput{}); rc.HasMorePages(); {
		page, err := rc.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range page.ReadinessChecks {
			name := StringValue(x.ReadinessCheckName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_route53recoveryreadiness_readiness_check", "aws", defaultAllowEmptyValues))
		}
	}
	for rg := route53recoveryreadiness.NewListRecoveryGroupsPaginator(svc, &route53recoveryreadiness.ListRecoveryGroupsInput{}); rg.HasMorePages(); {
		page, err := rg.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range page.RecoveryGroups {
			name := StringValue(x.RecoveryGroupName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_route53recoveryreadiness_recovery_group", "aws", defaultAllowEmptyValues))
		}
	}
	for rs := route53recoveryreadiness.NewListResourceSetsPaginator(svc, &route53recoveryreadiness.ListResourceSetsInput{}); rs.HasMorePages(); {
		page, err := rs.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range page.ResourceSets {
			name := StringValue(x.ResourceSetName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_route53recoveryreadiness_resource_set", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
