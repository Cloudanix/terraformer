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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/maintenance/armmaintenance"
)

// MaintenanceGenerator imports azurerm_maintenance_configuration. The
// Configurations API lists subscription-wide only (no per-RG list).
type MaintenanceGenerator struct {
	AzureService
}

func (g *MaintenanceGenerator) InitResources() error {
	subscriptionID, cred, opts := g.getClientOptions()
	if cred == nil {
		return nil
	}
	client, err := armmaintenance.NewConfigurationsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	return appendFromPager(&g.AzureService, client.NewListPager(nil),
		func(p armmaintenance.ConfigurationsClientListResponse) []*armmaintenance.Configuration {
			return p.Value
		},
		func(i *armmaintenance.Configuration) string { return valueOrEmpty(i.ID) },
		func(i *armmaintenance.Configuration) string { return valueOrEmpty(i.Name) },
		"azurerm_maintenance_configuration")
}
