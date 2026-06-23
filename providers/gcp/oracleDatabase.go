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
	"google.golang.org/api/oracledatabase/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var oracleDatabaseAllowEmptyValues = []string{""}

var oracleDatabaseAdditionalFields = map[string]interface{}{}

type OracleDatabaseGenerator struct {
	GCPService
}

// Run on infraList and create for each TerraformResource
func (g OracleDatabaseGenerator) createResources(ctx context.Context, infraList *oracledatabase.ProjectsLocationsCloudExadataInfrastructuresListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := infraList.Pages(ctx, func(page *oracledatabase.ListCloudExadataInfrastructuresResponse) error {
		for _, obj := range page.CloudExadataInfrastructures {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_oracle_database_cloud_exadata_infrastructure",
				g.ProviderName,
				map[string]string{
					"name":     name,
					"project":  g.GetArgs()["project"].(string),
					"location": location,
				},
				oracleDatabaseAllowEmptyValues,
				oracleDatabaseAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *OracleDatabaseGenerator) InitResources() error {
	ctx := context.Background()
	oracleDatabaseService, err := oracledatabase.NewService(ctx)
	if err != nil {
		return err
	}

	parent := "projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name
	infraList := oracleDatabaseService.Projects.Locations.CloudExadataInfrastructures.List(parent)
	g.Resources = append(g.Resources, g.createResources(ctx, infraList)...)

	loc := g.GetArgs()["region"].(compute.Region).Name
	proj := g.GetArgs()["project"].(string)
	if err := oracleDatabaseService.Projects.Locations.AutonomousDatabases.List(parent).Pages(ctx, func(p *oracledatabase.ListAutonomousDatabasesResponse) error {
		for _, o := range p.AutonomousDatabases {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_oracle_database_autonomous_database", g.ProviderName,
				map[string]string{"autonomous_database_id": name, "project": proj, "location": loc},
				oracleDatabaseAllowEmptyValues, oracleDatabaseAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := oracleDatabaseService.Projects.Locations.CloudVmClusters.List(parent).Pages(ctx, func(p *oracledatabase.ListCloudVmClustersResponse) error {
		for _, o := range p.CloudVmClusters {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_oracle_database_cloud_vm_cluster", g.ProviderName,
				map[string]string{"cloud_vm_cluster_id": name, "project": proj, "location": loc},
				oracleDatabaseAllowEmptyValues, oracleDatabaseAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := oracleDatabaseService.Projects.Locations.DbSystems.List(parent).Pages(ctx, func(p *oracledatabase.ListDbSystemsResponse) error {
		for _, o := range p.DbSystems {
			t := strings.Split(o.Name, "/")
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, t[len(t)-1], "google_oracle_database_db_system", g.ProviderName,
				map[string]string{"db_system_id": t[len(t)-1], "project": proj, "location": loc},
				oracleDatabaseAllowEmptyValues, oracleDatabaseAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := oracleDatabaseService.Projects.Locations.ExadbVmClusters.List(parent).Pages(ctx, func(p *oracledatabase.ListExadbVmClustersResponse) error {
		for _, o := range p.ExadbVmClusters {
			t := strings.Split(o.Name, "/")
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, t[len(t)-1], "google_oracle_database_exadb_vm_cluster", g.ProviderName,
				map[string]string{"exadb_vm_cluster_id": t[len(t)-1], "project": proj, "location": loc},
				oracleDatabaseAllowEmptyValues, oracleDatabaseAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := oracleDatabaseService.Projects.Locations.ExascaleDbStorageVaults.List(parent).Pages(ctx, func(p *oracledatabase.ListExascaleDbStorageVaultsResponse) error {
		for _, o := range p.ExascaleDbStorageVaults {
			t := strings.Split(o.Name, "/")
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, t[len(t)-1], "google_oracle_database_exascale_db_storage_vault", g.ProviderName,
				map[string]string{"exascale_db_storage_vault_id": t[len(t)-1], "project": proj, "location": loc},
				oracleDatabaseAllowEmptyValues, oracleDatabaseAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := oracleDatabaseService.Projects.Locations.GoldengateConnections.List(parent).Pages(ctx, func(p *oracledatabase.ListGoldengateConnectionsResponse) error {
		for _, o := range p.GoldengateConnections {
			t := strings.Split(o.Name, "/")
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, t[len(t)-1], "google_oracle_database_goldengate_connection", g.ProviderName,
				map[string]string{"goldengate_connection_id": t[len(t)-1], "project": proj, "location": loc},
				oracleDatabaseAllowEmptyValues, oracleDatabaseAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := oracleDatabaseService.Projects.Locations.GoldengateConnectionAssignments.List(parent).Pages(ctx, func(p *oracledatabase.ListGoldengateConnectionAssignmentsResponse) error {
		for _, o := range p.GoldengateConnectionAssignments {
			t := strings.Split(o.Name, "/")
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, t[len(t)-1], "google_oracle_database_goldengate_connection_assignment", g.ProviderName,
				map[string]string{"goldengate_connection_assignment_id": t[len(t)-1], "project": proj, "location": loc},
				oracleDatabaseAllowEmptyValues, oracleDatabaseAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := oracleDatabaseService.Projects.Locations.GoldengateDeployments.List(parent).Pages(ctx, func(p *oracledatabase.ListGoldengateDeploymentsResponse) error {
		for _, o := range p.GoldengateDeployments {
			t := strings.Split(o.Name, "/")
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, t[len(t)-1], "google_oracle_database_goldengate_deployment", g.ProviderName,
				map[string]string{"goldengate_deployment_id": t[len(t)-1], "project": proj, "location": loc},
				oracleDatabaseAllowEmptyValues, oracleDatabaseAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := oracleDatabaseService.Projects.Locations.OdbNetworks.List(parent).Pages(ctx, func(p *oracledatabase.ListOdbNetworksResponse) error {
		for _, o := range p.OdbNetworks {
			t := strings.Split(o.Name, "/")
			odbName := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, odbName, "google_oracle_database_odb_network", g.ProviderName,
				map[string]string{"odb_network_id": odbName, "project": proj, "location": loc},
				oracleDatabaseAllowEmptyValues, oracleDatabaseAdditionalFields))
			if serr := oracleDatabaseService.Projects.Locations.OdbNetworks.OdbSubnets.List(o.Name).Pages(ctx, func(sp *oracledatabase.ListOdbSubnetsResponse) error {
				for _, s := range sp.OdbSubnets {
					st := strings.Split(s.Name, "/")
					g.Resources = append(g.Resources, terraformutils.NewResource(
						s.Name, odbName+"_"+st[len(st)-1], "google_oracle_database_odb_subnet", g.ProviderName,
						map[string]string{"odb_subnet_id": st[len(st)-1], "odb_network": odbName, "project": proj, "location": loc},
						oracleDatabaseAllowEmptyValues, oracleDatabaseAdditionalFields))
				}
				return nil
			}); serr != nil {
				log.Println(serr)
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}
