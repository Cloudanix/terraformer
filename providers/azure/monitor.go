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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
)

type MonitorGenerator struct {
	AzureService
}

func (g *MonitorGenerator) InitResources() error {
	if _, cred, _ := g.getClientOptions(); cred == nil {
		return nil
	}
	for _, fn := range []func() error{
		g.initActionGroups,
		g.initActivityLogAlerts,
		g.initAutoscaleSettings,
		g.initMetricAlerts,
	} {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}

// initActionGroups — azurerm_monitor_action_group (resource-group scoped).
func (g *MonitorGenerator) initActionGroups() error {
	subscriptionID, cred, opts := g.getClientOptions()
	client, err := armmonitor.NewActionGroupsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	for _, rg := range g.resourceGroups() {
		if err := appendFromPager(&g.AzureService, client.NewListByResourceGroupPager(rg, nil),
			func(p armmonitor.ActionGroupsClientListByResourceGroupResponse) []*armmonitor.ActionGroupResource {
				return p.Value
			},
			func(i *armmonitor.ActionGroupResource) string { return valueOrEmpty(i.ID) },
			func(i *armmonitor.ActionGroupResource) string { return valueOrEmpty(i.Name) },
			"azurerm_monitor_action_group"); err != nil {
			return err
		}
	}
	return nil
}

// initActivityLogAlerts — azurerm_monitor_activity_log_alert (resource-group scoped).
func (g *MonitorGenerator) initActivityLogAlerts() error {
	subscriptionID, cred, opts := g.getClientOptions()
	client, err := armmonitor.NewActivityLogAlertsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	for _, rg := range g.resourceGroups() {
		if err := appendFromPager(&g.AzureService, client.NewListByResourceGroupPager(rg, nil),
			func(p armmonitor.ActivityLogAlertsClientListByResourceGroupResponse) []*armmonitor.ActivityLogAlertResource {
				return p.Value
			},
			func(i *armmonitor.ActivityLogAlertResource) string { return valueOrEmpty(i.ID) },
			func(i *armmonitor.ActivityLogAlertResource) string { return valueOrEmpty(i.Name) },
			"azurerm_monitor_activity_log_alert"); err != nil {
			return err
		}
	}
	return nil
}

// initAutoscaleSettings — azurerm_monitor_autoscale_setting.
func (g *MonitorGenerator) initAutoscaleSettings() error {
	subscriptionID, cred, opts := g.getClientOptions()
	client, err := armmonitor.NewAutoscaleSettingsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	id := func(i *armmonitor.AutoscaleSettingResource) string { return valueOrEmpty(i.ID) }
	name := func(i *armmonitor.AutoscaleSettingResource) string { return valueOrEmpty(i.Name) }
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		return appendFromPager(&g.AzureService, client.NewListBySubscriptionPager(nil),
			func(p armmonitor.AutoscaleSettingsClientListBySubscriptionResponse) []*armmonitor.AutoscaleSettingResource {
				return p.Value
			},
			id, name, "azurerm_monitor_autoscale_setting")
	}
	for _, rg := range rgs {
		if err := appendFromPager(&g.AzureService, client.NewListByResourceGroupPager(rg, nil),
			func(p armmonitor.AutoscaleSettingsClientListByResourceGroupResponse) []*armmonitor.AutoscaleSettingResource {
				return p.Value
			},
			id, name, "azurerm_monitor_autoscale_setting"); err != nil {
			return err
		}
	}
	return nil
}

// initMetricAlerts — azurerm_monitor_metric_alert.
func (g *MonitorGenerator) initMetricAlerts() error {
	subscriptionID, cred, opts := g.getClientOptions()
	client, err := armmonitor.NewMetricAlertsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	id := func(i *armmonitor.MetricAlertResource) string { return valueOrEmpty(i.ID) }
	name := func(i *armmonitor.MetricAlertResource) string { return valueOrEmpty(i.Name) }
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		return appendFromPager(&g.AzureService, client.NewListBySubscriptionPager(nil),
			func(p armmonitor.MetricAlertsClientListBySubscriptionResponse) []*armmonitor.MetricAlertResource {
				return p.Value
			},
			id, name, "azurerm_monitor_metric_alert")
	}
	for _, rg := range rgs {
		if err := appendFromPager(&g.AzureService, client.NewListByResourceGroupPager(rg, nil),
			func(p armmonitor.MetricAlertsClientListByResourceGroupResponse) []*armmonitor.MetricAlertResource {
				return p.Value
			},
			id, name, "azurerm_monitor_metric_alert"); err != nil {
			return err
		}
	}
	return nil
}
