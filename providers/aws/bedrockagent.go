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
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type BedrockAgentGenerator struct {
	AWSService
}

func (g *BedrockAgentGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := bedrockagent.NewFromConfig(config)
	ctx := context.TODO()
	var agentIDs []string
	p := bedrockagent.NewListAgentsPaginator(svc, &bedrockagent.ListAgentsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, item := range page.AgentSummaries {
			id := StringValue(item.AgentId)
			if id == "" {
				continue
			}
			agentIDs = append(agentIDs, id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(item.AgentName), "aws_bedrockagent_agent", "aws", defaultAllowEmptyValues))
		}
	}

	for _, agentID := range agentIDs {
		agent := agentID
		for ap := bedrockagent.NewListAgentAliasesPaginator(svc, &bedrockagent.ListAgentAliasesInput{AgentId: &agent}); ap.HasMorePages(); {
			page, err := ap.NextPage(ctx)
			if err != nil {
				break
			}
			for _, a := range page.AgentAliasSummaries {
				id := StringValue(a.AgentAliasId)
				if id == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id+","+agent, agent+"_"+id, "aws_bedrockagent_agent_alias", "aws", defaultAllowEmptyValues))
			}
		}
		for agp := bedrockagent.NewListAgentActionGroupsPaginator(svc, &bedrockagent.ListAgentActionGroupsInput{AgentId: &agent, AgentVersion: aws.String("DRAFT")}); agp.HasMorePages(); {
			page, err := agp.NextPage(ctx)
			if err != nil {
				break
			}
			for _, ag := range page.ActionGroupSummaries {
				id := StringValue(ag.ActionGroupId)
				if id == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id+","+agent+",DRAFT", agent+"_"+id, "aws_bedrockagent_agent_action_group", "aws", defaultAllowEmptyValues))
			}
		}
		for kbp := bedrockagent.NewListAgentKnowledgeBasesPaginator(svc, &bedrockagent.ListAgentKnowledgeBasesInput{AgentId: &agent, AgentVersion: aws.String("DRAFT")}); kbp.HasMorePages(); {
			page, err := kbp.NextPage(ctx)
			if err != nil {
				break
			}
			for _, kb := range page.AgentKnowledgeBaseSummaries {
				kbID := StringValue(kb.KnowledgeBaseId)
				if kbID == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					agent+","+kbID+",DRAFT", agent+"_"+kbID, "aws_bedrockagent_agent_knowledge_base_association", "aws", defaultAllowEmptyValues))
			}
		}
	}

	for kp := bedrockagent.NewListKnowledgeBasesPaginator(svc, &bedrockagent.ListKnowledgeBasesInput{}); kp.HasMorePages(); {
		page, err := kp.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, kb := range page.KnowledgeBaseSummaries {
			kbID := StringValue(kb.KnowledgeBaseId)
			if kbID == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				kbID, kbID, "aws_bedrockagent_knowledge_base", "aws", defaultAllowEmptyValues))
			for dp := bedrockagent.NewListDataSourcesPaginator(svc, &bedrockagent.ListDataSourcesInput{KnowledgeBaseId: aws.String(kbID)}); dp.HasMorePages(); {
				dpage, err := dp.NextPage(ctx)
				if err != nil {
					break
				}
				for _, ds := range dpage.DataSourceSummaries {
					id := StringValue(ds.DataSourceId)
					if id == "" {
						continue
					}
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id+","+kbID, kbID+"_"+id, "aws_bedrockagent_data_source", "aws", defaultAllowEmptyValues))
				}
			}
		}
	}
	return nil
}
