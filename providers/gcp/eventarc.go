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
	"google.golang.org/api/eventarc/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var eventarcAllowEmptyValues = []string{""}

var eventarcAdditionalFields = map[string]interface{}{}

type EventarcGenerator struct {
	GCPService
}

// Run on triggersList and create for each TerraformResource
func (g EventarcGenerator) createResources(ctx context.Context, triggersList *eventarc.ProjectsLocationsTriggersListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := triggersList.Pages(ctx, func(page *eventarc.ListTriggersResponse) error {
		for _, obj := range page.Triggers {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_eventarc_trigger",
				g.ProviderName,
				map[string]string{
					"name":     name,
					"project":  g.GetArgs()["project"].(string),
					"location": location,
				},
				eventarcAllowEmptyValues,
				eventarcAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *EventarcGenerator) InitResources() error {
	ctx := context.Background()
	eventarcService, err := eventarc.NewService(ctx)
	if err != nil {
		return err
	}

	parent := "projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name
	triggersList := eventarcService.Projects.Locations.Triggers.List(parent)
	g.Resources = append(g.Resources, g.createResources(ctx, triggersList)...)

	channelsList := eventarcService.Projects.Locations.Channels.List(parent)
	g.Resources = append(g.Resources, g.createChannelsResources(ctx, channelsList)...)

	loc := g.GetArgs()["region"].(compute.Region).Name
	proj := g.GetArgs()["project"].(string)
	if err := eventarcService.Projects.Locations.Pipelines.List(parent).Pages(ctx, func(p *eventarc.ListPipelinesResponse) error {
		for _, o := range p.Pipelines {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_eventarc_pipeline", g.ProviderName,
				map[string]string{"pipeline_id": name, "project": proj, "location": loc},
				eventarcAllowEmptyValues, eventarcAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := eventarcService.Projects.Locations.GoogleApiSources.List(parent).Pages(ctx, func(p *eventarc.ListGoogleApiSourcesResponse) error {
		for _, o := range p.GoogleApiSources {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_eventarc_google_api_source", g.ProviderName,
				map[string]string{"google_api_source_id": name, "project": proj, "location": loc},
				eventarcAllowEmptyValues, eventarcAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}

// Run on channelsList and create for each TerraformResource
func (g EventarcGenerator) createChannelsResources(ctx context.Context, list *eventarc.ProjectsLocationsChannelsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := list.Pages(ctx, func(page *eventarc.ListChannelsResponse) error {
		for _, obj := range page.Channels {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_eventarc_channel",
				g.ProviderName,
				map[string]string{
					"name":     name,
					"project":  g.GetArgs()["project"].(string),
					"location": location,
				},
				eventarcAllowEmptyValues,
				eventarcAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}
