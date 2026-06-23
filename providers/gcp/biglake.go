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

	"google.golang.org/api/biglake/v1"
	"google.golang.org/api/compute/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var biglakeAllowEmptyValues = []string{""}

var biglakeAdditionalFields = map[string]interface{}{}

type BiglakeGenerator struct {
	GCPService
}

// Run on catalogsList and create for each TerraformResource
func (g BiglakeGenerator) createResources(ctx context.Context, catalogsList *biglake.ProjectsLocationsCatalogsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := catalogsList.Pages(ctx, func(page *biglake.ListCatalogsResponse) error {
		for _, obj := range page.Catalogs {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_biglake_catalog",
				g.ProviderName,
				map[string]string{
					"name":     name,
					"project":  g.GetArgs()["project"].(string),
					"location": location,
				},
				biglakeAllowEmptyValues,
				biglakeAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *BiglakeGenerator) InitResources() error {
	ctx := context.Background()
	biglakeService, err := biglake.NewService(ctx)
	if err != nil {
		return err
	}

	catalogsList := biglakeService.Projects.Locations.Catalogs.List(
		"projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name)
	g.Resources = g.createResources(ctx, catalogsList)
	return nil
}
