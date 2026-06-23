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
	"google.golang.org/api/datapipelines/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var dataPipelinesAllowEmptyValues = []string{""}

var dataPipelinesAdditionalFields = map[string]interface{}{}

type DataPipelinesGenerator struct {
	GCPService
}

// Generate TerraformResources from GCP API,
func (g *DataPipelinesGenerator) InitResources() error {
	ctx := context.Background()
	dpService, err := datapipelines.NewService(ctx)
	if err != nil {
		return err
	}
	location := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + g.GetArgs()["project"].(string) + "/locations/" + location

	pipelinesList := dpService.Projects.Locations.Pipelines.List(parent)
	if err := pipelinesList.Pages(ctx, func(page *datapipelines.GoogleCloudDatapipelinesV1ListPipelinesResponse) error {
		for _, obj := range page.Pipelines {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name, name, "google_data_pipeline_pipeline", g.ProviderName,
				map[string]string{"name": name, "project": g.GetArgs()["project"].(string), "location": location},
				dataPipelinesAllowEmptyValues, dataPipelinesAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}
