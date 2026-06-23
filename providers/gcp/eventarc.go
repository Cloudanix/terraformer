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

	triggersList := eventarcService.Projects.Locations.Triggers.List(
		"projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name)
	g.Resources = g.createResources(ctx, triggersList)
	return nil
}
