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
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

const (
	containerIDFormat = "https://%s.blob.core.windows.net/%s"
)

type StorageContainerGenerator struct {
	AzureService
}

// ListBlobContainers enumerates azurerm_storage_container across the storage
// accounts in scope. Migrated to the Track 2 armstorage SDK.
func (g *StorageContainerGenerator) ListBlobContainers() ([]terraformutils.Resource, error) {
	var containerResources []terraformutils.Resource
	subscriptionID, cred, opts := g.getClientOptions()
	if cred == nil {
		return containerResources, nil
	}
	containersClient, err := armstorage.NewBlobContainersClient(subscriptionID, cred, opts)
	if err != nil {
		return nil, err
	}

	accounts, err := g.getStorageAccounts()
	if err != nil {
		return containerResources, err
	}

	for _, storageAccount := range accounts {
		accountID := valueOrEmpty(storageAccount.ID)
		if accountID == "" {
			continue
		}
		parsed, err := ParseAzureResourceID(accountID)
		if err != nil {
			break
		}
		accountName := valueOrEmpty(storageAccount.Name)
		pager := containersClient.NewListPager(parsed.ResourceGroup, accountName, nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return containerResources, err
			}
			for _, containerItem := range page.Value {
				if containerItem == nil {
					continue
				}
				name := valueOrEmpty(containerItem.Name)
				containerResources = append(containerResources, terraformutils.NewResource(
					fmt.Sprintf(containerIDFormat, accountName, name),
					name,
					"azurerm_storage_container",
					"azurerm",
					map[string]string{
						"storage_account_name": accountName,
						"name":                 name,
					},
					[]string{},
					map[string]interface{}{}))
			}
		}
	}

	return containerResources, nil
}

func (g *StorageContainerGenerator) getStorageAccounts() ([]*armstorage.Account, error) {
	subscriptionID, cred, opts := g.getClientOptions()
	client, err := armstorage.NewAccountsClient(subscriptionID, cred, opts)
	if err != nil {
		return nil, err
	}
	var accounts []*armstorage.Account
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return nil, err
			}
			accounts = append(accounts, page.Value...)
		}
		return accounts, nil
	}
	for _, rg := range rgs {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return nil, err
			}
			accounts = append(accounts, page.Value...)
		}
	}
	return accounts, nil
}

func (g *StorageContainerGenerator) InitResources() error {
	resources, err := g.ListBlobContainers()
	if err != nil {
		return err
	}
	g.Resources = resources
	return nil
}
