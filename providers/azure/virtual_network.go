// Copyright 2019 The Terraformer Authors.
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

type VirtualNetworkGenerator struct {
	AzureService
}

// initVirtualNetworks imports azurerm_virtual_network (Track 2 armnetwork). The
// PostConvertHook below strips inlined subnets to avoid circular dependencies.
func (g *VirtualNetworkGenerator) initVirtualNetworks() error {
	subscriptionID, cred, opts := g.getClientOptions()
	client, err := armnetwork.NewVirtualNetworksClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	var vnets []*armnetwork.VirtualNetwork
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		pager := client.NewListAllPager(nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return err
			}
			vnets = append(vnets, page.Value...)
		}
	} else {
		for _, rg := range rgs {
			pager := client.NewListPager(rg, nil)
			for pager.More() {
				page, err := pager.NextPage(context.TODO())
				if err != nil {
					return err
				}
				vnets = append(vnets, page.Value...)
			}
		}
	}
	for _, vnet := range vnets {
		id := valueOrEmpty(vnet.ID)
		if id == "" {
			continue
		}
		g.AppendSimpleResourceWithDuplicateCheck(id, valueOrEmpty(vnet.Name), "azurerm_virtual_network")
	}
	return nil
}

func (g *VirtualNetworkGenerator) InitResources() error {
	if _, cred, _ := g.getClientOptions(); cred == nil {
		return nil
	}
	if err := g.initVirtualNetworks(); err != nil {
		return err
	}

	if err := g.initVirtualNetworkGateways(); err != nil {
		return err
	}

	if err := g.initLocalNetworkGateways(); err != nil {
		return err
	}

	if err := g.initVirtualNetworkGatewayConnections(); err != nil {
		return err
	}

	if err := g.initVirtualNetworkPeerings(); err != nil {
		return err
	}

	return nil
}

// initVirtualNetworkGateways enumerates azurerm_virtual_network_gateway via the
// Track 2 armnetwork SDK. Virtual network gateways are resource-group scoped
// (the API has no subscription-wide list), so this requires -R and is skipped
// when no Track 2 credential is available (dual SDK path).
func (g *VirtualNetworkGenerator) initVirtualNetworkGateways() error {
	subscriptionID, cred, opts := g.getClientOptions()
	if cred == nil {
		return nil
	}
	client, err := armnetwork.NewVirtualNetworkGatewaysClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	for _, rg := range g.resourceGroups() {
		pager := client.NewListPager(rg, nil)
		if err := appendFromPager(&g.AzureService, pager,
			func(p armnetwork.VirtualNetworkGatewaysClientListResponse) []*armnetwork.VirtualNetworkGateway {
				return p.Value
			},
			func(i *armnetwork.VirtualNetworkGateway) string { return valueOrEmpty(i.ID) },
			func(i *armnetwork.VirtualNetworkGateway) string { return valueOrEmpty(i.Name) },
			"azurerm_virtual_network_gateway"); err != nil {
			return err
		}
	}
	return nil
}

// initLocalNetworkGateways enumerates azurerm_local_network_gateway via Track 2.
// Resource-group scoped; requires -R; skipped when no Track 2 credential.
func (g *VirtualNetworkGenerator) initLocalNetworkGateways() error {
	subscriptionID, cred, opts := g.getClientOptions()
	if cred == nil {
		return nil
	}
	client, err := armnetwork.NewLocalNetworkGatewaysClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	for _, rg := range g.resourceGroups() {
		pager := client.NewListPager(rg, nil)
		if err := appendFromPager(&g.AzureService, pager,
			func(p armnetwork.LocalNetworkGatewaysClientListResponse) []*armnetwork.LocalNetworkGateway {
				return p.Value
			},
			func(i *armnetwork.LocalNetworkGateway) string { return valueOrEmpty(i.ID) },
			func(i *armnetwork.LocalNetworkGateway) string { return valueOrEmpty(i.Name) },
			"azurerm_local_network_gateway"); err != nil {
			return err
		}
	}
	return nil
}

// initVirtualNetworkGatewayConnections enumerates
// azurerm_virtual_network_gateway_connection via Track 2. RG-scoped.
func (g *VirtualNetworkGenerator) initVirtualNetworkGatewayConnections() error {
	subscriptionID, cred, opts := g.getClientOptions()
	if cred == nil {
		return nil
	}
	client, err := armnetwork.NewVirtualNetworkGatewayConnectionsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	for _, rg := range g.resourceGroups() {
		pager := client.NewListPager(rg, nil)
		if err := appendFromPager(&g.AzureService, pager,
			func(p armnetwork.VirtualNetworkGatewayConnectionsClientListResponse) []*armnetwork.VirtualNetworkGatewayConnection {
				return p.Value
			},
			func(i *armnetwork.VirtualNetworkGatewayConnection) string { return valueOrEmpty(i.ID) },
			func(i *armnetwork.VirtualNetworkGatewayConnection) string { return valueOrEmpty(i.Name) },
			"azurerm_virtual_network_gateway_connection"); err != nil {
			return err
		}
	}
	return nil
}

// initVirtualNetworkPeerings enumerates azurerm_virtual_network_peering via
// Track 2. Peerings are nested under each virtual network, so this lists the
// VNets per resource group first, then the peerings per VNet. RG-scoped.
func (g *VirtualNetworkGenerator) initVirtualNetworkPeerings() error {
	subscriptionID, cred, opts := g.getClientOptions()
	if cred == nil {
		return nil
	}
	vnetClient, err := armnetwork.NewVirtualNetworksClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	peerClient, err := armnetwork.NewVirtualNetworkPeeringsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	for _, rg := range g.resourceGroups() {
		vnetNames, err := g.listVNetNames(vnetClient, rg)
		if err != nil {
			return err
		}
		for _, vnet := range vnetNames {
			pager := peerClient.NewListPager(rg, vnet, nil)
			if err := appendFromPager(&g.AzureService, pager,
				func(p armnetwork.VirtualNetworkPeeringsClientListResponse) []*armnetwork.VirtualNetworkPeering {
					return p.Value
				},
				func(i *armnetwork.VirtualNetworkPeering) string { return valueOrEmpty(i.ID) },
				func(i *armnetwork.VirtualNetworkPeering) string { return valueOrEmpty(i.Name) },
				"azurerm_virtual_network_peering"); err != nil {
				return err
			}
		}
	}
	return nil
}

// listVNetNames returns the virtual-network names in a resource group (Track 2),
// used to scope nested sub-resource enumeration (peerings).
func (g *VirtualNetworkGenerator) listVNetNames(client *armnetwork.VirtualNetworksClient, rg string) ([]string, error) {
	var names []string
	pager := client.NewListPager(rg, nil)
	for pager.More() {
		page, err := pager.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}
		for _, v := range page.Value {
			if v != nil && valueOrEmpty(v.Name) != "" {
				names = append(names, *v.Name)
			}
		}
	}
	return names, nil
}

// NOTE on Virtual Networks and Subnet's:
// Terraform currently provides both a standalone Subnet resource, and allows for Subnets to be defined in-line within the Virtual Network
// resource. At this time you cannot use a Virtual Network with in-line Subnets in conjunction with any Subnet resources.
// Doing so will cause a conflict of Subnet configurations and will overwrite Subnet's.
func (g *VirtualNetworkGenerator) PostConvertHook() error {
	for _, resource := range g.Resources {
		if resource.InstanceInfo.Type != "azurerm_virtual_network" {
			continue
		}
		delete(resource.Item, "subnet")
	}
	return nil
}
