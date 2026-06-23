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

// Run on agentsList and create for each TerraformResource
func (g DialogflowGenerator) createResources(ctx context.Context, agentsList *dialogflow.ProjectsLocationsAgentsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := agentsList.Pages(ctx, func(page *dialogflow.GoogleCloudDialogflowCxV3ListAgentsResponse) error {
		for _, obj := range page.Agents {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_dialogflow_cx_agent",
				g.ProviderName,
				map[string]string{
					"name":     name,
					"project":  g.GetArgs()["project"].(string),
					"location": location,
				},
				dialogflowAllowEmptyValues,
				dialogflowAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *DialogflowGenerator) InitResources() error {
	ctx := context.Background()
	dialogflowService, err := dialogflow.NewService(ctx)
	if err != nil {
		return err
	}

	agentsList := dialogflowService.Projects.Locations.Agents.List(
		"projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name)
	g.Resources = g.createResources(ctx, agentsList)
	return nil
}
