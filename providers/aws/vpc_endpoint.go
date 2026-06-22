// Copyright 2023 The Terraformer Authors.
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

var VpcEndpointAllowEmptyValues = []string{"tags."}

type VpcEndpointGenerator struct {
	AWSService
}

func (g *VpcEndpointGenerator) createResources(vpceps *ec2.DescribeVpcEndpointsOutput) []terraformutils.Resource {
	var resources []terraformutils.Resource
	for _, vpcEndpoint := range vpceps.VpcEndpoints {
		id := StringValue(vpcEndpoint.VpcEndpointId)
		resources = append(resources, terraformutils.NewSimpleResource(
			id,
			id,
			"aws_vpc_endpoint",
			"aws",
			VpcAllowEmptyValues,
		))
		if shouldEmitVpcEndpointPrivateDNS(string(vpcEndpoint.VpcEndpointType), vpcEndpoint.PrivateDnsEnabled) {
			resources = append(resources, terraformutils.NewSimpleResource(
				id, id, "aws_vpc_endpoint_private_dns", "aws", VpcEndpointAllowEmptyValues))
		}
	}
	return resources
}

// shouldEmitVpcEndpointPrivateDNS reports whether a VPC endpoint warrants a
// separate aws_vpc_endpoint_private_dns resource: only Interface endpoints with
// private DNS enabled carry that managed sub-resource.
func shouldEmitVpcEndpointPrivateDNS(endpointType string, privateDNSEnabled *bool) bool {
	return endpointType == "Interface" && privateDNSEnabled != nil && *privateDNSEnabled
}

// Generate TerraformResources from AWS API,
// from each vpc endpoint create 1 TerraformResource.
// Need VpcEndpointId as ID for terraform resource
func (g *VpcEndpointGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := ec2.NewFromConfig(config)
	vpceps, err := svc.DescribeVpcEndpoints(context.TODO(), &ec2.DescribeVpcEndpointsInput{})
	if err != nil {
		return err
	}
	g.Resources = g.createResources(vpceps)

	np := ec2.NewDescribeVpcEndpointConnectionNotificationsPaginator(svc, &ec2.DescribeVpcEndpointConnectionNotificationsInput{})
	for np.HasMorePages() {
		page, err := np.NextPage(context.TODO())
		if err != nil {
			break
		}
		for _, n := range page.ConnectionNotificationSet {
			id := StringValue(n.ConnectionNotificationId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_vpc_endpoint_connection_notification", "aws", VpcEndpointAllowEmptyValues))
		}
	}
	return nil
}
