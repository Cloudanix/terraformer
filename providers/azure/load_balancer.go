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

package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v7"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type LoadBalancerGenerator struct {
	AzureService
}

// drainLBChildren walks a load-balancer sub-resource pager and appends each item
// with the parent loadbalancer_id attribute that the azurerm data sources need.
func drainLBChildren[P any, I any](
	g *LoadBalancerGenerator,
	pager *runtime.Pager[P],
	loadBalancerID, tfType string,
	values func(P) []*I,
	id func(*I) string,
	name func(*I) string,
) error {
	for pager.More() {
		page, err := pager.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, item := range values(page) {
			if item == nil {
				continue
			}
			resourceID := id(item)
			if resourceID == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewResource(
				resourceID, name(item), tfType, g.ProviderName,
				map[string]string{"loadbalancer_id": loadBalancerID},
				[]string{},
				map[string]interface{}{},
			))
		}
	}
	return nil
}

func (g *LoadBalancerGenerator) listLoadBalancers(client *armnetwork.LoadBalancersClient) ([]*armnetwork.LoadBalancer, error) {
	var lbs []*armnetwork.LoadBalancer
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		pager := client.NewListAllPager(nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return nil, err
			}
			lbs = append(lbs, page.Value...)
		}
		return lbs, nil
	}
	for _, rg := range rgs {
		pager := client.NewListPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return nil, err
			}
			lbs = append(lbs, page.Value...)
		}
	}
	return lbs, nil
}

// InitResources imports azurerm_lb and its probes, NAT rules, backend address
// pools, load-balancing rules and outbound rules. Migrated to Track 2 armnetwork.
func (g *LoadBalancerGenerator) InitResources() error {
	subscriptionID, cred, opts := g.getClientOptions()
	if cred == nil {
		return nil
	}
	lbClient, err := armnetwork.NewLoadBalancersClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	probesClient, err := armnetwork.NewLoadBalancerProbesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	natClient, err := armnetwork.NewInboundNatRulesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	poolsClient, err := armnetwork.NewLoadBalancerBackendAddressPoolsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	rulesClient, err := armnetwork.NewLoadBalancerLoadBalancingRulesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	outboundClient, err := armnetwork.NewLoadBalancerOutboundRulesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}

	lbs, err := g.listLoadBalancers(lbClient)
	if err != nil {
		return err
	}
	for _, lb := range lbs {
		lbID := valueOrEmpty(lb.ID)
		if lbID == "" {
			continue
		}
		g.AppendSimpleResource(lbID, valueOrEmpty(lb.Name), "azurerm_lb")
		parsed, err := ParseAzureResourceID(lbID)
		if err != nil {
			return err
		}
		rg, lbName := parsed.ResourceGroup, valueOrEmpty(lb.Name)

		if err := drainLBChildren(g, probesClient.NewListPager(rg, lbName, nil), lbID, "azurerm_lb_probe",
			func(p armnetwork.LoadBalancerProbesClientListResponse) []*armnetwork.Probe { return p.Value },
			func(i *armnetwork.Probe) string { return valueOrEmpty(i.ID) },
			func(i *armnetwork.Probe) string { return valueOrEmpty(i.Name) }); err != nil {
			return err
		}
		if err := drainLBChildren(g, natClient.NewListPager(rg, lbName, nil), lbID, "azurerm_lb_nat_rule",
			func(p armnetwork.InboundNatRulesClientListResponse) []*armnetwork.InboundNatRule { return p.Value },
			func(i *armnetwork.InboundNatRule) string { return valueOrEmpty(i.ID) },
			func(i *armnetwork.InboundNatRule) string { return valueOrEmpty(i.Name) }); err != nil {
			return err
		}
		if err := drainLBChildren(g, poolsClient.NewListPager(rg, lbName, nil), lbID, "azurerm_lb_backend_address_pool",
			func(p armnetwork.LoadBalancerBackendAddressPoolsClientListResponse) []*armnetwork.BackendAddressPool {
				return p.Value
			},
			func(i *armnetwork.BackendAddressPool) string { return valueOrEmpty(i.ID) },
			func(i *armnetwork.BackendAddressPool) string { return valueOrEmpty(i.Name) }); err != nil {
			return err
		}
		if err := drainLBChildren(g, rulesClient.NewListPager(rg, lbName, nil), lbID, "azurerm_lb_rule",
			func(p armnetwork.LoadBalancerLoadBalancingRulesClientListResponse) []*armnetwork.LoadBalancingRule {
				return p.Value
			},
			func(i *armnetwork.LoadBalancingRule) string { return valueOrEmpty(i.ID) },
			func(i *armnetwork.LoadBalancingRule) string { return valueOrEmpty(i.Name) }); err != nil {
			return err
		}
		if err := drainLBChildren(g, outboundClient.NewListPager(rg, lbName, nil), lbID, "azurerm_lb_outbound_rule",
			func(p armnetwork.LoadBalancerOutboundRulesClientListResponse) []*armnetwork.OutboundRule {
				return p.Value
			},
			func(i *armnetwork.OutboundRule) string { return valueOrEmpty(i.ID) },
			func(i *armnetwork.OutboundRule) string { return valueOrEmpty(i.Name) }); err != nil {
			return err
		}
	}
	return nil
}
