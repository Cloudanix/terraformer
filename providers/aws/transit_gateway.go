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
	"log"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

var tgwAllowEmptyValues = []string{"tags."}

type TransitGatewayGenerator struct {
	AWSService
}

func (g *TransitGatewayGenerator) getTransitGateways(svc *ec2.Client) error {
	p := ec2.NewDescribeTransitGatewaysPaginator(svc, &ec2.DescribeTransitGatewaysInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, tgw := range page.TransitGateways {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(tgw.TransitGatewayId),
				StringValue(tgw.TransitGatewayId),
				"aws_ec2_transit_gateway",
				"aws",
				tgwAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *TransitGatewayGenerator) getTransitGatewayRouteTables(svc *ec2.Client) error {
	p := ec2.NewDescribeTransitGatewayRouteTablesPaginator(svc, &ec2.DescribeTransitGatewayRouteTablesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, tgwrt := range page.TransitGatewayRouteTables {
			rtbID := StringValue(tgwrt.TransitGatewayRouteTableId)
			// Default route tables are auto-created with the tgw, so not emitted as
			// aws_ec2_transit_gateway_route_table, but their associations/propagations
			// are still importable resources.
			if !*tgwrt.DefaultAssociationRouteTable {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					rtbID, rtbID, "aws_ec2_transit_gateway_route_table", "aws", tgwAllowEmptyValues))
			}
			g.loadRouteTableAssociations(svc, rtbID)
		}
	}
	return nil
}

func (g *TransitGatewayGenerator) loadRouteTableAssociations(svc *ec2.Client, rtbID string) {
	if rtbID == "" {
		return
	}
	ctx := context.TODO()
	rtb := rtbID
	for p := ec2.NewGetTransitGatewayRouteTableAssociationsPaginator(svc, &ec2.GetTransitGatewayRouteTableAssociationsInput{TransitGatewayRouteTableId: &rtb}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, a := range page.Associations {
			att := StringValue(a.TransitGatewayAttachmentId)
			if att == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				rtb+"_"+att, rtb+"_"+att, "aws_ec2_transit_gateway_route_table_association", "aws", tgwAllowEmptyValues))
		}
	}
	for p := ec2.NewGetTransitGatewayRouteTablePropagationsPaginator(svc, &ec2.GetTransitGatewayRouteTablePropagationsInput{TransitGatewayRouteTableId: &rtb}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, prop := range page.TransitGatewayRouteTablePropagations {
			att := StringValue(prop.TransitGatewayAttachmentId)
			if att == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				rtb+"_"+att, rtb+"_"+att, "aws_ec2_transit_gateway_route_table_propagation", "aws", tgwAllowEmptyValues))
		}
	}
	for p := ec2.NewGetTransitGatewayPrefixListReferencesPaginator(svc, &ec2.GetTransitGatewayPrefixListReferencesInput{TransitGatewayRouteTableId: &rtb}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, ref := range page.TransitGatewayPrefixListReferences {
			plID := StringValue(ref.PrefixListId)
			if plID == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				rtb+"_"+plID, rtb+"_"+plID, "aws_ec2_transit_gateway_prefix_list_reference", "aws", tgwAllowEmptyValues))
		}
	}
	// Static routes only (propagated routes are not Terraform-managed).
	if routes, err := svc.SearchTransitGatewayRoutes(ctx, &ec2.SearchTransitGatewayRoutesInput{
		TransitGatewayRouteTableId: &rtb,
		Filters:                    []ec2types.Filter{{Name: aws.String("type"), Values: []string{"static"}}},
	}); err == nil {
		for _, r := range routes.Routes {
			cidr := StringValue(r.DestinationCidrBlock)
			if cidr == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				rtb+"_"+cidr, rtb+"_"+cidr, "aws_ec2_transit_gateway_route", "aws", tgwAllowEmptyValues))
		}
	}
}

func (g *TransitGatewayGenerator) getTransitGatewayVpcAttachments(svc *ec2.Client) error {
	p := ec2.NewDescribeTransitGatewayVpcAttachmentsPaginator(svc, &ec2.DescribeTransitGatewayVpcAttachmentsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, tgwa := range page.TransitGatewayVpcAttachments {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(tgwa.TransitGatewayAttachmentId),
				StringValue(tgwa.TransitGatewayAttachmentId),
				"aws_ec2_transit_gateway_vpc_attachment",
				"aws",
				tgwAllowEmptyValues,
			))
		}
	}
	return nil
}

// Generate TerraformResources from AWS API,
// from each customer gateway create 1 TerraformResource.
// Need CustomerGatewayId as ID for terraform resource
func (g *TransitGatewayGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := ec2.NewFromConfig(config)
	g.Resources = []terraformutils.Resource{}
	err := g.getTransitGateways(svc)
	if err != nil {
		log.Println(err)
	}

	err = g.getTransitGatewayRouteTables(svc)
	if err != nil {
		log.Println(err)
	}

	err = g.getTransitGatewayVpcAttachments(svc)
	if err != nil {
		log.Println(err)
	}

	return nil
}
