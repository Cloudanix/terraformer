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
	"google.golang.org/api/migrationcenter/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var migrationCenterAllowEmptyValues = []string{""}

var migrationCenterAdditionalFields = map[string]interface{}{}

type MigrationCenterGenerator struct {
	GCPService
}

// Run on groupsList and create for each TerraformResource
func (g MigrationCenterGenerator) createResources(ctx context.Context, list *migrationcenter.ProjectsLocationsGroupsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := list.Pages(ctx, func(page *migrationcenter.ListGroupsResponse) error {
		for _, obj := range page.Groups {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_migration_center_group",
				g.ProviderName,
				map[string]string{
					"group_id": name,
					"project":  g.GetArgs()["project"].(string),
					"location": location,
				},
				migrationCenterAllowEmptyValues,
				migrationCenterAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *MigrationCenterGenerator) InitResources() error {
	ctx := context.Background()
	mcService, err := migrationcenter.NewService(ctx)
	if err != nil {
		return err
	}

	parent := "projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name
	groupsList := mcService.Projects.Locations.Groups.List(parent)
	g.Resources = append(g.Resources, g.createResources(ctx, groupsList)...)

	if err := mcService.Projects.Locations.Sources.List(parent).Pages(ctx, func(p *migrationcenter.ListSourcesResponse) error {
		for _, o := range p.Sources {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_migration_center_source", g.ProviderName,
				map[string]string{"source_id": name, "project": g.GetArgs()["project"].(string), "location": g.GetArgs()["region"].(compute.Region).Name},
				migrationCenterAllowEmptyValues, migrationCenterAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}
