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

	"google.golang.org/api/backupdr/v1"
	"google.golang.org/api/compute/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var backupdrAllowEmptyValues = []string{""}

var backupdrAdditionalFields = map[string]interface{}{}

type BackupdrGenerator struct {
	GCPService
}

// Run on managementServersList and create for each TerraformResource
func (g BackupdrGenerator) createResources(ctx context.Context, managementServersList *backupdr.ProjectsLocationsManagementServersListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := managementServersList.Pages(ctx, func(page *backupdr.ListManagementServersResponse) error {
		for _, obj := range page.ManagementServers {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_backup_dr_management_server",
				g.ProviderName,
				map[string]string{
					"name":     name,
					"project":  g.GetArgs()["project"].(string),
					"location": location,
				},
				backupdrAllowEmptyValues,
				backupdrAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *BackupdrGenerator) InitResources() error {
	ctx := context.Background()
	backupdrService, err := backupdr.NewService(ctx)
	if err != nil {
		return err
	}

	loc := g.GetArgs()["region"].(compute.Region).Name
	proj := g.GetArgs()["project"].(string)
	parent := "projects/" + proj + "/locations/" + loc
	managementServersList := backupdrService.Projects.Locations.ManagementServers.List(parent)
	g.Resources = append(g.Resources, g.createResources(ctx, managementServersList)...)

	if err := backupdrService.Projects.Locations.BackupPlans.List(parent).Pages(ctx, func(p *backupdr.ListBackupPlansResponse) error {
		for _, o := range p.BackupPlans {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_backup_dr_backup_plan", g.ProviderName,
				map[string]string{"backup_plan_id": name, "project": proj, "location": loc},
				backupdrAllowEmptyValues, backupdrAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := backupdrService.Projects.Locations.BackupVaults.List(parent).Pages(ctx, func(p *backupdr.ListBackupVaultsResponse) error {
		for _, o := range p.BackupVaults {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_backup_dr_backup_vault", g.ProviderName,
				map[string]string{"backup_vault_id": name, "project": proj, "location": loc},
				backupdrAllowEmptyValues, backupdrAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := backupdrService.Projects.Locations.BackupPlanAssociations.List(parent).Pages(ctx, func(p *backupdr.ListBackupPlanAssociationsResponse) error {
		for _, o := range p.BackupPlanAssociations {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_backup_dr_backup_plan_association", g.ProviderName,
				map[string]string{"backup_plan_association_id": name, "project": proj, "location": loc},
				backupdrAllowEmptyValues, backupdrAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}
