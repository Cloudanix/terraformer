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
	if err := g.loadMoreEc2(svc); err != nil {
		return err
	}
	return nil
}

// loadMoreEc2 enumerates additional top-level EC2 resources that each have a
// Describe* paginator returning an id. Snapshots are scoped to "self".
func (g *Ec2Generator) loadMoreEc2(svc *ec2.Client) error {
	ctx := context.TODO()
	add := func(id, tfType string) {
		if id != "" {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, tfType, "aws", ec2AllowEmptyValues))
		}
	}

	snaps := ec2.NewDescribeSnapshotsPaginator(svc, &ec2.DescribeSnapshotsInput{OwnerIds: []string{"self"}})
	for snaps.HasMorePages() {
		p, err := snaps.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, s := range p.Snapshots {
			add(aws.ToString(s.SnapshotId), "aws_ebs_snapshot")
		}
	}
	for p := ec2.NewDescribeCapacityReservationsPaginator(svc, &ec2.DescribeCapacityReservationsInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range pg.CapacityReservations {
			add(aws.ToString(x.CapacityReservationId), "aws_ec2_capacity_reservation")
		}
	}
	for p := ec2.NewDescribeCarrierGatewaysPaginator(svc, &ec2.DescribeCarrierGatewaysInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range pg.CarrierGateways {
			add(aws.ToString(x.CarrierGatewayId), "aws_ec2_carrier_gateway")
		}
	}
	for p := ec2.NewDescribeClientVpnEndpointsPaginator(svc, &ec2.DescribeClientVpnEndpointsInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range pg.ClientVpnEndpoints {
			add(aws.ToString(x.ClientVpnEndpointId), "aws_ec2_client_vpn_endpoint")
		}
	}
	for p := ec2.NewDescribeFleetsPaginator(svc, &ec2.DescribeFleetsInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range pg.Fleets {
			add(aws.ToString(x.FleetId), "aws_ec2_fleet")
		}
	}
	for p := ec2.NewDescribeHostsPaginator(svc, &ec2.DescribeHostsInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range pg.Hosts {
			add(aws.ToString(x.HostId), "aws_ec2_host")
		}
	}
	for p := ec2.NewDescribeTrafficMirrorFiltersPaginator(svc, &ec2.DescribeTrafficMirrorFiltersInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range pg.TrafficMirrorFilters {
			add(aws.ToString(x.TrafficMirrorFilterId), "aws_ec2_traffic_mirror_filter")
		}
	}
	for p := ec2.NewDescribeTrafficMirrorTargetsPaginator(svc, &ec2.DescribeTrafficMirrorTargetsInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range pg.TrafficMirrorTargets {
			add(aws.ToString(x.TrafficMirrorTargetId), "aws_ec2_traffic_mirror_target")
		}
	}
	for p := ec2.NewDescribeTrafficMirrorSessionsPaginator(svc, &ec2.DescribeTrafficMirrorSessionsInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range pg.TrafficMirrorSessions {
			add(aws.ToString(x.TrafficMirrorSessionId), "aws_ec2_traffic_mirror_session")
		}
	}
	for p := ec2.NewDescribeTransitGatewayPeeringAttachmentsPaginator(svc, &ec2.DescribeTransitGatewayPeeringAttachmentsInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range pg.TransitGatewayPeeringAttachments {
			add(aws.ToString(x.TransitGatewayAttachmentId), "aws_ec2_transit_gateway_peering_attachment")
		}
	}
	for p := ec2.NewDescribeSpotInstanceRequestsPaginator(svc, &ec2.DescribeSpotInstanceRequestsInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range pg.SpotInstanceRequests {
			add(aws.ToString(x.SpotInstanceRequestId), "aws_spot_instance_request")
		}
	}
	for p := ec2.NewDescribeSpotFleetRequestsPaginator(svc, &ec2.DescribeSpotFleetRequestsInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range pg.SpotFleetRequestConfigs {
			add(aws.ToString(x.SpotFleetRequestId), "aws_spot_fleet_request")
		}
	}
	for p := ec2.NewDescribeVpcEndpointServiceConfigurationsPaginator(svc, &ec2.DescribeVpcEndpointServiceConfigurationsInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range pg.ServiceConfigurations {
			add(aws.ToString(x.ServiceId), "aws_vpc_endpoint_service")
		}
	}
	for p := ec2.NewDescribeIpamsPaginator(svc, &ec2.DescribeIpamsInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range pg.Ipams {
			add(aws.ToString(x.IpamId), "aws_vpc_ipam")
		}
	}
	for p := ec2.NewDescribeIpamPoolsPaginator(svc, &ec2.DescribeIpamPoolsInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range pg.IpamPools {
			add(aws.ToString(x.IpamPoolId), "aws_vpc_ipam_pool")
		}
	}
	for p := ec2.NewDescribeIpamScopesPaginator(svc, &ec2.DescribeIpamScopesInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range pg.IpamScopes {
			add(aws.ToString(x.IpamScopeId), "aws_vpc_ipam_scope")
		}
	}
	for p := ec2.NewDescribeNetworkInsightsPathsPaginator(svc, &ec2.DescribeNetworkInsightsPathsInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range pg.NetworkInsightsPaths {
			add(aws.ToString(x.NetworkInsightsPathId), "aws_ec2_network_insights_path")
		}
	}
	for p := ec2.NewDescribeNetworkInsightsAnalysesPaginator(svc, &ec2.DescribeNetworkInsightsAnalysesInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range pg.NetworkInsightsAnalyses {
			add(aws.ToString(x.NetworkInsightsAnalysisId), "aws_ec2_network_insights_analysis")
		}
	}
	for p := ec2.NewDescribeTransitGatewayConnectsPaginator(svc, &ec2.DescribeTransitGatewayConnectsInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range pg.TransitGatewayConnects {
			add(aws.ToString(x.TransitGatewayAttachmentId), "aws_ec2_transit_gateway_connect")
		}
	}
	for p := ec2.NewDescribeTransitGatewayMulticastDomainsPaginator(svc, &ec2.DescribeTransitGatewayMulticastDomainsInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range pg.TransitGatewayMulticastDomains {
			add(aws.ToString(x.TransitGatewayMulticastDomainId), "aws_ec2_transit_gateway_multicast_domain")
		}
	}
	for p := ec2.NewDescribeTransitGatewayPolicyTablesPaginator(svc, &ec2.DescribeTransitGatewayPolicyTablesInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range pg.TransitGatewayPolicyTables {
			add(aws.ToString(x.TransitGatewayPolicyTableId), "aws_ec2_transit_gateway_policy_table")
		}
	}
	for p := ec2.NewDescribeSecurityGroupRulesPaginator(svc, &ec2.DescribeSecurityGroupRulesInput{}); p.HasMorePages(); {
		pg, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range pg.SecurityGroupRules {
			id := aws.ToString(x.SecurityGroupRuleId)
			if id == "" {
				continue
			}
			tfType := "aws_vpc_security_group_ingress_rule"
			if aws.ToBool(x.IsEgress) {
				tfType = "aws_vpc_security_group_egress_rule"
			}
			add(id, tfType)
		}
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
