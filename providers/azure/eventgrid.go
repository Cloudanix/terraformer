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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventgrid/armeventgrid"
)

type EventGridGenerator struct {
	AzureService
}

func (g *EventGridGenerator) InitResources() error {
	if _, cred, _ := g.getClientOptions(); cred == nil {
		return nil
	}
	if err := g.initTopics(); err != nil {
		return err
	}
	return g.initDomains()
}

func (g *EventGridGenerator) initTopics() error {
	subscriptionID, cred, opts := g.getClientOptions()
	client, err := armeventgrid.NewTopicsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	id := func(i *armeventgrid.Topic) string { return valueOrEmpty(i.ID) }
	name := func(i *armeventgrid.Topic) string { return valueOrEmpty(i.Name) }
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		return appendFromPager(&g.AzureService, client.NewListBySubscriptionPager(nil),
			func(p armeventgrid.TopicsClientListBySubscriptionResponse) []*armeventgrid.Topic { return p.Value },
			id, name, "azurerm_eventgrid_topic")
	}
	for _, rg := range rgs {
		if err := appendFromPager(&g.AzureService, client.NewListByResourceGroupPager(rg, nil),
			func(p armeventgrid.TopicsClientListByResourceGroupResponse) []*armeventgrid.Topic { return p.Value },
			id, name, "azurerm_eventgrid_topic"); err != nil {
			return err
		}
	}
	return nil
}

func (g *EventGridGenerator) initDomains() error {
	subscriptionID, cred, opts := g.getClientOptions()
	client, err := armeventgrid.NewDomainsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	id := func(i *armeventgrid.Domain) string { return valueOrEmpty(i.ID) }
	name := func(i *armeventgrid.Domain) string { return valueOrEmpty(i.Name) }
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		return appendFromPager(&g.AzureService, client.NewListBySubscriptionPager(nil),
			func(p armeventgrid.DomainsClientListBySubscriptionResponse) []*armeventgrid.Domain { return p.Value },
			id, name, "azurerm_eventgrid_domain")
	}
	for _, rg := range rgs {
		if err := appendFromPager(&g.AzureService, client.NewListByResourceGroupPager(rg, nil),
			func(p armeventgrid.DomainsClientListByResourceGroupResponse) []*armeventgrid.Domain { return p.Value },
			id, name, "azurerm_eventgrid_domain"); err != nil {
			return err
		}
	}
	return nil
}
