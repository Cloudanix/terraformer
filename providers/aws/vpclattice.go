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

	"github.com/aws/aws-sdk-go-v2/service/vpclattice"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type VPCLatticeGenerator struct {
	AWSService
}

// InitResources enumerates VPC Lattice services and service networks. Import IDs
// are the resource id.
func (g *VPCLatticeGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := vpclattice.NewFromConfig(config)
	ctx := context.TODO()

	services := vpclattice.NewListServicesPaginator(svc, &vpclattice.ListServicesInput{})
	for services.HasMorePages() {
		page, err := services.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, s := range page.Items {
			id := StringValue(s.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(s.Name), "aws_vpclattice_service", "aws", defaultAllowEmptyValues))
		}
	}

	networks := vpclattice.NewListServiceNetworksPaginator(svc, &vpclattice.ListServiceNetworksInput{})
	for networks.HasMorePages() {
		page, err := networks.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, n := range page.Items {
			id := StringValue(n.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(n.Name), "aws_vpclattice_service_network", "aws", defaultAllowEmptyValues))
		}
	}

	targetGroups := vpclattice.NewListTargetGroupsPaginator(svc, &vpclattice.ListTargetGroupsInput{})
	for targetGroups.HasMorePages() {
		page, err := targetGroups.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, t := range page.Items {
			id := StringValue(t.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(t.Name), "aws_vpclattice_target_group", "aws", defaultAllowEmptyValues))
		}
	}

	snVpcAssocs := vpclattice.NewListServiceNetworkVpcAssociationsPaginator(svc, &vpclattice.ListServiceNetworkVpcAssociationsInput{})
	for snVpcAssocs.HasMorePages() {
		page, err := snVpcAssocs.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, a := range page.Items {
			id := StringValue(a.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_vpclattice_service_network_vpc_association", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
