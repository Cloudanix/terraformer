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
	"google.golang.org/api/workstations/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var workstationsAllowEmptyValues = []string{""}

var workstationsAdditionalFields = map[string]interface{}{}

type WorkstationsGenerator struct {
	GCPService
}

// Run on configsList and create for each TerraformResource (per-cluster walk)
func (g WorkstationsGenerator) createConfigResources(ctx context.Context, list *workstations.ProjectsLocationsWorkstationClustersWorkstationConfigsListCall, clusterID string) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := list.Pages(ctx, func(page *workstations.ListWorkstationConfigsResponse) error {
		for _, obj := range page.WorkstationConfigs {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_workstations_workstation_config",
				g.ProviderName,
				map[string]string{
					"workstation_config_id":  name,
					"workstation_cluster_id": clusterID,
					"project":                g.GetArgs()["project"].(string),
					"location":               location,
				},
				workstationsAllowEmptyValues,
				workstationsAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *WorkstationsGenerator) InitResources() error {
	ctx := context.Background()
	workstationsService, err := workstations.NewService(ctx)
	if err != nil {
		return err
	}
	project := g.GetArgs()["project"].(string)
	location := g.GetArgs()["region"].(compute.Region).Name

	clusterIDs := []string{}
	clustersList := workstationsService.Projects.Locations.WorkstationClusters.List("projects/" + project + "/locations/" + location)
	if err := clustersList.Pages(ctx, func(page *workstations.ListWorkstationClustersResponse) error {
		for _, obj := range page.WorkstationClusters {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			clusterIDs = append(clusterIDs, name)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_workstations_workstation_cluster",
				g.ProviderName,
				map[string]string{"name": name, "project": project, "location": location},
				workstationsAllowEmptyValues,
				workstationsAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	for _, cluster := range clusterIDs {
		configsList := workstationsService.Projects.Locations.WorkstationClusters.WorkstationConfigs.List(
			"projects/" + project + "/locations/" + location + "/workstationClusters/" + cluster)
		g.Resources = append(g.Resources, g.createConfigResources(ctx, configsList, cluster)...)
	}
	return nil
}
