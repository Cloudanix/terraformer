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

package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/notifications"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type NotificationsGenerator struct {
	AWSService
}

// InitResources enumerates User Notifications hubs (import by region) and
// notification configurations (import by ARN).
func (g *NotificationsGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := notifications.NewFromConfig(config)
	for p := notifications.NewListNotificationHubsPaginator(svc, &notifications.ListNotificationHubsInput{}); p.HasMorePages(); {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, h := range page.NotificationHubs {
			region := StringValue(h.NotificationHubRegion)
			if region == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				region, region, "aws_notifications_notification_hub", "aws", defaultAllowEmptyValues))
		}
	}
	for p := notifications.NewListNotificationConfigurationsPaginator(svc, &notifications.ListNotificationConfigurationsInput{}); p.HasMorePages(); {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, c := range page.NotificationConfigurations {
			arn := StringValue(c.Arn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_notifications_notification_configuration", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
