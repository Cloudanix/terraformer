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

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

var VpcAllowEmptyValues = []string{"tags."}

type VpcGenerator struct {
	AWSService
}

// isSecondaryVpcCidr reports whether a CIDR-block association should be emitted
// as a standalone aws_vpc_ipv4_cidr_block_association: it needs an association
// ID and must not be the VPC's primary CIDR (which aws_vpc already manages).
func isSecondaryVpcCidr(associationID, assocCidr, primaryCidr string) bool {
	return associationID != "" && assocCidr != primaryCidr
}

func (VpcGenerator) createResources(vpcs *ec2.DescribeVpcsOutput) []terraformutils.Resource {
	var resources []terraformutils.Resource
	for _, vpc := range vpcs.Vpcs {
		vpcID := StringValue(vpc.VpcId)
		resources = append(resources, terraformutils.NewSimpleResource(
			vpcID,
			vpcID,
			"aws_vpc",
			"aws",
			VpcAllowEmptyValues,
		))
		// Secondary IPv4 CIDR associations (the primary CIDR belongs to aws_vpc).
		primaryCidr := StringValue(vpc.CidrBlock)
		for _, assoc := range vpc.CidrBlockAssociationSet {
			assocID := StringValue(assoc.AssociationId)
			if !isSecondaryVpcCidr(assocID, StringValue(assoc.CidrBlock), primaryCidr) {
				continue
			}
			resources = append(resources, terraformutils.NewSimpleResource(
				vpcID+","+assocID, assocID, "aws_vpc_ipv4_cidr_block_association", "aws", VpcAllowEmptyValues))
		}
		for _, assoc := range vpc.Ipv6CidrBlockAssociationSet {
			assocID := StringValue(assoc.AssociationId)
			if assocID == "" {
				continue
			}
			resources = append(resources, terraformutils.NewSimpleResource(
				assocID, assocID, "aws_vpc_ipv6_cidr_block_association", "aws", VpcAllowEmptyValues))
		}
	}
	return resources
}

// Generate TerraformResources from AWS API,
// from each vpc create 1 TerraformResource.
// Need VpcId as ID for terraform resource
func (g *VpcGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := ec2.NewFromConfig(config)
	p := ec2.NewDescribeVpcsPaginator(svc, &ec2.DescribeVpcsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		g.Resources = append(g.Resources, g.createResources(page)...)
	}

	if subs, err := svc.DescribeAwsNetworkPerformanceMetricSubscriptions(context.TODO(),
		&ec2.DescribeAwsNetworkPerformanceMetricSubscriptionsInput{}); err == nil {
		for _, s := range subs.Subscriptions {
			src, dst := StringValue(s.Source), StringValue(s.Destination)
			if src == "" || dst == "" {
				continue
			}
			id := src + "/" + dst + "/" + string(s.Metric) + "/" + string(s.Statistic)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_vpc_network_performance_metric_subscription", "aws", VpcAllowEmptyValues))
		}
	}
	return nil
}
