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

	armci "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerinstance/armcontainerinstance/v2"
	armcr "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry/v2"
)

type ContainerGenerator struct {
	AzureService
}

// InitResources imports azurerm_container_group and azurerm_container_registry
// (+ webhook/replication/task). Migrated to the Track 2 armcontainerinstance and
// armcontainerregistry SDKs.
func (g *ContainerGenerator) InitResources() error {
	if _, cred, _ := g.getClientOptions(); cred == nil {
		return nil
	}
	if err := g.initContainerGroups(); err != nil {
		return err
	}
	return g.initRegistries()
}

func (g *ContainerGenerator) initContainerGroups() error {
	subscriptionID, cred, opts := g.getClientOptions()
	client, err := armci.NewContainerGroupsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	id := func(i *armci.ContainerGroup) string { return valueOrEmpty(i.ID) }
	name := func(i *armci.ContainerGroup) string { return valueOrEmpty(i.Name) }
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		return appendFromPager(&g.AzureService, client.NewListPager(nil),
			func(p armci.ContainerGroupsClientListResponse) []*armci.ContainerGroup { return p.Value },
			id, name, "azurerm_container_group")
	}
	for _, rg := range rgs {
		if err := appendFromPager(&g.AzureService, client.NewListByResourceGroupPager(rg, nil),
			func(p armci.ContainerGroupsClientListByResourceGroupResponse) []*armci.ContainerGroup {
				return p.Value
			},
			id, name, "azurerm_container_group"); err != nil {
			return err
		}
	}
	return nil
}

func (g *ContainerGenerator) initRegistries() error {
	subscriptionID, cred, opts := g.getClientOptions()
	regClient, err := armcr.NewRegistriesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	webhookClient, err := armcr.NewWebhooksClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	replClient, err := armcr.NewReplicationsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	taskClient, err := armcr.NewTasksClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}

	registries, err := g.listRegistries(regClient)
	if err != nil {
		return err
	}
	for _, reg := range registries {
		regID := valueOrEmpty(reg.ID)
		if regID == "" {
			continue
		}
		g.AppendSimpleResource(regID, valueOrEmpty(reg.Name), "azurerm_container_registry")
		parsed, err := ParseAzureResourceID(regID)
		if err != nil {
			return err
		}
		rg, regName := parsed.ResourceGroup, valueOrEmpty(reg.Name)

		if err := appendFromPager(&g.AzureService, webhookClient.NewListPager(rg, regName, nil),
			func(p armcr.WebhooksClientListResponse) []*armcr.Webhook { return p.Value },
			func(i *armcr.Webhook) string { return valueOrEmpty(i.ID) },
			func(i *armcr.Webhook) string { return valueOrEmpty(i.Name) },
			"azurerm_container_registry_webhook"); err != nil {
			return err
		}
		if err := appendFromPager(&g.AzureService, replClient.NewListPager(rg, regName, nil),
			func(p armcr.ReplicationsClientListResponse) []*armcr.Replication { return p.Value },
			func(i *armcr.Replication) string { return valueOrEmpty(i.ID) },
			func(i *armcr.Replication) string { return valueOrEmpty(i.Name) },
			"azurerm_container_registry_replication"); err != nil {
			return err
		}
		if err := appendFromPager(&g.AzureService, taskClient.NewListPager(rg, regName, nil),
			func(p armcr.TasksClientListResponse) []*armcr.Task { return p.Value },
			func(i *armcr.Task) string { return valueOrEmpty(i.ID) },
			func(i *armcr.Task) string { return valueOrEmpty(i.Name) },
			"azurerm_container_registry_task"); err != nil {
			return err
		}
	}
	return nil
}

func (g *ContainerGenerator) listRegistries(client *armcr.RegistriesClient) ([]*armcr.Registry, error) {
	var registries []*armcr.Registry
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return nil, err
			}
			registries = append(registries, page.Value...)
		}
		return registries, nil
	}
	for _, rg := range rgs {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return nil, err
			}
			registries = append(registries, page.Value...)
		}
	}
	return registries, nil
}
