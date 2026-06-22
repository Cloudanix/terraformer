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

	"github.com/aws/aws-sdk-go-v2/service/chatbot"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type ChatbotGenerator struct {
	AWSService
}

// InitResources enumerates Chatbot Slack and Microsoft Teams channel
// configurations. Import ID is the chat configuration ARN.
func (g *ChatbotGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := chatbot.NewFromConfig(config)
	ctx := context.TODO()

	slack := chatbot.NewDescribeSlackChannelConfigurationsPaginator(svc, &chatbot.DescribeSlackChannelConfigurationsInput{})
	for slack.HasMorePages() {
		page, err := slack.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, c := range page.SlackChannelConfigurations {
			arn := StringValue(c.ChatConfigurationArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(c.SlackChannelName), "aws_chatbot_slack_channel_configuration", "aws", defaultAllowEmptyValues))
		}
	}

	teams := chatbot.NewListMicrosoftTeamsChannelConfigurationsPaginator(svc, &chatbot.ListMicrosoftTeamsChannelConfigurationsInput{})
	for teams.HasMorePages() {
		page, err := teams.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, c := range page.TeamChannelConfigurations {
			arn := StringValue(c.ChatConfigurationArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_chatbot_teams_channel_configuration", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
