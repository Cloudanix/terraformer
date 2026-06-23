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
	"google.golang.org/api/privateca/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var privatecaAllowEmptyValues = []string{""}

var privatecaAdditionalFields = map[string]interface{}{}

type PrivatecaGenerator struct {
	GCPService
}

// Run on caPoolsList and create for each TerraformResource
func (g PrivatecaGenerator) createResources(ctx context.Context, caPoolsList *privateca.ProjectsLocationsCaPoolsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := caPoolsList.Pages(ctx, func(page *privateca.ListCaPoolsResponse) error {
		for _, obj := range page.CaPools {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_privateca_ca_pool",
				g.ProviderName,
				map[string]string{
					"name":     name,
					"project":  g.GetArgs()["project"].(string),
					"location": location,
				},
				privatecaAllowEmptyValues,
				privatecaAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *PrivatecaGenerator) InitResources() error {
	ctx := context.Background()
	privatecaService, err := privateca.NewService(ctx)
	if err != nil {
		return err
	}

	caPoolsList := privatecaService.Projects.Locations.CaPools.List(
		"projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name)
	g.Resources = g.createResources(ctx, caPoolsList)
	return nil
}
