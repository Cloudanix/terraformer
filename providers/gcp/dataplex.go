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
	"google.golang.org/api/dataplex/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var dataplexAllowEmptyValues = []string{""}

var dataplexAdditionalFields = map[string]interface{}{}

type DataplexGenerator struct {
	GCPService
}

// Generate TerraformResources from GCP API,
func (g *DataplexGenerator) InitResources() error {
	ctx := context.Background()
	dataplexService, err := dataplex.NewService(ctx)
	if err != nil {
		return err
	}
	project := g.GetArgs()["project"].(string)
	location := g.GetArgs()["region"].(compute.Region).Name

	entryGroupsList := dataplexService.Projects.Locations.EntryGroups.List("projects/" + project + "/locations/" + location)
	if err := entryGroupsList.Pages(ctx, func(page *dataplex.GoogleCloudDataplexV1ListEntryGroupsResponse) error {
		for _, obj := range page.EntryGroups {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name, name, "google_dataplex_entry_group", g.ProviderName,
				map[string]string{"entry_group_id": name, "project": project, "location": location},
				dataplexAllowEmptyValues, dataplexAdditionalFields,
			))
			if policy, perr := dataplexService.Projects.Locations.EntryGroups.GetIamPolicy(obj.Name).Do(); perr == nil {
				for _, b := range policy.Bindings {
					for _, m := range b.Members {
						g.Resources = append(g.Resources, terraformutils.NewResource(
							obj.Name+" "+b.Role+" "+m, name+"_"+b.Role+"_"+m,
							"google_dataplex_entry_group_iam_member", g.ProviderName,
							map[string]string{"entry_group_id": name, "role": b.Role, "member": m, "project": project, "location": location},
							dataplexAllowEmptyValues, dataplexAdditionalFields))
					}
				}
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	if err := dataplexService.Projects.Locations.EntryTypes.List("projects/"+project+"/locations/"+location).Pages(ctx, func(p *dataplex.GoogleCloudDataplexV1ListEntryTypesResponse) error {
		for _, o := range p.EntryTypes {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_dataplex_entry_type", g.ProviderName,
				map[string]string{"entry_type_id": name, "project": project, "location": location},
				dataplexAllowEmptyValues, dataplexAdditionalFields))
			if policy, perr := dataplexService.Projects.Locations.EntryTypes.GetIamPolicy(o.Name).Do(); perr == nil {
				for _, b := range policy.Bindings {
					for _, m := range b.Members {
						g.Resources = append(g.Resources, terraformutils.NewResource(
							o.Name+" "+b.Role+" "+m, name+"_"+b.Role+"_"+m,
							"google_dataplex_entry_type_iam_member", g.ProviderName,
							map[string]string{"entry_type_id": name, "role": b.Role, "member": m, "project": project, "location": location},
							dataplexAllowEmptyValues, dataplexAdditionalFields))
					}
				}
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := dataplexService.Projects.Locations.DataScans.List("projects/"+project+"/locations/"+location).Pages(ctx, func(p *dataplex.GoogleCloudDataplexV1ListDataScansResponse) error {
		for _, o := range p.DataScans {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_dataplex_datascan", g.ProviderName,
				map[string]string{"data_scan_id": name, "project": project, "location": location},
				dataplexAllowEmptyValues, dataplexAdditionalFields))
			if policy, perr := dataplexService.Projects.Locations.DataScans.GetIamPolicy(o.Name).Do(); perr == nil {
				for _, b := range policy.Bindings {
					for _, m := range b.Members {
						g.Resources = append(g.Resources, terraformutils.NewResource(
							o.Name+" "+b.Role+" "+m, name+"_"+b.Role+"_"+m,
							"google_dataplex_datascan_iam_member", g.ProviderName,
							map[string]string{"data_scan_id": name, "role": b.Role, "member": m, "project": project, "location": location},
							dataplexAllowEmptyValues, dataplexAdditionalFields))
					}
				}
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	if err := dataplexService.Projects.Locations.Glossaries.List("projects/"+project+"/locations/"+location).Pages(ctx, func(p *dataplex.GoogleCloudDataplexV1ListGlossariesResponse) error {
		for _, o := range p.Glossaries {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_dataplex_glossary", g.ProviderName,
				map[string]string{"glossary_id": name, "project": project, "location": location},
				dataplexAllowEmptyValues, dataplexAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	aspectTypesList := dataplexService.Projects.Locations.AspectTypes.List("projects/" + project + "/locations/" + location)
	if err := aspectTypesList.Pages(ctx, func(page *dataplex.GoogleCloudDataplexV1ListAspectTypesResponse) error {
		for _, obj := range page.AspectTypes {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name, name, "google_dataplex_aspect_type", g.ProviderName,
				map[string]string{"aspect_type_id": name, "project": project, "location": location},
				dataplexAllowEmptyValues, dataplexAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	lakeNames := []string{}
	lakesList := dataplexService.Projects.Locations.Lakes.List("projects/" + project + "/locations/" + location)
	if err := lakesList.Pages(ctx, func(page *dataplex.GoogleCloudDataplexV1ListLakesResponse) error {
		for _, obj := range page.Lakes {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			lakeNames = append(lakeNames, name)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_dataplex_lake",
				g.ProviderName,
				map[string]string{"name": name, "project": project, "location": location},
				dataplexAllowEmptyValues,
				dataplexAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	// Walk each lake for its zones and tasks.
	for _, lake := range lakeNames {
		lakePath := "projects/" + project + "/locations/" + location + "/lakes/" + lake
		if policy, perr := dataplexService.Projects.Locations.Lakes.GetIamPolicy(lakePath).Do(); perr == nil {
			for _, b := range policy.Bindings {
				for _, m := range b.Members {
					g.Resources = append(g.Resources, terraformutils.NewResource(
						lakePath+" "+b.Role+" "+m, lake+"_"+b.Role+"_"+m,
						"google_dataplex_lake_iam_member", g.ProviderName,
						map[string]string{"lake": lake, "role": b.Role, "member": m, "project": project, "location": location},
						dataplexAllowEmptyValues, dataplexAdditionalFields))
				}
			}
		}
		tasksList := dataplexService.Projects.Locations.Lakes.Tasks.List(
			"projects/" + project + "/locations/" + location + "/lakes/" + lake)
		if err := tasksList.Pages(ctx, func(page *dataplex.GoogleCloudDataplexV1ListTasksResponse) error {
			for _, obj := range page.Tasks {
				tt := strings.Split(obj.Name, "/")
				name := tt[len(tt)-1]
				g.Resources = append(g.Resources, terraformutils.NewResource(
					obj.Name,
					name,
					"google_dataplex_task",
					g.ProviderName,
					map[string]string{
						"task_id":  name,
						"lake":     lake,
						"project":  project,
						"location": location,
					},
					dataplexAllowEmptyValues,
					dataplexAdditionalFields,
				))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}

		zonesList := dataplexService.Projects.Locations.Lakes.Zones.List(
			"projects/" + project + "/locations/" + location + "/lakes/" + lake)
		if err := zonesList.Pages(ctx, func(page *dataplex.GoogleCloudDataplexV1ListZonesResponse) error {
			for _, obj := range page.Zones {
				t := strings.Split(obj.Name, "/")
				name := t[len(t)-1]
				g.Resources = append(g.Resources, terraformutils.NewResource(
					obj.Name,
					name,
					"google_dataplex_zone",
					g.ProviderName,
					map[string]string{
						"name":     name,
						"lake":     lake,
						"project":  project,
						"location": location,
					},
					dataplexAllowEmptyValues,
					dataplexAdditionalFields,
				))
				if policy, perr := dataplexService.Projects.Locations.Lakes.Zones.GetIamPolicy(obj.Name).Do(); perr == nil {
					for _, b := range policy.Bindings {
						for _, m := range b.Members {
							g.Resources = append(g.Resources, terraformutils.NewResource(
								obj.Name+" "+b.Role+" "+m, name+"_"+b.Role+"_"+m,
								"google_dataplex_zone_iam_member", g.ProviderName,
								map[string]string{"dataplex_zone": name, "lake": lake, "role": b.Role, "member": m, "project": project, "location": location},
								dataplexAllowEmptyValues, dataplexAdditionalFields))
						}
					}
				}
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
	}
	return nil
}
