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

	"google.golang.org/api/aiplatform/v1"
	"google.golang.org/api/compute/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var vertexAIAllowEmptyValues = []string{""}

var vertexAIAdditionalFields = map[string]interface{}{}

type VertexAIGenerator struct {
	GCPService
}

// Run on endpointsList and create for each TerraformResource
func (g VertexAIGenerator) createResources(ctx context.Context, endpointsList *aiplatform.ProjectsLocationsEndpointsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := endpointsList.Pages(ctx, func(page *aiplatform.GoogleCloudAiplatformV1ListEndpointsResponse) error {
		for _, obj := range page.Endpoints {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_vertex_ai_endpoint",
				g.ProviderName,
				map[string]string{
					"name":     name,
					"project":  g.GetArgs()["project"].(string),
					"location": location,
				},
				vertexAIAllowEmptyValues,
				vertexAIAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *VertexAIGenerator) InitResources() error {
	ctx := context.Background()
	vertexAIService, err := aiplatform.NewService(ctx)
	if err != nil {
		return err
	}

	endpointsList := vertexAIService.Projects.Locations.Endpoints.List(
		"projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name)
	g.Resources = g.createResources(ctx, endpointsList)
	return nil
}
