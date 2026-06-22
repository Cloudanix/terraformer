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

	"github.com/aws/aws-sdk-go-v2/aws"
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
			g.loadBotChildren(svc, id)
		}
	}
	return nil
}

// loadBotChildren enumerates a bot's versions plus its DRAFT-version locales,
// intents, slot types, and slots. Import IDs are comma-separated component lists.
func (g *LexModelsV2Generator) loadBotChildren(svc *lexmodelsv2.Client, botID string) {
	ctx := context.TODO()
	const draft = "DRAFT"

	for vp := lexmodelsv2.NewListBotVersionsPaginator(svc, &lexmodelsv2.ListBotVersionsInput{BotId: aws.String(botID)}); vp.HasMorePages(); {
		page, err := vp.NextPage(ctx)
		if err != nil {
			break
		}
		for _, v := range page.BotVersionSummaries {
			version := StringValue(v.BotVersion)
			if version == "" || version == draft {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				botID+","+version, botID+"_"+version, "aws_lexv2models_bot_version", "aws", defaultAllowEmptyValues))
		}
	}

	for lp := lexmodelsv2.NewListBotLocalesPaginator(svc, &lexmodelsv2.ListBotLocalesInput{BotId: aws.String(botID), BotVersion: aws.String(draft)}); lp.HasMorePages(); {
		page, err := lp.NextPage(ctx)
		if err != nil {
			break
		}
		for _, l := range page.BotLocaleSummaries {
			locale := StringValue(l.LocaleId)
			if locale == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				botID+","+draft+","+locale, botID+"_"+locale, "aws_lexv2models_bot_locale", "aws", defaultAllowEmptyValues))

			for ip := lexmodelsv2.NewListIntentsPaginator(svc, &lexmodelsv2.ListIntentsInput{BotId: aws.String(botID), BotVersion: aws.String(draft), LocaleId: aws.String(locale)}); ip.HasMorePages(); {
				ipage, err := ip.NextPage(ctx)
				if err != nil {
					break
				}
				for _, intent := range ipage.IntentSummaries {
					intentID := StringValue(intent.IntentId)
					if intentID == "" {
						continue
					}
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						intentID+","+botID+","+draft+","+locale, botID+"_"+locale+"_"+intentID,
						"aws_lexv2models_intent", "aws", defaultAllowEmptyValues))

					for sp := lexmodelsv2.NewListSlotsPaginator(svc, &lexmodelsv2.ListSlotsInput{BotId: aws.String(botID), BotVersion: aws.String(draft), LocaleId: aws.String(locale), IntentId: aws.String(intentID)}); sp.HasMorePages(); {
						spage, err := sp.NextPage(ctx)
						if err != nil {
							break
						}
						for _, slot := range spage.SlotSummaries {
							slotID := StringValue(slot.SlotId)
							if slotID == "" {
								continue
							}
							g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
								slotID+","+botID+","+draft+","+intentID+","+locale, botID+"_"+intentID+"_"+slotID,
								"aws_lexv2models_slot", "aws", defaultAllowEmptyValues))
						}
					}
				}
			}

			for stp := lexmodelsv2.NewListSlotTypesPaginator(svc, &lexmodelsv2.ListSlotTypesInput{BotId: aws.String(botID), BotVersion: aws.String(draft), LocaleId: aws.String(locale)}); stp.HasMorePages(); {
				stpage, err := stp.NextPage(ctx)
				if err != nil {
					break
				}
				for _, st := range stpage.SlotTypeSummaries {
					stID := StringValue(st.SlotTypeId)
					if stID == "" {
						continue
					}
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						stID+","+botID+","+draft+","+locale, botID+"_"+locale+"_"+stID,
						"aws_lexv2models_slot_type", "aws", defaultAllowEmptyValues))
				}
			}
		}
	}
}
