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
	"google.golang.org/api/developerconnect/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var developerConnectAllowEmptyValues = []string{""}

var developerConnectAdditionalFields = map[string]interface{}{}

type DeveloperConnectGenerator struct {
	GCPService
}

// Run on connectionsList and create for each TerraformResource
func (g DeveloperConnectGenerator) createResources(ctx context.Context, list *developerconnect.ProjectsLocationsConnectionsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := list.Pages(ctx, func(page *developerconnect.ListConnectionsResponse) error {
		for _, obj := range page.Connections {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_developer_connect_connection",
				g.ProviderName,
				map[string]string{
					"connection_id": name,
					"project":       g.GetArgs()["project"].(string),
					"location":      location,
				},
				developerConnectAllowEmptyValues,
				developerConnectAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *DeveloperConnectGenerator) InitResources() error {
	ctx := context.Background()
	dcService, err := developerconnect.NewService(ctx)
	if err != nil {
		return err
	}

	project := g.GetArgs()["project"].(string)
	location := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + location
	tail := func(s string) string { p := strings.Split(s, "/"); return p[len(p)-1] }

	connectionsList := dcService.Projects.Locations.Connections.List(parent)
	connResources := g.createResources(ctx, connectionsList)
	g.Resources = connResources
	for _, cr := range connResources {
		conn := tail(cr.InstanceState.ID)
		if e := dcService.Projects.Locations.Connections.GitRepositoryLinks.List(cr.InstanceState.ID).Pages(ctx, func(rp *developerconnect.ListGitRepositoryLinksResponse) error {
			for _, o := range rp.GitRepositoryLinks {
				g.Resources = append(g.Resources, terraformutils.NewResource(
					o.Name, conn+"_"+tail(o.Name), "google_developer_connect_git_repository_link", g.ProviderName,
					map[string]string{"git_repository_link_id": tail(o.Name), "parent_connection": conn, "location": location, "project": project},
					developerConnectAllowEmptyValues, developerConnectAdditionalFields))
			}
			return nil
		}); e != nil {
			log.Println(e)
		}
	}

	if err := dcService.Projects.Locations.AccountConnectors.List(parent).Pages(ctx, func(p *developerconnect.ListAccountConnectorsResponse) error {
		for _, o := range p.AccountConnectors {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, tail(o.Name), "google_developer_connect_account_connector", g.ProviderName,
				map[string]string{"account_connector_id": tail(o.Name), "location": location, "project": project},
				developerConnectAllowEmptyValues, developerConnectAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := dcService.Projects.Locations.InsightsConfigs.List(parent).Pages(ctx, func(p *developerconnect.ListInsightsConfigsResponse) error {
		for _, o := range p.InsightsConfigs {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, tail(o.Name), "google_developer_connect_insights_config", g.ProviderName,
				map[string]string{"insights_config_id": tail(o.Name), "location": location, "project": project},
				developerConnectAllowEmptyValues, developerConnectAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}
