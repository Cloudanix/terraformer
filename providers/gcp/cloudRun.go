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
	"google.golang.org/api/run/v2"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var cloudRunAllowEmptyValues = []string{""}

var cloudRunAdditionalFields = map[string]interface{}{}

type CloudRunGenerator struct {
	GCPService
}

// Run on servicesList and create for each TerraformResource
func (g CloudRunGenerator) createResources(ctx context.Context, servicesList *run.ProjectsLocationsServicesListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := servicesList.Pages(ctx, func(page *run.GoogleCloudRunV2ListServicesResponse) error {
		for _, obj := range page.Services {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_cloud_run_v2_service",
				g.ProviderName,
				map[string]string{
					"name":     name,
					"project":  g.GetArgs()["project"].(string),
					"location": location,
				},
				cloudRunAllowEmptyValues,
				cloudRunAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Run on jobsList and create for each TerraformResource
func (g CloudRunGenerator) createJobsResources(ctx context.Context, jobsList *run.ProjectsLocationsJobsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := jobsList.Pages(ctx, func(page *run.GoogleCloudRunV2ListJobsResponse) error {
		for _, obj := range page.Jobs {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_cloud_run_v2_job",
				g.ProviderName,
				map[string]string{
					"name":     name,
					"project":  g.GetArgs()["project"].(string),
					"location": location,
				},
				cloudRunAllowEmptyValues,
				cloudRunAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *CloudRunGenerator) InitResources() error {
	ctx := context.Background()
	runService, err := run.NewService(ctx)
	if err != nil {
		return err
	}

	parent := "projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name
	servicesList := runService.Projects.Locations.Services.List(parent)
	g.Resources = append(g.Resources, g.createResources(ctx, servicesList)...)

	jobsList := runService.Projects.Locations.Jobs.List(parent)
	g.Resources = append(g.Resources, g.createJobsResources(ctx, jobsList)...)
	return nil
}
