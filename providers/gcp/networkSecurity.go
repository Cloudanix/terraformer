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
	"google.golang.org/api/networksecurity/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var networkSecurityAllowEmptyValues = []string{""}

var networkSecurityAdditionalFields = map[string]interface{}{}

type NetworkSecurityGenerator struct {
	GCPService
}

func (g NetworkSecurityGenerator) createServerTLSResources(ctx context.Context, list *networksecurity.ProjectsLocationsServerTlsPoliciesListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := list.Pages(ctx, func(page *networksecurity.ListServerTlsPoliciesResponse) error {
		for _, obj := range page.ServerTlsPolicies {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name, name, "google_network_security_server_tls_policy", g.ProviderName,
				map[string]string{"name": name, "project": g.GetArgs()["project"].(string), "location": location},
				networkSecurityAllowEmptyValues, networkSecurityAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

func (g NetworkSecurityGenerator) createClientTLSResources(ctx context.Context, list *networksecurity.ProjectsLocationsClientTlsPoliciesListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := list.Pages(ctx, func(page *networksecurity.ListClientTlsPoliciesResponse) error {
		for _, obj := range page.ClientTlsPolicies {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name, name, "google_network_security_client_tls_policy", g.ProviderName,
				map[string]string{"name": name, "project": g.GetArgs()["project"].(string), "location": location},
				networkSecurityAllowEmptyValues, networkSecurityAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *NetworkSecurityGenerator) InitResources() error {
	ctx := context.Background()
	nsService, err := networksecurity.NewService(ctx)
	if err != nil {
		return err
	}
	parent := "projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name
	g.Resources = append(g.Resources, g.createServerTLSResources(ctx, nsService.Projects.Locations.ServerTlsPolicies.List(parent))...)
	g.Resources = append(g.Resources, g.createClientTLSResources(ctx, nsService.Projects.Locations.ClientTlsPolicies.List(parent))...)
	return nil
}
