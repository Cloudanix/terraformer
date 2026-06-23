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

	project := g.GetArgs()["project"].(string)
	location := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + location
	clustersList := managedKafkaService.Projects.Locations.Clusters.List(parent)
	clusterResources := g.createResources(ctx, clustersList)
	g.Resources = clusterResources
	for _, cr := range clusterResources {
		clFull := cr.InstanceState.ID
		clName := strings.Split(clFull, "/")[len(strings.Split(clFull, "/"))-1]
		if terr := managedKafkaService.Projects.Locations.Clusters.Topics.List(clFull).Pages(ctx, func(p *managedkafka.ListTopicsResponse) error {
			for _, o := range p.Topics {
				t := strings.Split(o.Name, "/")
				g.Resources = append(g.Resources, terraformutils.NewResource(
					o.Name, t[len(t)-1], "google_managed_kafka_topic", g.ProviderName,
					map[string]string{"topic_id": t[len(t)-1], "cluster": clName, "location": location, "project": project},
					managedKafkaAllowEmptyValues, managedKafkaAdditionalFields))
			}
			return nil
		}); terr != nil {
			log.Println(terr)
		}
		if aerr := managedKafkaService.Projects.Locations.Clusters.Acls.List(clFull).Pages(ctx, func(p *managedkafka.ListAclsResponse) error {
			for _, o := range p.Acls {
				t := strings.Split(o.Name, "/")
				g.Resources = append(g.Resources, terraformutils.NewResource(
					o.Name, t[len(t)-1], "google_managed_kafka_acl", g.ProviderName,
					map[string]string{"acl_id": t[len(t)-1], "cluster": clName, "location": location, "project": project},
					managedKafkaAllowEmptyValues, managedKafkaAdditionalFields))
			}
			return nil
		}); aerr != nil {
			log.Println(aerr)
		}
	}

	if err := managedKafkaService.Projects.Locations.ConnectClusters.List(parent).Pages(ctx, func(p *managedkafka.ListConnectClustersResponse) error {
		for _, o := range p.ConnectClusters {
			t := strings.Split(o.Name, "/")
			ccName := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, ccName, "google_managed_kafka_connect_cluster", g.ProviderName,
				map[string]string{"connect_cluster_id": ccName, "location": location, "project": project},
				managedKafkaAllowEmptyValues, managedKafkaAdditionalFields))
			if cerr := managedKafkaService.Projects.Locations.ConnectClusters.Connectors.List(o.Name).Pages(ctx, func(cp *managedkafka.ListConnectorsResponse) error {
				for _, c := range cp.Connectors {
					ct := strings.Split(c.Name, "/")
					g.Resources = append(g.Resources, terraformutils.NewResource(
						c.Name, ct[len(ct)-1], "google_managed_kafka_connector", g.ProviderName,
						map[string]string{"connector_id": ct[len(ct)-1], "connect_cluster": ccName, "location": location, "project": project},
						managedKafkaAllowEmptyValues, managedKafkaAdditionalFields))
				}
				return nil
			}); cerr != nil {
				log.Println(cerr)
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}
