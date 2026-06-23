// Copyright 2021 The Terraformer Authors.
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

package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v7"
)

type SubnetGenerator struct {
	AzureService
}

// listSubnets enumerates subnets across all virtual networks (Track 2 armnetwork).
func (g *SubnetGenerator) listSubnets() ([]*armnetwork.Subnet, error) {
	subscriptionID, cred, opts := g.getClientOptions()
	vnetClient, err := armnetwork.NewVirtualNetworksClient(subscriptionID, cred, opts)
	if err != nil {
		return nil, err
	}
	subnetClient, err := armnetwork.NewSubnetsClient(subscriptionID, cred, opts)
	if err != nil {
		return nil, err
	}

	var vnets []*armnetwork.VirtualNetwork
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		pager := vnetClient.NewListAllPager(nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return nil, err
			}
			vnets = append(vnets, page.Value...)
		}
	} else {
		for _, rg := range rgs {
			pager := vnetClient.NewListPager(rg, nil)
			for pager.More() {
				page, err := pager.NextPage(context.TODO())
				if err != nil {
					return nil, err
				}
				vnets = append(vnets, page.Value...)
			}
		}
	}

	var subnets []*armnetwork.Subnet
	for _, vnet := range vnets {
		vnetID := valueOrEmpty(vnet.ID)
		if vnetID == "" {
			continue
		}
		parsed, err := ParseAzureResourceID(vnetID)
		if err != nil {
			return nil, err
		}
		pager := subnetClient.NewListPager(parsed.ResourceGroup, valueOrEmpty(vnet.Name), nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return nil, err
			}
			subnets = append(subnets, page.Value...)
		}
	}
	return subnets, nil
}

func (g *SubnetGenerator) appendSubnet(subnet *armnetwork.Subnet) {
	g.AppendSimpleResource(valueOrEmpty(subnet.ID), valueOrEmpty(subnet.Name), "azurerm_subnet")
}

func (g *SubnetGenerator) appendRouteTable(subnet *armnetwork.Subnet) {
	if props := subnet.Properties; props != nil && props.RouteTable != nil {
		g.appendSimpleAssociation(
			valueOrEmpty(subnet.ID), valueOrEmpty(subnet.Name), props.RouteTable.Name,
			"azurerm_subnet_route_table_association",
			map[string]string{
				"subnet_id":      valueOrEmpty(subnet.ID),
				"route_table_id": valueOrEmpty(props.RouteTable.ID),
			})
	}
}

func (g *SubnetGenerator) appendNetworkSecurityGroupAssociation(subnet *armnetwork.Subnet) {
	if props := subnet.Properties; props != nil && props.NetworkSecurityGroup != nil {
		g.appendSimpleAssociation(
			valueOrEmpty(subnet.ID), valueOrEmpty(subnet.Name), props.NetworkSecurityGroup.Name,
			"azurerm_subnet_network_security_group_association",
			map[string]string{
				"subnet_id":                 valueOrEmpty(subnet.ID),
				"network_security_group_id": valueOrEmpty(props.NetworkSecurityGroup.ID),
			})
	}
}

func (g *SubnetGenerator) appendNatGateway(subnet *armnetwork.Subnet) {
	if props := subnet.Properties; props != nil && props.NatGateway != nil {
		g.appendSimpleAssociation(
			valueOrEmpty(subnet.ID), valueOrEmpty(subnet.Name), nil,
			"azurerm_subnet_nat_gateway_association",
			map[string]string{
				"subnet_id":      valueOrEmpty(subnet.ID),
				"nat_gateway_id": valueOrEmpty(props.NatGateway.ID),
			})
	}
}

func (g *SubnetGenerator) appendServiceEndpointPolicies() error {
	subscriptionID, cred, opts := g.getClientOptions()
	client, err := armnetwork.NewServiceEndpointPoliciesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	id := func(i *armnetwork.ServiceEndpointPolicy) string { return valueOrEmpty(i.ID) }
	name := func(i *armnetwork.ServiceEndpointPolicy) string { return valueOrEmpty(i.Name) }
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		return appendFromPager(&g.AzureService, client.NewListPager(nil),
			func(p armnetwork.ServiceEndpointPoliciesClientListResponse) []*armnetwork.ServiceEndpointPolicy {
				return p.Value
			},
			id, name, "azurerm_subnet_service_endpoint_storage_policy")
	}
	for _, rg := range rgs {
		if err := appendFromPager(&g.AzureService, client.NewListByResourceGroupPager(rg, nil),
			func(p armnetwork.ServiceEndpointPoliciesClientListByResourceGroupResponse) []*armnetwork.ServiceEndpointPolicy {
				return p.Value
			},
			id, name, "azurerm_subnet_service_endpoint_storage_policy"); err != nil {
			return err
		}
	}
	return nil
}

// InitResources imports azurerm_subnet (+ route_table/nsg/nat_gateway
// associations) and azurerm_subnet_service_endpoint_storage_policy. Migrated to
// the Track 2 armnetwork SDK. The PostConvertHook strips address_prefix.
func (g *SubnetGenerator) InitResources() error {
	if _, cred, _ := g.getClientOptions(); cred == nil {
		return nil
	}
	subnets, err := g.listSubnets()
	if err != nil {
		return err
	}
	for _, subnet := range subnets {
		if subnet == nil {
			continue
		}
		g.appendSubnet(subnet)
		g.appendRouteTable(subnet)
		g.appendNetworkSecurityGroupAssociation(subnet)
		g.appendNatGateway(subnet)
	}
	return g.appendServiceEndpointPolicies()
}

func (g *SubnetGenerator) PostConvertHook() error {
	for _, resource := range g.Resources {
		if resource.InstanceInfo.Type != "azurerm_subnet" {
			continue
		}
		delete(resource.Item, "address_prefix")
	}
	return nil
}
