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

	"github.com/aws/aws-sdk-go-v2/service/lexmodelsv2"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type LexModelsV2Generator struct {
	AWSService
}

// InitResources enumerates Lex v2 bots. Import ID is the bot id.
func (g *LexModelsV2Generator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := lexmodelsv2.NewFromConfig(config)

	p := lexmodelsv2.NewListBotsPaginator(svc, &lexmodelsv2.ListBotsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, bot := range page.BotSummaries {
			id := StringValue(bot.BotId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(bot.BotName), "aws_lexv2models_bot", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
