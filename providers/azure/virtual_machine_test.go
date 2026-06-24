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
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
)

func TestVMResourceType(t *testing.T) {
	win := armcompute.OperatingSystemTypesWindows
	lin := armcompute.OperatingSystemTypesLinux

	windowsByProfile := &armcompute.VirtualMachine{Properties: &armcompute.VirtualMachineProperties{
		OSProfile: &armcompute.OSProfile{WindowsConfiguration: &armcompute.WindowsConfiguration{}},
	}}
	linuxByProfile := &armcompute.VirtualMachine{Properties: &armcompute.VirtualMachineProperties{
		OSProfile: &armcompute.OSProfile{},
	}}
	windowsByDisk := &armcompute.VirtualMachine{Properties: &armcompute.VirtualMachineProperties{
		StorageProfile: &armcompute.StorageProfile{OSDisk: &armcompute.OSDisk{OSType: &win}},
	}}
	linuxByDisk := &armcompute.VirtualMachine{Properties: &armcompute.VirtualMachineProperties{
		StorageProfile: &armcompute.StorageProfile{OSDisk: &armcompute.OSDisk{OSType: &lin}},
	}}
	empty := &armcompute.VirtualMachine{}

	cases := map[string]struct {
		vm   *armcompute.VirtualMachine
		want string
	}{
		"windows by profile": {windowsByProfile, "azurerm_windows_virtual_machine"},
		"linux by profile":   {linuxByProfile, "azurerm_linux_virtual_machine"},
		"windows by disk":    {windowsByDisk, "azurerm_windows_virtual_machine"},
		"linux by disk":      {linuxByDisk, "azurerm_linux_virtual_machine"},
		"empty -> linux":     {empty, "azurerm_linux_virtual_machine"},
	}
	for name, c := range cases {
		if got := vmResourceType(c.vm); got != c.want {
			t.Errorf("%s: vmResourceType = %q, want %q", name, got, c.want)
		}
	}
}
