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

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/dialogflow/v3"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var dialogflowAllowEmptyValues = []string{""}

var dialogflowAdditionalFields = map[string]interface{}{}

type DialogflowGenerator struct {
	GCPService
}

// Generate TerraformResources from GCP API,
func (g *DialogflowGenerator) InitResources() error {
	ctx := context.Background()
	dialogflowService, err := dialogflow.NewService(ctx)
	if err != nil {
		return err
	}

	location := g.GetArgs()["region"].(compute.Region).Name
	project := g.GetArgs()["project"].(string)
	agentNames := []string{}
	agentsList := dialogflowService.Projects.Locations.Agents.List("projects/" + project + "/locations/" + location)
	if err := agentsList.Pages(ctx, func(page *dialogflow.GoogleCloudDialogflowCxV3ListAgentsResponse) error {
		for _, obj := range page.Agents {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			agentNames = append(agentNames, obj.Name)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name, name, "google_dialogflow_cx_agent", g.ProviderName,
				map[string]string{"name": name, "project": project, "location": location},
				dialogflowAllowEmptyValues, dialogflowAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	for _, agent := range agentNames {
		if err := dialogflowService.Projects.Locations.Agents.Flows.List(agent).Pages(ctx, func(p *dialogflow.GoogleCloudDialogflowCxV3ListFlowsResponse) error {
			for _, o := range p.Flows {
				t := strings.Split(o.Name, "/")
				g.Resources = append(g.Resources, terraformutils.NewResource(
					o.Name, t[len(t)-1], "google_dialogflow_cx_flow", g.ProviderName,
					map[string]string{"parent": agent, "project": project},
					dialogflowAllowEmptyValues, dialogflowAdditionalFields))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
		if err := dialogflowService.Projects.Locations.Agents.Intents.List(agent).Pages(ctx, func(p *dialogflow.GoogleCloudDialogflowCxV3ListIntentsResponse) error {
			for _, o := range p.Intents {
				t := strings.Split(o.Name, "/")
				g.Resources = append(g.Resources, terraformutils.NewResource(
					o.Name, t[len(t)-1], "google_dialogflow_cx_intent", g.ProviderName,
					map[string]string{"parent": agent, "project": project},
					dialogflowAllowEmptyValues, dialogflowAdditionalFields))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
		if err := dialogflowService.Projects.Locations.Agents.EntityTypes.List(agent).Pages(ctx, func(p *dialogflow.GoogleCloudDialogflowCxV3ListEntityTypesResponse) error {
			for _, o := range p.EntityTypes {
				t := strings.Split(o.Name, "/")
				g.Resources = append(g.Resources, terraformutils.NewResource(
					o.Name, t[len(t)-1], "google_dialogflow_cx_entity_type", g.ProviderName,
					map[string]string{"parent": agent, "project": project},
					dialogflowAllowEmptyValues, dialogflowAdditionalFields))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
		if err := dialogflowService.Projects.Locations.Agents.Webhooks.List(agent).Pages(ctx, func(p *dialogflow.GoogleCloudDialogflowCxV3ListWebhooksResponse) error {
			for _, o := range p.Webhooks {
				t := strings.Split(o.Name, "/")
				g.Resources = append(g.Resources, terraformutils.NewResource(
					o.Name, t[len(t)-1], "google_dialogflow_cx_webhook", g.ProviderName,
					map[string]string{"parent": agent, "project": project},
					dialogflowAllowEmptyValues, dialogflowAdditionalFields))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
	}
	return nil
}
