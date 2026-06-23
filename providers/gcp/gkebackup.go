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
	"google.golang.org/api/gkebackup/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var gkebackupAllowEmptyValues = []string{""}

var gkebackupAdditionalFields = map[string]interface{}{}

type GkebackupGenerator struct {
	GCPService
}

// Run on backupPlansList and create for each TerraformResource
func (g GkebackupGenerator) createResources(ctx context.Context, list *gkebackup.ProjectsLocationsBackupPlansListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := list.Pages(ctx, func(page *gkebackup.ListBackupPlansResponse) error {
		for _, obj := range page.BackupPlans {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_gke_backup_backup_plan",
				g.ProviderName,
				map[string]string{
					"name":     name,
					"project":  g.GetArgs()["project"].(string),
					"location": location,
				},
				gkebackupAllowEmptyValues,
				gkebackupAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *GkebackupGenerator) InitResources() error {
	ctx := context.Background()
	gkebackupService, err := gkebackup.NewService(ctx)
	if err != nil {
		return err
	}

	parent := "projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name
	backupPlansList := gkebackupService.Projects.Locations.BackupPlans.List(parent)
	g.Resources = append(g.Resources, g.createResources(ctx, backupPlansList)...)

	restorePlansList := gkebackupService.Projects.Locations.RestorePlans.List(parent)
	g.Resources = append(g.Resources, g.createRestorePlansResources(ctx, restorePlansList)...)

	channelsList := gkebackupService.Projects.Locations.BackupChannels.List(parent)
	if err := channelsList.Pages(ctx, func(page *gkebackup.ListBackupChannelsResponse) error {
		for _, obj := range page.BackupChannels {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name, name, "google_gke_backup_backup_channel", g.ProviderName,
				map[string]string{"name": name, "project": g.GetArgs()["project"].(string), "location": g.GetArgs()["region"].(compute.Region).Name},
				gkebackupAllowEmptyValues, gkebackupAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}

// Run on restorePlansList and create for each TerraformResource
func (g GkebackupGenerator) createRestorePlansResources(ctx context.Context, list *gkebackup.ProjectsLocationsRestorePlansListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := list.Pages(ctx, func(page *gkebackup.ListRestorePlansResponse) error {
		for _, obj := range page.RestorePlans {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name, name, "google_gke_backup_restore_plan", g.ProviderName,
				map[string]string{"name": name, "project": g.GetArgs()["project"].(string), "location": location},
				gkebackupAllowEmptyValues, gkebackupAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}
