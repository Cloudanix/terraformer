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

	"google.golang.org/api/dialogflow/v2"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var dialogflowEsAllowEmptyValues = []string{""}

var dialogflowEsAdditionalFields = map[string]interface{}{}

type DialogflowEsGenerator struct {
	GCPService
}

// Generate TerraformResources from GCP Dialogflow ES (v2) API.
func (g *DialogflowEsGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := dialogflow.NewService(ctx)
	if err != nil {
		return err
	}
	project := g.GetArgs()["project"].(string)
	agentParent := "projects/" + project + "/agent"
	projParent := "projects/" + project
	tail := func(s string) string { p := strings.Split(s, "/"); return p[len(p)-1] }
	add := func(id, name, tfType string) {
		g.Resources = append(g.Resources, terraformutils.NewResource(
			id, name, tfType, g.ProviderName, map[string]string{"project": project},
			dialogflowEsAllowEmptyValues, dialogflowEsAdditionalFields))
	}

	// Single ES agent per project (emit only if one exists).
	if _, ae := svc.Projects.GetAgent(projParent).Do(); ae == nil {
		add(agentParent, project+"_agent", "google_dialogflow_agent")
	}

	if err := svc.Projects.Agent.EntityTypes.List(agentParent).Pages(ctx, func(p *dialogflow.GoogleCloudDialogflowV2ListEntityTypesResponse) error {
		for _, o := range p.EntityTypes {
			add(o.Name, tail(o.Name), "google_dialogflow_entity_type")
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := svc.Projects.Agent.Intents.List(agentParent).Pages(ctx, func(p *dialogflow.GoogleCloudDialogflowV2ListIntentsResponse) error {
		for _, o := range p.Intents {
			add(o.Name, tail(o.Name), "google_dialogflow_intent")
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := svc.Projects.Agent.Versions.List(agentParent).Pages(ctx, func(p *dialogflow.GoogleCloudDialogflowV2ListVersionsResponse) error {
		for _, o := range p.Versions {
			add(o.Name, tail(o.Name), "google_dialogflow_version")
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := svc.Projects.Agent.Environments.List(agentParent).Pages(ctx, func(p *dialogflow.GoogleCloudDialogflowV2ListEnvironmentsResponse) error {
		for _, o := range p.Environments {
			add(o.Name, tail(o.Name), "google_dialogflow_environment")
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := svc.Projects.ConversationProfiles.List(projParent).Pages(ctx, func(p *dialogflow.GoogleCloudDialogflowV2ListConversationProfilesResponse) error {
		for _, o := range p.ConversationProfiles {
			add(o.Name, tail(o.Name), "google_dialogflow_conversation_profile")
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := svc.Projects.Generators.List(projParent).Pages(ctx, func(p *dialogflow.GoogleCloudDialogflowV2ListGeneratorsResponse) error {
		for _, o := range p.Generators {
			add(o.Name, tail(o.Name), "google_dialogflow_generator")
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}
