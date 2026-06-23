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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventhub/armeventhub"
)

type EventHubGenerator struct {
	AzureService
}

// InitResources imports eventhub namespaces, hubs, consumer groups,
// authorization rules (namespace + hub) and geo-DR configs. Migrated to the
// Track 2 armeventhub SDK (was Track 1 services/eventhub).
func (g *EventHubGenerator) InitResources() error {
	subscriptionID, cred, opts := g.getClientOptions()
	if cred == nil {
		return nil
	}
	nsClient, err := armeventhub.NewNamespacesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	hubsClient, err := armeventhub.NewEventHubsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	cgClient, err := armeventhub.NewConsumerGroupsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	drClient, err := armeventhub.NewDisasterRecoveryConfigsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}

	namespaces, err := g.listNamespaces(nsClient)
	if err != nil {
		return err
	}
	for _, ns := range namespaces {
		nsID := valueOrEmpty(ns.ID)
		if nsID == "" {
			continue
		}
		g.AppendSimpleResource(nsID, valueOrEmpty(ns.Name), "azurerm_eventhub_namespace")
		parsed, err := ParseAzureResourceID(nsID)
		if err != nil {
			return err
		}
		rg, nsName := parsed.ResourceGroup, valueOrEmpty(ns.Name)

		if err := appendFromPager(&g.AzureService, nsClient.NewListAuthorizationRulesPager(rg, nsName, nil),
			func(p armeventhub.NamespacesClientListAuthorizationRulesResponse) []*armeventhub.AuthorizationRule {
				return p.Value
			},
			func(i *armeventhub.AuthorizationRule) string { return valueOrEmpty(i.ID) },
			func(i *armeventhub.AuthorizationRule) string { return valueOrEmpty(i.Name) },
			"azurerm_eventhub_namespace_authorization_rule"); err != nil {
			return err
		}
		if err := appendFromPager(&g.AzureService, drClient.NewListPager(rg, nsName, nil),
			func(p armeventhub.DisasterRecoveryConfigsClientListResponse) []*armeventhub.ArmDisasterRecovery {
				return p.Value
			},
			func(i *armeventhub.ArmDisasterRecovery) string { return valueOrEmpty(i.ID) },
			func(i *armeventhub.ArmDisasterRecovery) string { return valueOrEmpty(i.Name) },
			"azurerm_eventhub_namespace_disaster_recovery_config"); err != nil {
			return err
		}

		if err := g.appendEventHubs(hubsClient, cgClient, rg, nsName); err != nil {
			return err
		}
	}
	return nil
}

func (g *EventHubGenerator) appendEventHubs(hubsClient *armeventhub.EventHubsClient, cgClient *armeventhub.ConsumerGroupsClient, rg, nsName string) error {
	pager := hubsClient.NewListByNamespacePager(rg, nsName, nil)
	for pager.More() {
		page, err := pager.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, hub := range page.Value {
			if hub == nil {
				continue
			}
			hubID := valueOrEmpty(hub.ID)
			if hubID == "" {
				continue
			}
			hubName := valueOrEmpty(hub.Name)
			g.AppendSimpleResource(hubID, hubName, "azurerm_eventhub")

			if err := appendFromPager(&g.AzureService, cgClient.NewListByEventHubPager(rg, nsName, hubName, nil),
				func(p armeventhub.ConsumerGroupsClientListByEventHubResponse) []*armeventhub.ConsumerGroup {
					return p.Value
				},
				func(i *armeventhub.ConsumerGroup) string { return valueOrEmpty(i.ID) },
				func(i *armeventhub.ConsumerGroup) string { return valueOrEmpty(i.Name) },
				"azurerm_eventhub_consumer_group"); err != nil {
				return err
			}
			if err := appendFromPager(&g.AzureService, hubsClient.NewListAuthorizationRulesPager(rg, nsName, hubName, nil),
				func(p armeventhub.EventHubsClientListAuthorizationRulesResponse) []*armeventhub.AuthorizationRule {
					return p.Value
				},
				func(i *armeventhub.AuthorizationRule) string { return valueOrEmpty(i.ID) },
				func(i *armeventhub.AuthorizationRule) string { return valueOrEmpty(i.Name) },
				"azurerm_eventhub_authorization_rule"); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *EventHubGenerator) listNamespaces(client *armeventhub.NamespacesClient) ([]*armeventhub.EHNamespace, error) {
	var namespaces []*armeventhub.EHNamespace
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return nil, err
			}
			namespaces = append(namespaces, page.Value...)
		}
		return namespaces, nil
	}
	for _, rg := range rgs {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return nil, err
			}
			namespaces = append(namespaces, page.Value...)
		}
	}
	return namespaces, nil
}
