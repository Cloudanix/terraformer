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
	"google.golang.org/api/managedkafka/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var managedKafkaAllowEmptyValues = []string{""}

var managedKafkaAdditionalFields = map[string]interface{}{}

type ManagedKafkaGenerator struct {
	GCPService
}

// Run on clustersList and create for each TerraformResource
func (g ManagedKafkaGenerator) createResources(ctx context.Context, clustersList *managedkafka.ProjectsLocationsClustersListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := clustersList.Pages(ctx, func(page *managedkafka.ListClustersResponse) error {
		for _, obj := range page.Clusters {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_managed_kafka_cluster",
				g.ProviderName,
				map[string]string{
					"name":       name,
					"cluster_id": name,
					"project":    g.GetArgs()["project"].(string),
					"location":   location,
				},
				managedKafkaAllowEmptyValues,
				managedKafkaAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *ManagedKafkaGenerator) InitResources() error {
	ctx := context.Background()
	managedKafkaService, err := managedkafka.NewService(ctx)
	if err != nil {
		return err
	}

	clustersList := managedKafkaService.Projects.Locations.Clusters.List(
		"projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name)
	g.Resources = g.createResources(ctx, clustersList)
	return nil
}
