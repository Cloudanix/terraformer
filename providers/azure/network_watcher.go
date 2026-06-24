// Copyright 2021 The Terraformer Authors.
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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v7"
)

type NetworkWatcherGenerator struct {
	AzureService
}

// InitResources imports azurerm_network_watcher and its nested flow logs, packet
// captures and connection monitors. Migrated to the Track 2 armnetwork SDK.
func (g *NetworkWatcherGenerator) InitResources() error {
	subscriptionID, cred, opts := g.getClientOptions()
	if cred == nil {
		return nil
	}
	watchersClient, err := armnetwork.NewWatchersClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	flowLogsClient, err := armnetwork.NewFlowLogsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	capturesClient, err := armnetwork.NewPacketCapturesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	monitorsClient, err := armnetwork.NewConnectionMonitorsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}

	watchers, err := g.listWatchers(watchersClient)
	if err != nil {
		return err
	}
	for _, watcher := range watchers {
		watcherID := valueOrEmpty(watcher.ID)
		if watcherID == "" {
			continue
		}
		g.AppendSimpleResource(watcherID, valueOrEmpty(watcher.Name), "azurerm_network_watcher")
		parsed, err := ParseAzureResourceID(watcherID)
		if err != nil {
			return err
		}
		rg, watcherName := parsed.ResourceGroup, valueOrEmpty(watcher.Name)

		if err := appendFromPager(&g.AzureService, flowLogsClient.NewListPager(rg, watcherName, nil),
			func(p armnetwork.FlowLogsClientListResponse) []*armnetwork.FlowLog { return p.Value },
			func(i *armnetwork.FlowLog) string { return valueOrEmpty(i.ID) },
			func(i *armnetwork.FlowLog) string { return valueOrEmpty(i.Name) },
			"azurerm_network_watcher_flow_log"); err != nil {
			return err
		}
		if err := appendFromPager(&g.AzureService, capturesClient.NewListPager(rg, watcherName, nil),
			func(p armnetwork.PacketCapturesClientListResponse) []*armnetwork.PacketCaptureResult { return p.Value },
			func(i *armnetwork.PacketCaptureResult) string { return valueOrEmpty(i.ID) },
			func(i *armnetwork.PacketCaptureResult) string { return valueOrEmpty(i.Name) },
			"azurerm_network_packet_capture"); err != nil {
			return err
		}
		if err := appendFromPager(&g.AzureService, monitorsClient.NewListPager(rg, watcherName, nil),
			func(p armnetwork.ConnectionMonitorsClientListResponse) []*armnetwork.ConnectionMonitorResult {
				return p.Value
			},
			func(i *armnetwork.ConnectionMonitorResult) string { return valueOrEmpty(i.ID) },
			func(i *armnetwork.ConnectionMonitorResult) string { return valueOrEmpty(i.Name) },
			"azurerm_network_connection_monitor"); err != nil {
			return err
		}
	}
	return nil
}

func (g *NetworkWatcherGenerator) listWatchers(client *armnetwork.WatchersClient) ([]*armnetwork.Watcher, error) {
	var watchers []*armnetwork.Watcher
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		pager := client.NewListAllPager(nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return nil, err
			}
			watchers = append(watchers, page.Value...)
		}
		return watchers, nil
	}
	for _, rg := range rgs {
		pager := client.NewListPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return nil, err
			}
			watchers = append(watchers, page.Value...)
		}
	}
	return watchers, nil
}
