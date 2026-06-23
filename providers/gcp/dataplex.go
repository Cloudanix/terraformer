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

	// Walk each lake for its zones.
	for _, lake := range lakeNames {
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
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
	}
	return nil
}
