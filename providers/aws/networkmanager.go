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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type NetworkManagerGenerator struct {
	AWSService
}

func (g *NetworkManagerGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := networkmanager.NewFromConfig(config)
	ctx := context.TODO()

	var globalNetworkIDs []string
	p := networkmanager.NewDescribeGlobalNetworksPaginator(svc, &networkmanager.DescribeGlobalNetworksInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, item := range page.GlobalNetworks {
			id := StringValue(item.GlobalNetworkId)
			if id == "" {
				continue
			}
			globalNetworkIDs = append(globalNetworkIDs, id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_networkmanager_global_network", "aws", defaultAllowEmptyValues))
		}
	}

	add := func(id, tfType string) {
		if id != "" {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, tfType, "aws", defaultAllowEmptyValues))
		}
	}

	for cn := networkmanager.NewListCoreNetworksPaginator(svc, &networkmanager.ListCoreNetworksInput{}); cn.HasMorePages(); {
		page, err := cn.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, c := range page.CoreNetworks {
			add(StringValue(c.CoreNetworkId), "aws_networkmanager_core_network")
		}
	}
	for cp := networkmanager.NewListConnectPeersPaginator(svc, &networkmanager.ListConnectPeersInput{}); cp.HasMorePages(); {
		page, err := cp.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, c := range page.ConnectPeers {
			add(StringValue(c.ConnectPeerId), "aws_networkmanager_connect_peer")
		}
	}

	for pp := networkmanager.NewListPeeringsPaginator(svc, &networkmanager.ListPeeringsInput{}); pp.HasMorePages(); {
		page, err := pp.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, peering := range page.Peerings {
			if peering.PeeringType == "TRANSIT_GATEWAY" {
				add(StringValue(peering.PeeringId), "aws_networkmanager_transit_gateway_peering")
			}
		}
	}

	// Per-global-network children (imported by ARN).
	for _, gnID := range globalNetworkIDs {
		for tp := networkmanager.NewGetTransitGatewayRegistrationsPaginator(svc, &networkmanager.GetTransitGatewayRegistrationsInput{GlobalNetworkId: aws.String(gnID)}); tp.HasMorePages(); {
			page, err := tp.NextPage(ctx)
			if err != nil {
				break
			}
			for _, r := range page.TransitGatewayRegistrations {
				tgwArn := StringValue(r.TransitGatewayArn)
				if tgwArn == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					gnID+","+tgwArn, gnID+"_"+tgwArn, "aws_networkmanager_transit_gateway_registration", "aws", defaultAllowEmptyValues))
			}
		}
		for sp := networkmanager.NewGetSitesPaginator(svc, &networkmanager.GetSitesInput{GlobalNetworkId: aws.String(gnID)}); sp.HasMorePages(); {
			page, err := sp.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, s := range page.Sites {
				add(StringValue(s.SiteArn), "aws_networkmanager_site")
			}
		}
		for dp := networkmanager.NewGetDevicesPaginator(svc, &networkmanager.GetDevicesInput{GlobalNetworkId: aws.String(gnID)}); dp.HasMorePages(); {
			page, err := dp.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, d := range page.Devices {
				add(StringValue(d.DeviceArn), "aws_networkmanager_device")
			}
		}
		for lp := networkmanager.NewGetLinksPaginator(svc, &networkmanager.GetLinksInput{GlobalNetworkId: aws.String(gnID)}); lp.HasMorePages(); {
			page, err := lp.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, l := range page.Links {
				add(StringValue(l.LinkArn), "aws_networkmanager_link")
			}
		}
		for cp := networkmanager.NewGetConnectionsPaginator(svc, &networkmanager.GetConnectionsInput{GlobalNetworkId: aws.String(gnID)}); cp.HasMorePages(); {
			page, err := cp.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, c := range page.Connections {
				add(StringValue(c.ConnectionArn), "aws_networkmanager_connection")
			}
		}
	}
	return nil
}
