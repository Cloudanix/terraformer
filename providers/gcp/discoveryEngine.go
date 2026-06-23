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
	"google.golang.org/api/discoveryengine/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var discoveryEngineAllowEmptyValues = []string{""}

var discoveryEngineAdditionalFields = map[string]interface{}{}

type DiscoveryEngineGenerator struct {
	GCPService
}

// Generate TerraformResources from GCP API,
func (g *DiscoveryEngineGenerator) InitResources() error {
	ctx := context.Background()
	discoveryEngineService, err := discoveryengine.NewService(ctx)
	if err != nil {
		return err
	}
	project := g.GetArgs()["project"].(string)
	location := g.GetArgs()["region"].(compute.Region).Name

	// Data stores live under the always-present default_collection.
	parent := "projects/" + project + "/locations/" + location + "/collections/default_collection"
	dataStoresList := discoveryEngineService.Projects.Locations.Collections.DataStores.List(parent)
	if err := dataStoresList.Pages(ctx, func(page *discoveryengine.GoogleCloudDiscoveryengineV1ListDataStoresResponse) error {
		for _, obj := range page.DataStores {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_discovery_engine_data_store",
				g.ProviderName,
				map[string]string{
					"data_store_id": name,
					"location":      location,
					"project":       project,
				},
				discoveryEngineAllowEmptyValues,
				discoveryEngineAdditionalFields,
			))
			dsAttrs := func(id string) map[string]string {
				return map[string]string{"data_store_id": name, "location": location, "project": project, "id": id}
			}
			if e := discoveryEngineService.Projects.Locations.Collections.DataStores.Schemas.List(obj.Name).Pages(ctx, func(sp *discoveryengine.GoogleCloudDiscoveryengineV1ListSchemasResponse) error {
				for _, o := range sp.Schemas {
					g.Resources = append(g.Resources, terraformutils.NewResource(o.Name, name+"_"+lastSeg(o.Name), "google_discovery_engine_schema", g.ProviderName, dsAttrs(lastSeg(o.Name)), discoveryEngineAllowEmptyValues, discoveryEngineAdditionalFields))
				}
				return nil
			}); e != nil {
				log.Println(e)
			}
			if e := discoveryEngineService.Projects.Locations.Collections.DataStores.Controls.List(obj.Name).Pages(ctx, func(cp *discoveryengine.GoogleCloudDiscoveryengineV1ListControlsResponse) error {
				for _, o := range cp.Controls {
					g.Resources = append(g.Resources, terraformutils.NewResource(o.Name, name+"_"+lastSeg(o.Name), "google_discovery_engine_control", g.ProviderName, dsAttrs(lastSeg(o.Name)), discoveryEngineAllowEmptyValues, discoveryEngineAdditionalFields))
				}
				return nil
			}); e != nil {
				log.Println(e)
			}
			if e := discoveryEngineService.Projects.Locations.Collections.DataStores.SiteSearchEngine.TargetSites.List(obj.Name+"/siteSearchEngine").Pages(ctx, func(tp *discoveryengine.GoogleCloudDiscoveryengineV1ListTargetSitesResponse) error {
				for _, o := range tp.TargetSites {
					g.Resources = append(g.Resources, terraformutils.NewResource(o.Name, name+"_"+lastSeg(o.Name), "google_discovery_engine_target_site", g.ProviderName, dsAttrs(lastSeg(o.Name)), discoveryEngineAllowEmptyValues, discoveryEngineAdditionalFields))
				}
				return nil
			}); e != nil {
				log.Println(e)
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	if err := discoveryEngineService.Projects.Locations.Collections.Engines.List(parent).Pages(ctx, func(page *discoveryengine.GoogleCloudDiscoveryengineV1ListEnginesResponse) error {
		for _, e := range page.Engines {
			name := lastSeg(e.Name)
			tfType := "google_discovery_engine_search_engine"
			switch e.SolutionType {
			case "SOLUTION_TYPE_CHAT":
				tfType = "google_discovery_engine_chat_engine"
			case "SOLUTION_TYPE_RECOMMENDATION":
				tfType = "google_discovery_engine_recommendation_engine"
			}
			g.Resources = append(g.Resources, terraformutils.NewResource(
				e.Name, name, tfType, g.ProviderName,
				map[string]string{"engine_id": name, "location": location, "project": project},
				discoveryEngineAllowEmptyValues, discoveryEngineAdditionalFields))
			if ae := discoveryEngineService.Projects.Locations.Collections.Engines.Assistants.List(e.Name).Pages(ctx, func(ap *discoveryengine.GoogleCloudDiscoveryengineV1ListAssistantsResponse) error {
				for _, o := range ap.Assistants {
					g.Resources = append(g.Resources, terraformutils.NewResource(o.Name, name+"_"+lastSeg(o.Name), "google_discovery_engine_assistant", g.ProviderName, map[string]string{"location": location, "project": project}, discoveryEngineAllowEmptyValues, discoveryEngineAdditionalFields))
				}
				return nil
			}); ae != nil {
				log.Println(ae)
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}

func lastSeg(s string) string { p := strings.Split(s, "/"); return p[len(p)-1] }
