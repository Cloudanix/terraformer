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
	"google.golang.org/api/gkehub/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var gkeHubAllowEmptyValues = []string{""}

var gkeHubAdditionalFields = map[string]interface{}{}

type GkeHubGenerator struct {
	GCPService
}

// Run on membershipsList and create for each TerraformResource
func (g GkeHubGenerator) createResources(ctx context.Context, membershipsList *gkehub.ProjectsLocationsMembershipsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := membershipsList.Pages(ctx, func(page *gkehub.ListMembershipsResponse) error {
		for _, obj := range page.Resources {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_gke_hub_membership",
				g.ProviderName,
				map[string]string{
					"name":          name,
					"membership_id": name,
					"project":       g.GetArgs()["project"].(string),
					"location":      location,
				},
				gkeHubAllowEmptyValues,
				gkeHubAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *GkeHubGenerator) InitResources() error {
	ctx := context.Background()
	gkeHubService, err := gkehub.NewService(ctx)
	if err != nil {
		return err
	}

	parent := "projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name
	membershipsList := gkeHubService.Projects.Locations.Memberships.List(parent)
	g.Resources = append(g.Resources, g.createResources(ctx, membershipsList)...)

	featuresList := gkeHubService.Projects.Locations.Features.List(parent)
	g.Resources = append(g.Resources, g.createFeaturesResources(ctx, featuresList)...)
	return nil
}

// Run on featuresList and create for each TerraformResource (response field is Resources)
func (g GkeHubGenerator) createFeaturesResources(ctx context.Context, list *gkehub.ProjectsLocationsFeaturesListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := list.Pages(ctx, func(page *gkehub.ListFeaturesResponse) error {
		for _, obj := range page.Resources {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_gke_hub_feature",
				g.ProviderName,
				map[string]string{
					"name":     name,
					"project":  g.GetArgs()["project"].(string),
					"location": location,
				},
				gkeHubAllowEmptyValues,
				gkeHubAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}
