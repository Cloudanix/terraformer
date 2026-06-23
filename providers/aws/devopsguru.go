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

	"github.com/aws/aws-sdk-go-v2/service/devopsguru"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type DevOpsGuruGenerator struct {
	AWSService
}

// InitResources enumerates DevOps Guru notification channels plus the account
// singletons (service integration, event sources config), keyed by account id.
func (g *DevOpsGuruGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := devopsguru.NewFromConfig(config)
	ctx := context.TODO()

	for p := devopsguru.NewListNotificationChannelsPaginator(svc, &devopsguru.ListNotificationChannelsInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, c := range page.Channels {
			id := StringValue(c.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_devopsguru_notification_channel", "aws", defaultAllowEmptyValues))
		}
	}

	account, err := g.getAccountNumber(config)
	if err != nil {
		return err
	}
	accountID := StringValue(account)
	if accountID == "" {
		return nil
	}
	if _, err := svc.DescribeServiceIntegration(ctx, &devopsguru.DescribeServiceIntegrationInput{}); err == nil {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			accountID, accountID, "aws_devopsguru_service_integration", "aws", defaultAllowEmptyValues))
	}
	if _, err := svc.DescribeEventSourcesConfig(ctx, &devopsguru.DescribeEventSourcesConfigInput{}); err == nil {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			accountID, accountID, "aws_devopsguru_event_sources_config", "aws", defaultAllowEmptyValues))
	}
	return nil
}
