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
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

var ec2AllowEmptyValues = []string{"tags."}

type Ec2Generator struct {
	AWSService
}

func (g *Ec2Generator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := ec2.NewFromConfig(config)
	var filters []types.Filter
	for _, filter := range g.Filter {
		if strings.HasPrefix(filter.FieldPath, "tags.") && filter.IsApplicable("instance") {
			filters = append(filters, types.Filter{
				Name:   aws.String("tag:" + strings.TrimPrefix(filter.FieldPath, "tags.")),
				Values: filter.AcceptableValues,
			})
		}
	}
	p := ec2.NewDescribeInstancesPaginator(svc, &ec2.DescribeInstancesInput{
		Filters: filters,
	})
	for p.HasMorePages() {
		page, e := p.NextPage(context.TODO())
		if e != nil {
			return e
		}
		for _, reservation := range page.Reservations {
			for _, instance := range reservation.Instances {
				name := ""
				for _, tag := range instance.Tags {
					if strings.ToLower(*tag.Key) == "name" {
						name = *tag.Value
					}
				}
				attr, err := svc.DescribeInstanceAttribute(context.TODO(), &ec2.DescribeInstanceAttributeInput{
					Attribute:  types.InstanceAttributeNameUserData,
					InstanceId: instance.InstanceId,
				})
				userDataBase64 := ""
				if err == nil && attr.UserData != nil && attr.UserData.Value != nil {
					userDataBase64 = *attr.UserData.Value
				}
				r := terraformutils.NewResource(
					*instance.InstanceId,
					*instance.InstanceId+"_"+name,
					"aws_instance",
					"aws",
					map[string]string{
						"user_data_base64":  userDataBase64,
						"source_dest_check": "true",
					},
					ec2AllowEmptyValues,
					map[string]interface{}{},
				)
				g.Resources = append(g.Resources, r)
			}
		}
	}

	if err := g.loadEc2Extras(svc); err != nil {
		return err
	}
	return nil
}

// loadEc2Extras enumerates standalone EC2 resources that live alongside
// instances: self-owned AMIs, key pairs, placement groups, flow logs,
// self-owned managed prefix lists, DHCP option sets, and egress-only internet
// gateways. AMIs and prefix lists are scoped to "self" so AWS-owned public
// resources aren't dumped.
func (g *Ec2Generator) loadEc2Extras(svc *ec2.Client) error {
	ctx := context.TODO()

	images := ec2.NewDescribeImagesPaginator(svc, &ec2.DescribeImagesInput{Owners: []string{"self"}})
	for images.HasMorePages() {
		page, err := images.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, img := range page.Images {
			id := aws.ToString(img.ImageId)
			if id == "" {
				continue
			}
			name := aws.ToString(img.Name)
			if name == "" {
				name = id
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, name, "aws_ami", "aws", ec2AllowEmptyValues))
		}
	}

	keyPairs, err := svc.DescribeKeyPairs(ctx, &ec2.DescribeKeyPairsInput{})
	if err != nil {
		return err
	}
	for _, kp := range keyPairs.KeyPairs {
		name := aws.ToString(kp.KeyName)
		if name == "" {
			continue
		}
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			name, name, "aws_key_pair", "aws", ec2AllowEmptyValues))
	}

	placementGroups, err := svc.DescribePlacementGroups(ctx, &ec2.DescribePlacementGroupsInput{})
	if err != nil {
		return err
	}
	for _, pg := range placementGroups.PlacementGroups {
		name := aws.ToString(pg.GroupName)
		if name == "" {
			continue
		}
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			name, name, "aws_placement_group", "aws", ec2AllowEmptyValues))
	}

	flowLogs := ec2.NewDescribeFlowLogsPaginator(svc, &ec2.DescribeFlowLogsInput{})
	for flowLogs.HasMorePages() {
		page, err := flowLogs.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, fl := range page.FlowLogs {
			id := aws.ToString(fl.FlowLogId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_flow_log", "aws", ec2AllowEmptyValues))
		}
	}

	prefixLists := ec2.NewDescribeManagedPrefixListsPaginator(svc, &ec2.DescribeManagedPrefixListsInput{
		Filters: []types.Filter{{Name: aws.String("owner-id"), Values: []string{"self"}}},
	})
	for prefixLists.HasMorePages() {
		page, err := prefixLists.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, pl := range page.PrefixLists {
			id := aws.ToString(pl.PrefixListId)
			if id == "" {
				continue
			}
			name := aws.ToString(pl.PrefixListName)
			if name == "" {
				name = id
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, name, "aws_ec2_managed_prefix_list", "aws", ec2AllowEmptyValues))
		}
	}

	dhcpOptions := ec2.NewDescribeDhcpOptionsPaginator(svc, &ec2.DescribeDhcpOptionsInput{})
	for dhcpOptions.HasMorePages() {
		page, err := dhcpOptions.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, d := range page.DhcpOptions {
			id := aws.ToString(d.DhcpOptionsId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_vpc_dhcp_options", "aws", ec2AllowEmptyValues))
		}
	}

	egressGateways := ec2.NewDescribeEgressOnlyInternetGatewaysPaginator(svc, &ec2.DescribeEgressOnlyInternetGatewaysInput{})
	for egressGateways.HasMorePages() {
		page, err := egressGateways.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, eig := range page.EgressOnlyInternetGateways {
			id := aws.ToString(eig.EgressOnlyInternetGatewayId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_egress_only_internet_gateway", "aws", ec2AllowEmptyValues))
		}
	}

	return nil
}

func (g *Ec2Generator) PostConvertHook() error {
	for _, r := range g.Resources {
		if r.InstanceInfo.Type != "aws_instance" {
			continue
		}
		if r.Item["root_block_device"] == nil {
			continue
		}

		rootDeviceVolumeType := r.InstanceState.Attributes["root_block_device.0.volume_type"]
		if !(rootDeviceVolumeType == "io1" || rootDeviceVolumeType == "io2" || rootDeviceVolumeType == "gp3") {
			delete(r.Item["root_block_device"].([]interface{})[0].(map[string]interface{}), "iops")
		}
		if rootDeviceVolumeType != "gp3" {
			delete(r.Item["root_block_device"].([]interface{})[0].(map[string]interface{}), "throughput")
		}
	}

	return nil
}
