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
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}
