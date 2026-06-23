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
	"google.golang.org/api/networkconnectivity/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var networkConnectivityAllowEmptyValues = []string{""}

var networkConnectivityAdditionalFields = map[string]interface{}{}

type NetworkConnectivityGenerator struct {
	GCPService
}

// Run on hubsList and create for each TerraformResource
func (g NetworkConnectivityGenerator) createResources(ctx context.Context, hubsList *networkconnectivity.ProjectsLocationsGlobalHubsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	if err := hubsList.Pages(ctx, func(page *networkconnectivity.ListHubsResponse) error {
		for _, obj := range page.Hubs {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_network_connectivity_hub",
				g.ProviderName,
				map[string]string{
					"name":    name,
					"project": g.GetArgs()["project"].(string),
				},
				networkConnectivityAllowEmptyValues,
				networkConnectivityAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *NetworkConnectivityGenerator) InitResources() error {
	ctx := context.Background()
	networkConnectivityService, err := networkconnectivity.NewService(ctx)
	if err != nil {
		return err
	}

	hubsList := networkConnectivityService.Projects.Locations.Global.Hubs.List(
		"projects/" + g.GetArgs()["project"].(string) + "/locations/global")
	g.Resources = append(g.Resources, g.createResources(ctx, hubsList)...)

	regionalParent := "projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name
	spokesList := networkConnectivityService.Projects.Locations.Spokes.List(regionalParent)
	g.Resources = append(g.Resources, g.createSpokesResources(ctx, spokesList)...)

	scpList := networkConnectivityService.Projects.Locations.ServiceConnectionPolicies.List(regionalParent)
	g.Resources = append(g.Resources, g.createSCPResources(ctx, scpList)...)
	return nil
}

// Run on serviceConnectionPoliciesList and create for each TerraformResource
func (g NetworkConnectivityGenerator) createSCPResources(ctx context.Context, list *networkconnectivity.ProjectsLocationsServiceConnectionPoliciesListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := list.Pages(ctx, func(page *networkconnectivity.ListServiceConnectionPoliciesResponse) error {
		for _, obj := range page.ServiceConnectionPolicies {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name, name, "google_network_connectivity_service_connection_policy", g.ProviderName,
				map[string]string{"name": name, "project": g.GetArgs()["project"].(string), "location": location},
				networkConnectivityAllowEmptyValues, networkConnectivityAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Run on spokesList and create for each TerraformResource
func (g NetworkConnectivityGenerator) createSpokesResources(ctx context.Context, list *networkconnectivity.ProjectsLocationsSpokesListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := list.Pages(ctx, func(page *networkconnectivity.ListSpokesResponse) error {
		for _, obj := range page.Spokes {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_network_connectivity_spoke",
				g.ProviderName,
				map[string]string{
					"name":     name,
					"project":  g.GetArgs()["project"].(string),
					"location": location,
				},
				networkConnectivityAllowEmptyValues,
				networkConnectivityAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}
