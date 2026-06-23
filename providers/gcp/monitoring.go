// Copyright 2018 The Terraformer Authors.
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

package gcp

import (
	"context"
	"log"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"

	"google.golang.org/api/iterator"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	monitoringpb "cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
)

var monitoringAllowEmptyValues = []string{}

var monitoringAdditionalFields = map[string]interface{}{}

type MonitoringGenerator struct {
	GCPService
}

func (g *MonitoringGenerator) loadAlerts(ctx context.Context, project string) error {
	client, err := monitoring.NewAlertPolicyClient(ctx)
	if err != nil {
		return err
	}

	req := &monitoringpb.ListAlertPoliciesRequest{
		Name: "projects/" + project,
	}

	alertIterator := client.ListAlertPolicies(ctx, req)

	for {
		alert, err := alertIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Println("error with alert:", err)
			continue
		}
		g.Resources = append(g.Resources, terraformutils.NewResource(
			alert.Name,
			alert.Name,
			"google_monitoring_alert_policy",
			g.ProviderName,
			map[string]string{
				"name":    alert.Name,
				"project": project,
			},
			monitoringAllowEmptyValues,
			monitoringAdditionalFields,
		))
	}
	return nil
}

func (g *MonitoringGenerator) loadGroups(ctx context.Context, project string) error {
	client, err := monitoring.NewGroupClient(ctx)
	if err != nil {
		return err
	}

	req := &monitoringpb.ListGroupsRequest{
		Name: "projects/" + project,
	}

	groupsIterator := client.ListGroups(ctx, req)
	for {
		group, err := groupsIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Println("error with group:", err)
			continue
		}
		g.Resources = append(g.Resources, terraformutils.NewResource(
			group.Name,
			group.Name,
			"google_monitoring_group",
			g.ProviderName,
			map[string]string{
				"name":    group.Name,
				"project": project,
			},
			monitoringAllowEmptyValues,
			monitoringAdditionalFields,
		))
	}
	return nil
}

func (g *MonitoringGenerator) loadNotificationChannel(ctx context.Context, project string) error {
	client, err := monitoring.NewNotificationChannelClient(ctx)
	if err != nil {
		return err
	}

	req := &monitoringpb.ListNotificationChannelsRequest{
		Name: "projects/" + project,
	}

	notificationChannelIterator := client.ListNotificationChannels(ctx, req)
	for {
		notificationChannel, err := notificationChannelIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Println("error with notification Channel:", err)
			continue
		}
		g.Resources = append(g.Resources, terraformutils.NewResource(
			notificationChannel.Name,
			notificationChannel.Name,
			"google_monitoring_notification_channel",
			g.ProviderName,
			map[string]string{
				"name":    notificationChannel.Name,
				"project": project,
			},
			monitoringAllowEmptyValues,
			monitoringAdditionalFields,
		))
	}
	return nil
}
func (g *MonitoringGenerator) loadUptimeCheck(ctx context.Context, project string) error {
	client, err := monitoring.NewUptimeCheckClient(ctx)
	if err != nil {
		return err
	}

	req := &monitoringpb.ListUptimeCheckConfigsRequest{
		Parent: "projects/" + project,
	}

	uptimeCheckConfigsIterator := client.ListUptimeCheckConfigs(ctx, req)
	for {
		uptimeCheckConfigs, err := uptimeCheckConfigsIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Println("error with uptimeCheckConfigs:", err)
			continue
		}
		g.Resources = append(g.Resources, terraformutils.NewResource(
			uptimeCheckConfigs.Name,
			uptimeCheckConfigs.Name,
			"google_monitoring_uptime_check_config",
			g.ProviderName,
			map[string]string{
				"name":    uptimeCheckConfigs.Name,
				"project": project,
			},
			monitoringAllowEmptyValues,
			monitoringAdditionalFields,
		))
	}
	return nil
}

// Generate TerraformResources from GCP API,
// from each alert  create 1 TerraformResource
// Need alert name as ID for terraform resource
func (g *MonitoringGenerator) InitResources() error {
	project := g.GetArgs()["project"].(string)
	ctx := context.Background()

	if err := g.loadAlerts(ctx, project); err != nil {
		return err
	}

	if err := g.loadGroups(ctx, project); err != nil {
		return err
	}

	if err := g.loadNotificationChannel(ctx, project); err != nil {
		return err
	}

	if err := g.loadUptimeCheck(ctx, project); err != nil {
		return err
	}

	g.loadServices(ctx, project)
	return nil
}

func (g *MonitoringGenerator) loadServices(ctx context.Context, project string) {
	client, err := monitoring.NewServiceMonitoringClient(ctx)
	if err != nil {
		log.Println(err)
		return
	}
	it := client.ListServices(ctx, &monitoringpb.ListServicesRequest{Parent: "projects/" + project})
	for {
		svc, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Println("error with monitoring service:", err)
			break
		}
		t := strings.Split(svc.Name, "/")
		name := t[len(t)-1]
		g.Resources = append(g.Resources, terraformutils.NewResource(
			svc.Name,
			name,
			"google_monitoring_service",
			g.ProviderName,
			map[string]string{
				"service_id": name,
				"project":    project,
			},
			monitoringAllowEmptyValues,
			monitoringAdditionalFields,
		))
		sloIt := client.ListServiceLevelObjectives(ctx, &monitoringpb.ListServiceLevelObjectivesRequest{Parent: svc.Name})
		for {
			slo, serr := sloIt.Next()
			if serr == iterator.Done {
				break
			}
			if serr != nil {
				log.Println("error with slo:", serr)
				break
			}
			st := strings.Split(slo.Name, "/")
			sloName := st[len(st)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				slo.Name, sloName, "google_monitoring_slo", g.ProviderName,
				map[string]string{"slo_id": sloName, "service": name, "project": project},
				monitoringAllowEmptyValues, monitoringAdditionalFields))
		}
	}
}
