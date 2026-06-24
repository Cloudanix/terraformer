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
	"github.com/aws/aws-sdk-go-v2/service/ec2"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type VPCRouteServerGenerator struct {
	AWSService
}

// InitResources enumerates VPC route servers, their endpoints/peers, and each
// route server's VPC associations and route-table propagations.
func (g *VPCRouteServerGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := ec2.NewFromConfig(config)
	ctx := awsContext()

	for p := ec2.NewDescribeRouteServersPaginator(svc, &ec2.DescribeRouteServersInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, rs := range page.RouteServers {
			id := StringValue(rs.RouteServerId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_vpc_route_server", "aws", defaultAllowEmptyValues))
			if assoc, err := svc.GetRouteServerAssociations(ctx, &ec2.GetRouteServerAssociationsInput{RouteServerId: aws.String(id)}); err == nil {
				for _, a := range assoc.RouteServerAssociations {
					vpcID := StringValue(a.VpcId)
					if vpcID == "" {
						continue
					}
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id+","+vpcID, id+"_"+vpcID, "aws_vpc_route_server_vpc_association", "aws", defaultAllowEmptyValues))
				}
			}
			if prop, err := svc.GetRouteServerPropagations(ctx, &ec2.GetRouteServerPropagationsInput{RouteServerId: aws.String(id)}); err == nil {
				for _, pr := range prop.RouteServerPropagations {
					rtb := StringValue(pr.RouteTableId)
					if rtb == "" {
						continue
					}
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id+","+rtb, id+"_"+rtb, "aws_vpc_route_server_propagation", "aws", defaultAllowEmptyValues))
				}
			}
		}
	}
	for p := ec2.NewDescribeRouteServerEndpointsPaginator(svc, &ec2.DescribeRouteServerEndpointsInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, ep := range page.RouteServerEndpoints {
			if id := StringValue(ep.RouteServerEndpointId); id != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, id, "aws_vpc_route_server_endpoint", "aws", defaultAllowEmptyValues))
			}
		}
	}
	for p := ec2.NewDescribeRouteServerPeersPaginator(svc, &ec2.DescribeRouteServerPeersInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, pr := range page.RouteServerPeers {
			if id := StringValue(pr.RouteServerPeerId); id != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, id, "aws_vpc_route_server_peer", "aws", defaultAllowEmptyValues))
			}
		}
	}
	return nil
}
