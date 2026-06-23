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

	"google.golang.org/api/alloydb/v1"
	"google.golang.org/api/compute/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var alloydbAllowEmptyValues = []string{""}

var alloydbAdditionalFields = map[string]interface{}{}

type AlloydbGenerator struct {
	GCPService
}

// Generate TerraformResources from GCP API,
func (g *AlloydbGenerator) InitResources() error {
	ctx := context.Background()
	alloydbService, err := alloydb.NewService(ctx)
	if err != nil {
		return err
	}
	project := g.GetArgs()["project"].(string)
	location := g.GetArgs()["region"].(compute.Region).Name

	clusterNames := []string{}
	clustersList := alloydbService.Projects.Locations.Clusters.List("projects/" + project + "/locations/" + location)
	if err := clustersList.Pages(ctx, func(page *alloydb.ListClustersResponse) error {
		for _, obj := range page.Clusters {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			clusterNames = append(clusterNames, name)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name, name, "google_alloydb_cluster", g.ProviderName,
				map[string]string{"cluster_id": name, "project": project, "location": location},
				alloydbAllowEmptyValues, alloydbAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	backupsList := alloydbService.Projects.Locations.Backups.List("projects/" + project + "/locations/" + location)
	if err := backupsList.Pages(ctx, func(page *alloydb.ListBackupsResponse) error {
		for _, obj := range page.Backups {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name, name, "google_alloydb_backup", g.ProviderName,
				map[string]string{"backup_id": name, "project": project, "location": location},
				alloydbAllowEmptyValues, alloydbAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	for _, cluster := range clusterNames {
		instList := alloydbService.Projects.Locations.Clusters.Instances.List(
			"projects/" + project + "/locations/" + location + "/clusters/" + cluster)
		if err := instList.Pages(ctx, func(page *alloydb.ListInstancesResponse) error {
			for _, obj := range page.Instances {
				t := strings.Split(obj.Name, "/")
				name := t[len(t)-1]
				g.Resources = append(g.Resources, terraformutils.NewResource(
					obj.Name, name, "google_alloydb_instance", g.ProviderName,
					map[string]string{"instance_id": name, "cluster": cluster, "project": project, "location": location},
					alloydbAllowEmptyValues, alloydbAdditionalFields,
				))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
	}
	return nil
}
