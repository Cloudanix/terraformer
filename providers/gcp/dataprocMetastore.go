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
	"google.golang.org/api/metastore/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var dataprocMetastoreAllowEmptyValues = []string{""}

var dataprocMetastoreAdditionalFields = map[string]interface{}{}

type DataprocMetastoreGenerator struct {
	GCPService
}

// Run on servicesList and create for each TerraformResource
func (g DataprocMetastoreGenerator) createResources(ctx context.Context, servicesList *metastore.ProjectsLocationsServicesListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := servicesList.Pages(ctx, func(page *metastore.ListServicesResponse) error {
		for _, obj := range page.Services {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_dataproc_metastore_service",
				g.ProviderName,
				map[string]string{
					"name":       name,
					"service_id": name,
					"project":    g.GetArgs()["project"].(string),
					"location":   location,
				},
				dataprocMetastoreAllowEmptyValues,
				dataprocMetastoreAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *DataprocMetastoreGenerator) InitResources() error {
	ctx := context.Background()
	metastoreService, err := metastore.NewService(ctx)
	if err != nil {
		return err
	}

	parent := "projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name
	servicesList := metastoreService.Projects.Locations.Services.List(parent)
	g.Resources = append(g.Resources, g.createResources(ctx, servicesList)...)

	if err := metastoreService.Projects.Locations.Federations.List(parent).Pages(ctx, func(p *metastore.ListFederationsResponse) error {
		for _, o := range p.Federations {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_dataproc_metastore_federation", g.ProviderName,
				map[string]string{"federation_id": name, "project": g.GetArgs()["project"].(string), "location": g.GetArgs()["region"].(compute.Region).Name},
				dataprocMetastoreAllowEmptyValues, dataprocMetastoreAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}
