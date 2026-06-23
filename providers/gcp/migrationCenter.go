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
	"google.golang.org/api/migrationcenter/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var migrationCenterAllowEmptyValues = []string{""}

var migrationCenterAdditionalFields = map[string]interface{}{}

type MigrationCenterGenerator struct {
	GCPService
}

// Run on groupsList and create for each TerraformResource
func (g MigrationCenterGenerator) createResources(ctx context.Context, list *migrationcenter.ProjectsLocationsGroupsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := list.Pages(ctx, func(page *migrationcenter.ListGroupsResponse) error {
		for _, obj := range page.Groups {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_migration_center_group",
				g.ProviderName,
				map[string]string{
					"group_id": name,
					"project":  g.GetArgs()["project"].(string),
					"location": location,
				},
				migrationCenterAllowEmptyValues,
				migrationCenterAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *MigrationCenterGenerator) InitResources() error {
	ctx := context.Background()
	mcService, err := migrationcenter.NewService(ctx)
	if err != nil {
		return err
	}

	parent := "projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name
	groupsList := mcService.Projects.Locations.Groups.List(parent)
	g.Resources = append(g.Resources, g.createResources(ctx, groupsList)...)

	project := g.GetArgs()["project"].(string)
	loc := g.GetArgs()["region"].(compute.Region).Name
	tail := func(s string) string { p := strings.Split(s, "/"); return p[len(p)-1] }
	if err := mcService.Projects.Locations.Sources.List(parent).Pages(ctx, func(p *migrationcenter.ListSourcesResponse) error {
		for _, o := range p.Sources {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, tail(o.Name), "google_migration_center_source", g.ProviderName,
				map[string]string{"source_id": tail(o.Name), "project": project, "location": loc},
				migrationCenterAllowEmptyValues, migrationCenterAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := mcService.Projects.Locations.PreferenceSets.List(parent).Pages(ctx, func(p *migrationcenter.ListPreferenceSetsResponse) error {
		for _, o := range p.PreferenceSets {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, tail(o.Name), "google_migration_center_preference_set", g.ProviderName,
				map[string]string{"preference_set_id": tail(o.Name), "project": project, "location": loc},
				migrationCenterAllowEmptyValues, migrationCenterAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := mcService.Projects.Locations.DiscoveryClients.List(parent).Pages(ctx, func(p *migrationcenter.ListDiscoveryClientsResponse) error {
		for _, o := range p.DiscoveryClients {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, tail(o.Name), "google_migration_center_discovery_client", g.ProviderName,
				map[string]string{"discovery_client_id": tail(o.Name), "project": project, "location": loc},
				migrationCenterAllowEmptyValues, migrationCenterAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := mcService.Projects.Locations.ImportJobs.List(parent).Pages(ctx, func(p *migrationcenter.ListImportJobsResponse) error {
		for _, o := range p.ImportJobs {
			job := tail(o.Name)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, job, "google_migration_center_import_job", g.ProviderName,
				map[string]string{"import_job_id": job, "project": project, "location": loc},
				migrationCenterAllowEmptyValues, migrationCenterAdditionalFields))
			if e := mcService.Projects.Locations.ImportJobs.ImportDataFiles.List(o.Name).Pages(ctx, func(fp *migrationcenter.ListImportDataFilesResponse) error {
				for _, f := range fp.ImportDataFiles {
					g.Resources = append(g.Resources, terraformutils.NewResource(
						f.Name, job+"_"+tail(f.Name), "google_migration_center_import_data_file", g.ProviderName,
						map[string]string{"import_data_file_id": tail(f.Name), "import_job": job, "project": project, "location": loc},
						migrationCenterAllowEmptyValues, migrationCenterAdditionalFields))
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
	if err := mcService.Projects.Locations.ReportConfigs.List(parent).Pages(ctx, func(p *migrationcenter.ListReportConfigsResponse) error {
		for _, o := range p.ReportConfigs {
			cfg := tail(o.Name)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, cfg, "google_migration_center_report_config", g.ProviderName,
				map[string]string{"report_config_id": cfg, "project": project, "location": loc},
				migrationCenterAllowEmptyValues, migrationCenterAdditionalFields))
			if e := mcService.Projects.Locations.ReportConfigs.Reports.List(o.Name).Pages(ctx, func(rp *migrationcenter.ListReportsResponse) error {
				for _, r := range rp.Reports {
					g.Resources = append(g.Resources, terraformutils.NewResource(
						r.Name, cfg+"_"+tail(r.Name), "google_migration_center_report", g.ProviderName,
						map[string]string{"report_id": tail(r.Name), "report_config": cfg, "project": project, "location": loc},
						migrationCenterAllowEmptyValues, migrationCenterAdditionalFields))
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
	return nil
}
