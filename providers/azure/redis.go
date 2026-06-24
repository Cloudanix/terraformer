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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
)

type RedisGenerator struct {
	AzureService
}

// InitResources imports azurerm_redis_cache and its firewall rules and linked
// servers. Migrated to the Track 2 armredis SDK (was Track 1 services/redis).
func (g *RedisGenerator) InitResources() error {
	subscriptionID, cred, opts := g.getClientOptions()
	if cred == nil {
		return nil
	}
	cacheClient, err := armredis.NewClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	firewallClient, err := armredis.NewFirewallRulesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	linkedClient, err := armredis.NewLinkedServerClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}

	caches, err := g.listCaches(cacheClient)
	if err != nil {
		return err
	}
	for _, cache := range caches {
		cacheID := valueOrEmpty(cache.ID)
		if cacheID == "" {
			continue
		}
		g.AppendSimpleResource(cacheID, valueOrEmpty(cache.Name), "azurerm_redis_cache")
		parsed, err := ParseAzureResourceID(cacheID)
		if err != nil {
			return err
		}
		rg, cacheName := parsed.ResourceGroup, valueOrEmpty(cache.Name)

		if err := appendFromPager(&g.AzureService, firewallClient.NewListPager(rg, cacheName, nil),
			func(p armredis.FirewallRulesClientListResponse) []*armredis.FirewallRule { return p.Value },
			func(i *armredis.FirewallRule) string { return valueOrEmpty(i.ID) },
			func(i *armredis.FirewallRule) string { return valueOrEmpty(i.Name) },
			"azurerm_redis_firewall_rule"); err != nil {
			return err
		}
		if err := appendFromPager(&g.AzureService, linkedClient.NewListPager(rg, cacheName, nil),
			func(p armredis.LinkedServerClientListResponse) []*armredis.LinkedServerWithProperties { return p.Value },
			func(i *armredis.LinkedServerWithProperties) string { return valueOrEmpty(i.ID) },
			func(i *armredis.LinkedServerWithProperties) string { return valueOrEmpty(i.Name) },
			"azurerm_redis_linked_server"); err != nil {
			return err
		}
	}
	return nil
}

func (g *RedisGenerator) listCaches(client *armredis.Client) ([]*armredis.ResourceInfo, error) {
	var caches []*armredis.ResourceInfo
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		pager := client.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return nil, err
			}
			caches = append(caches, page.Value...)
		}
		return caches, nil
	}
	for _, rg := range rgs {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return nil, err
			}
			caches = append(caches, page.Value...)
		}
	}
	return caches, nil
}
