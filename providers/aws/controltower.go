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
	"github.com/aws/aws-sdk-go-v2/service/controltower"
	"github.com/aws/aws-sdk-go-v2/service/organizations"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type ControlTowerGenerator struct {
	AWSService
}

// InitResources enumerates Control Tower landing zones. Import ID is the ARN.
func (g *ControlTowerGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := controltower.NewFromConfig(config)

	p := controltower.NewListLandingZonesPaginator(svc, &controltower.ListLandingZonesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, lz := range page.LandingZones {
			arn := StringValue(lz.Arn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_controltower_landing_zone", "aws", defaultAllowEmptyValues))
		}
	}

	g.loadControls(svc, organizations.NewFromConfig(config))
	return nil
}

// loadControls emits enabled Control Tower controls per organizational unit.
// ListEnabledControls is OU-scoped, so walk the org tree via Organizations.
// Import id is "<target_identifier>,<control_identifier>".
func (g *ControlTowerGenerator) loadControls(svc *controltower.Client, orgs *organizations.Client) {
	for _, ouArn := range g.organizationalUnitARNs(orgs) {
		for p := controltower.NewListEnabledControlsPaginator(svc, &controltower.ListEnabledControlsInput{TargetIdentifier: &ouArn}); p.HasMorePages(); {
			page, err := p.NextPage(awsContext())
			if err != nil {
				break
			}
			for _, c := range page.EnabledControls {
				target, control := StringValue(c.TargetIdentifier), StringValue(c.ControlIdentifier)
				if target == "" || control == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					target+","+control, control, "aws_controltower_control", "aws", defaultAllowEmptyValues))
			}
		}
	}
}

// organizationalUnitARNs returns every OU ARN by walking the org tree from each
// root. Returns nil if Organizations is not in use (controls then can't exist).
func (g *ControlTowerGenerator) organizationalUnitARNs(orgs *organizations.Client) []string {
	roots, err := orgs.ListRoots(awsContext(), &organizations.ListRootsInput{})
	if err != nil {
		return nil
	}
	var arns []string
	var walk func(parentID string)
	walk = func(parentID string) {
		for p := organizations.NewListOrganizationalUnitsForParentPaginator(orgs, &organizations.ListOrganizationalUnitsForParentInput{ParentId: &parentID}); p.HasMorePages(); {
			page, err := p.NextPage(awsContext())
			if err != nil {
				return
			}
			for _, ou := range page.OrganizationalUnits {
				if arn := StringValue(ou.Arn); arn != "" {
					arns = append(arns, arn)
				}
				if id := StringValue(ou.Id); id != "" {
					walk(id)
				}
			}
		}
	}
	for _, r := range roots.Roots {
		if id := StringValue(r.Id); id != "" {
			walk(id)
		}
	}
	return arns
}
