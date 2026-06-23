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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
)

type VirtualMachineGenerator struct {
	AzureService
}

// vmResourceType picks the azurerm resource type for a VM by OS. It mirrors the
// original heuristic: prefer the OS profile's WindowsConfiguration; if there is
// no OS profile, fall back to the OS disk's OSType.
func vmResourceType(vm *armcompute.VirtualMachine) string {
	windows := false
	if props := vm.Properties; props != nil {
		switch {
		case props.OSProfile != nil:
			windows = props.OSProfile.WindowsConfiguration != nil
		case props.StorageProfile != nil && props.StorageProfile.OSDisk != nil && props.StorageProfile.OSDisk.OSType != nil:
			windows = *props.StorageProfile.OSDisk.OSType == armcompute.OperatingSystemTypesWindows
		}
	}
	if windows {
		return "azurerm_windows_virtual_machine"
	}
	return "azurerm_linux_virtual_machine"
}

// InitResources imports azurerm_linux_virtual_machine / _windows_virtual_machine.
// Migrated to the Track 2 armcompute SDK (was Track 1 services/compute).
func (g *VirtualMachineGenerator) InitResources() error {
	subscriptionID, cred, opts := g.getClientOptions()
	if cred == nil {
		return nil
	}
	client, err := armcompute.NewVirtualMachinesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}

	var vms []*armcompute.VirtualMachine
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		pager := client.NewListAllPager(nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return err
			}
			vms = append(vms, page.Value...)
		}
	} else {
		for _, rg := range rgs {
			pager := client.NewListPager(rg, nil)
			for pager.More() {
				page, err := pager.NextPage(context.TODO())
				if err != nil {
					return err
				}
				vms = append(vms, page.Value...)
			}
		}
	}

	for _, vm := range vms {
		if vm == nil {
			continue
		}
		id := valueOrEmpty(vm.ID)
		if id == "" {
			continue
		}
		g.AppendSimpleResource(id, valueOrEmpty(vm.Name), vmResourceType(vm))
	}
	return nil
}
