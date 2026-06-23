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

// Generate TerraformResources from GCP API,
func (g *BiglakeGenerator) InitResources() error {
	ctx := context.Background()
	biglakeService, err := biglake.NewService(ctx)
	if err != nil {
		return err
	}
	project := g.GetArgs()["project"].(string)
	location := g.GetArgs()["region"].(compute.Region).Name

	catalogIDs := []string{}
	catalogsList := biglakeService.Projects.Locations.Catalogs.List("projects/" + project + "/locations/" + location)
	if err := catalogsList.Pages(ctx, func(page *biglake.ListCatalogsResponse) error {
		for _, obj := range page.Catalogs {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			catalogIDs = append(catalogIDs, name)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name, name, "google_biglake_catalog", g.ProviderName,
				map[string]string{"name": name, "project": project, "location": location},
				biglakeAllowEmptyValues, biglakeAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	for _, catalog := range catalogIDs {
		dbList := biglakeService.Projects.Locations.Catalogs.Databases.List(
			"projects/" + project + "/locations/" + location + "/catalogs/" + catalog)
		dbNames := []string{}
		if err := dbList.Pages(ctx, func(page *biglake.ListDatabasesResponse) error {
			for _, obj := range page.Databases {
				t := strings.Split(obj.Name, "/")
				name := t[len(t)-1]
				dbNames = append(dbNames, name)
				g.Resources = append(g.Resources, terraformutils.NewResource(
					obj.Name, name, "google_biglake_database", g.ProviderName,
					map[string]string{"name": name, "catalog": catalog, "project": project, "location": location},
					biglakeAllowEmptyValues, biglakeAdditionalFields,
				))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
		for _, db := range dbNames {
			if err := biglakeService.Projects.Locations.Catalogs.Databases.Tables.List(
				"projects/"+project+"/locations/"+location+"/catalogs/"+catalog+"/databases/"+db).Pages(ctx, func(page *biglake.ListTablesResponse) error {
				for _, obj := range page.Tables {
					t := strings.Split(obj.Name, "/")
					name := t[len(t)-1]
					g.Resources = append(g.Resources, terraformutils.NewResource(
						obj.Name, name, "google_biglake_table", g.ProviderName,
						map[string]string{"name": name, "database": db, "catalog": catalog, "project": project, "location": location},
						biglakeAllowEmptyValues, biglakeAdditionalFields))
				}
				return nil
			}); err != nil {
				log.Println(err)
			}
		}
	}
	return nil
}
