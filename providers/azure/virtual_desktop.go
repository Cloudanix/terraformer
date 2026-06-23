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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/desktopvirtualization/armdesktopvirtualization"
)

type VirtualDesktopGenerator struct {
	AzureService
}

func (g *VirtualDesktopGenerator) InitResources() error {
	if _, cred, _ := g.getClientOptions(); cred == nil {
		return nil
	}
	for _, fn := range []func() error{
		g.initHostPools,
		g.initWorkspaces,
		g.initScalingPlans,
		g.initApplicationGroups,
	} {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}

func (g *VirtualDesktopGenerator) initHostPools() error {
	subscriptionID, cred, opts := g.getClientOptions()
	client, err := armdesktopvirtualization.NewHostPoolsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	id := func(i *armdesktopvirtualization.HostPool) string { return valueOrEmpty(i.ID) }
	name := func(i *armdesktopvirtualization.HostPool) string { return valueOrEmpty(i.Name) }
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		return appendFromPager(&g.AzureService, client.NewListPager(nil),
			func(p armdesktopvirtualization.HostPoolsClientListResponse) []*armdesktopvirtualization.HostPool {
				return p.Value
			},
			id, name, "azurerm_virtual_desktop_host_pool")
	}
	for _, rg := range rgs {
		if err := appendFromPager(&g.AzureService, client.NewListByResourceGroupPager(rg, nil),
			func(p armdesktopvirtualization.HostPoolsClientListByResourceGroupResponse) []*armdesktopvirtualization.HostPool {
				return p.Value
			},
			id, name, "azurerm_virtual_desktop_host_pool"); err != nil {
			return err
		}
	}
	return nil
}

func (g *VirtualDesktopGenerator) initWorkspaces() error {
	subscriptionID, cred, opts := g.getClientOptions()
	client, err := armdesktopvirtualization.NewWorkspacesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	id := func(i *armdesktopvirtualization.Workspace) string { return valueOrEmpty(i.ID) }
	name := func(i *armdesktopvirtualization.Workspace) string { return valueOrEmpty(i.Name) }
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		return appendFromPager(&g.AzureService, client.NewListBySubscriptionPager(nil),
			func(p armdesktopvirtualization.WorkspacesClientListBySubscriptionResponse) []*armdesktopvirtualization.Workspace {
				return p.Value
			},
			id, name, "azurerm_virtual_desktop_workspace")
	}
	for _, rg := range rgs {
		if err := appendFromPager(&g.AzureService, client.NewListByResourceGroupPager(rg, nil),
			func(p armdesktopvirtualization.WorkspacesClientListByResourceGroupResponse) []*armdesktopvirtualization.Workspace {
				return p.Value
			},
			id, name, "azurerm_virtual_desktop_workspace"); err != nil {
			return err
		}
	}
	return nil
}

func (g *VirtualDesktopGenerator) initScalingPlans() error {
	subscriptionID, cred, opts := g.getClientOptions()
	client, err := armdesktopvirtualization.NewScalingPlansClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	id := func(i *armdesktopvirtualization.ScalingPlan) string { return valueOrEmpty(i.ID) }
	name := func(i *armdesktopvirtualization.ScalingPlan) string { return valueOrEmpty(i.Name) }
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		return appendFromPager(&g.AzureService, client.NewListBySubscriptionPager(nil),
			func(p armdesktopvirtualization.ScalingPlansClientListBySubscriptionResponse) []*armdesktopvirtualization.ScalingPlan {
				return p.Value
			},
			id, name, "azurerm_virtual_desktop_scaling_plan")
	}
	for _, rg := range rgs {
		if err := appendFromPager(&g.AzureService, client.NewListByResourceGroupPager(rg, nil),
			func(p armdesktopvirtualization.ScalingPlansClientListByResourceGroupResponse) []*armdesktopvirtualization.ScalingPlan {
				return p.Value
			},
			id, name, "azurerm_virtual_desktop_scaling_plan"); err != nil {
			return err
		}
	}
	return nil
}

func (g *VirtualDesktopGenerator) initApplicationGroups() error {
	subscriptionID, cred, opts := g.getClientOptions()
	client, err := armdesktopvirtualization.NewApplicationGroupsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	id := func(i *armdesktopvirtualization.ApplicationGroup) string { return valueOrEmpty(i.ID) }
	name := func(i *armdesktopvirtualization.ApplicationGroup) string { return valueOrEmpty(i.Name) }
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		return appendFromPager(&g.AzureService, client.NewListBySubscriptionPager(nil),
			func(p armdesktopvirtualization.ApplicationGroupsClientListBySubscriptionResponse) []*armdesktopvirtualization.ApplicationGroup {
				return p.Value
			},
			id, name, "azurerm_virtual_desktop_application_group")
	}
	for _, rg := range rgs {
		if err := appendFromPager(&g.AzureService, client.NewListByResourceGroupPager(rg, nil),
			func(p armdesktopvirtualization.ApplicationGroupsClientListByResourceGroupResponse) []*armdesktopvirtualization.ApplicationGroup {
				return p.Value
			},
			id, name, "azurerm_virtual_desktop_application_group"); err != nil {
			return err
		}
	}
	return nil
}
