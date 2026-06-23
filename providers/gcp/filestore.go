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
	"google.golang.org/api/file/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var filestoreAllowEmptyValues = []string{""}

var filestoreAdditionalFields = map[string]interface{}{}

type FilestoreGenerator struct {
	GCPService
}

// Run on instancesList and create for each TerraformResource
func (g FilestoreGenerator) createResources(ctx context.Context, instancesList *file.ProjectsLocationsInstancesListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := instancesList.Pages(ctx, func(page *file.ListInstancesResponse) error {
		for _, obj := range page.Instances {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_filestore_instance",
				g.ProviderName,
				map[string]string{
					"name":     name,
					"project":  g.GetArgs()["project"].(string),
					"location": location,
				},
				filestoreAllowEmptyValues,
				filestoreAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *FilestoreGenerator) InitResources() error {
	ctx := context.Background()
	filestoreService, err := file.NewService(ctx)
	if err != nil {
		return err
	}

	parent := "projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name
	instancesList := filestoreService.Projects.Locations.Instances.List(parent)
	g.Resources = append(g.Resources, g.createResources(ctx, instancesList)...)

	backupsList := filestoreService.Projects.Locations.Backups.List(parent)
	g.Resources = append(g.Resources, g.createBackupsResources(ctx, backupsList)...)

	// Walk each instance for its snapshots.
	for _, r := range g.Resources {
		if r.InstanceInfo.Type != "google_filestore_instance" {
			continue
		}
		instanceName, ok := r.Item["name"].(string)
		if !ok {
			continue
		}
		instancePath := parent + "/instances/" + instanceName
		snapsList := filestoreService.Projects.Locations.Instances.Snapshots.List(instancePath)
		if err := snapsList.Pages(ctx, func(page *file.ListSnapshotsResponse) error {
			for _, obj := range page.Snapshots {
				t := strings.Split(obj.Name, "/")
				name := t[len(t)-1]
				g.Resources = append(g.Resources, terraformutils.NewResource(
					obj.Name, name, "google_filestore_snapshot", g.ProviderName,
					map[string]string{
						"name":     name,
						"instance": instanceName,
						"project":  g.GetArgs()["project"].(string),
						"location": g.GetArgs()["region"].(compute.Region).Name,
					},
					filestoreAllowEmptyValues, filestoreAdditionalFields,
				))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
	}
	return nil
}

// Run on backupsList and create for each TerraformResource
func (g FilestoreGenerator) createBackupsResources(ctx context.Context, list *file.ProjectsLocationsBackupsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := list.Pages(ctx, func(page *file.ListBackupsResponse) error {
		for _, obj := range page.Backups {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_filestore_backup",
				g.ProviderName,
				map[string]string{
					"name":     name,
					"project":  g.GetArgs()["project"].(string),
					"location": location,
				},
				filestoreAllowEmptyValues,
				filestoreAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}
