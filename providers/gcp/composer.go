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

	"google.golang.org/api/composer/v1"
	"google.golang.org/api/compute/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var composerAllowEmptyValues = []string{""}

var composerAdditionalFields = map[string]interface{}{}

type ComposerGenerator struct {
	GCPService
}

// Run on environmentsList and create for each TerraformResource
func (g ComposerGenerator) createResources(ctx context.Context, environmentsList *composer.ProjectsLocationsEnvironmentsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	region := g.GetArgs()["region"].(compute.Region).Name
	if err := environmentsList.Pages(ctx, func(page *composer.ListEnvironmentsResponse) error {
		for _, obj := range page.Environments {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_composer_environment",
				g.ProviderName,
				map[string]string{
					"name":    name,
					"project": g.GetArgs()["project"].(string),
					"region":  region,
				},
				composerAllowEmptyValues,
				composerAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *ComposerGenerator) InitResources() error {
	ctx := context.Background()
	composerService, err := composer.NewService(ctx)
	if err != nil {
		return err
	}

	environmentsList := composerService.Projects.Locations.Environments.List(
		"projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name)
	g.Resources = g.createResources(ctx, environmentsList)
	return nil
}
