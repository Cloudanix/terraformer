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

	"google.golang.org/api/artifactregistry/v1"
	"google.golang.org/api/compute/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var artifactRegistryAllowEmptyValues = []string{""}

var artifactRegistryAdditionalFields = map[string]interface{}{}

type ArtifactRegistryGenerator struct {
	GCPService
}

// Run on repositoriesList and create for each TerraformResource
func (g ArtifactRegistryGenerator) createResources(ctx context.Context, repositoriesList *artifactregistry.ProjectsLocationsRepositoriesListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	if err := repositoriesList.Pages(ctx, func(page *artifactregistry.ListRepositoriesResponse) error {
		for _, obj := range page.Repositories {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_artifact_registry_repository",
				g.ProviderName,
				map[string]string{
					"name":          name,
					"repository_id": name,
					"project":       g.GetArgs()["project"].(string),
					"location":      g.GetArgs()["region"].(compute.Region).Name,
				},
				artifactRegistryAllowEmptyValues,
				artifactRegistryAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *ArtifactRegistryGenerator) InitResources() error {
	ctx := context.Background()
	artifactRegistryService, err := artifactregistry.NewService(ctx)
	if err != nil {
		return err
	}

	repositoriesList := artifactRegistryService.Projects.Locations.Repositories.List(
		"projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name)

	g.Resources = g.createResources(ctx, repositoriesList)
	return nil
}
