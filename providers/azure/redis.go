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
	"log"

	"github.com/Azure/azure-sdk-for-go/services/redis/mgmt/2018-03-01/redis"
	"github.com/Azure/go-autorest/autorest"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/hashicorp/go-azure-helpers/authentication"
)

type RedisGenerator struct {
	AzureService
}

func (g *RedisGenerator) listRedisServers() ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	subscriptionID := g.Args["config"].(authentication.Config).SubscriptionID
	resourceManagerEndpoint := g.Args["config"].(authentication.Config).CustomResourceManagerEndpoint
	RedisClient := redis.NewClientWithBaseURI(resourceManagerEndpoint, subscriptionID)

	redisServersIterator, err := RedisClient.ListComplete(ctx)
	if err != nil {
		return nil, err
	}

	for redisServersIterator.NotDone() {
		redisServer := redisServersIterator.Value()
		resources = append(resources, terraformutils.NewSimpleResource(
			*redisServer.ID,
			*redisServer.Name,
			"azurerm_redis_cache",
			g.ProviderName,
			[]string{}))

		id, err := ParseAzureResourceID(*redisServer.ID)
		if err != nil {
			return nil, err
		}
		firewallRules, err := g.listRedisFirewallRules(id.ResourceGroup, *redisServer.Name)
		if err != nil {
			return nil, err
		}
		resources = append(resources, firewallRules...)

		linkedServers, err := g.listRedisLinkedServers(id.ResourceGroup, *redisServer.Name)
		if err != nil {
			return nil, err
		}
		resources = append(resources, linkedServers...)

		if err := redisServersIterator.Next(); err != nil {
			log.Println(err)
			break
		}
	}

	return resources, nil
}

// listRedisFirewallRules enumerates azurerm_redis_firewall_rule for one cache.
func (g *RedisGenerator) listRedisFirewallRules(resourceGroupName string, cacheName string) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	subscriptionID := g.Args["config"].(authentication.Config).SubscriptionID
	resourceManagerEndpoint := g.Args["config"].(authentication.Config).CustomResourceManagerEndpoint
	client := redis.NewFirewallRulesClientWithBaseURI(resourceManagerEndpoint, subscriptionID)
	client.Authorizer = g.Args["authorizer"].(autorest.Authorizer)

	iterator, err := client.ListByRedisResourceComplete(ctx, resourceGroupName, cacheName)
	if err != nil {
		return nil, err
	}
	for iterator.NotDone() {
		rule := iterator.Value()
		resources = append(resources, terraformutils.NewSimpleResource(
			*rule.ID,
			*rule.Name,
			"azurerm_redis_firewall_rule",
			g.ProviderName,
			[]string{}))
		if err := iterator.NextWithContext(ctx); err != nil {
			log.Println(err)
			break
		}
	}

	return resources, nil
}

// listRedisLinkedServers enumerates azurerm_redis_linked_server for one cache.
func (g *RedisGenerator) listRedisLinkedServers(resourceGroupName string, cacheName string) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	subscriptionID := g.Args["config"].(authentication.Config).SubscriptionID
	resourceManagerEndpoint := g.Args["config"].(authentication.Config).CustomResourceManagerEndpoint
	client := redis.NewLinkedServerClientWithBaseURI(resourceManagerEndpoint, subscriptionID)
	client.Authorizer = g.Args["authorizer"].(autorest.Authorizer)

	iterator, err := client.ListComplete(ctx, resourceGroupName, cacheName)
	if err != nil {
		return nil, err
	}
	for iterator.NotDone() {
		linkedServer := iterator.Value()
		resources = append(resources, terraformutils.NewSimpleResource(
			*linkedServer.ID,
			*linkedServer.Name,
			"azurerm_redis_linked_server",
			g.ProviderName,
			[]string{}))
		if err := iterator.NextWithContext(ctx); err != nil {
			log.Println(err)
			break
		}
	}

	return resources, nil
}

func (g *RedisGenerator) InitResources() error {
	functions := []func() ([]terraformutils.Resource, error){
		g.listRedisServers,
	}

	for _, f := range functions {
		resources, err := f()
		if err != nil {
			return err
		}
		g.Resources = append(g.Resources, resources...)
	}

	return nil
}
