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
	serviceResources := g.createResources(ctx, servicesList)
	g.Resources = append(g.Resources, serviceResources...)

	// Per-service IAM (member form).
	for _, r := range serviceResources {
		res := r.InstanceState.ID
		policy, perr := runService.Projects.Locations.Services.GetIamPolicy(res).Do()
		if perr != nil {
			continue
		}
		for _, b := range policy.Bindings {
			for _, m := range b.Members {
				g.Resources = append(g.Resources, terraformutils.NewResource(
					res+" "+b.Role+" "+m,
					res+"_"+b.Role+"_"+m,
					"google_cloud_run_v2_service_iam_member",
					g.ProviderName,
					map[string]string{
						"name":     strings.Split(res, "/")[len(strings.Split(res, "/"))-1],
						"role":     b.Role,
						"member":   m,
						"project":  g.GetArgs()["project"].(string),
						"location": g.GetArgs()["region"].(compute.Region).Name,
					},
					cloudRunAllowEmptyValues,
					cloudRunAdditionalFields,
				))
			}
		}
	}

	jobsList := runService.Projects.Locations.Jobs.List(parent)
	g.Resources = append(g.Resources, g.createJobsResources(ctx, jobsList)...)

	workerPoolsList := runService.Projects.Locations.WorkerPools.List(parent)
	g.Resources = append(g.Resources, g.createWorkerPoolsResources(ctx, workerPoolsList)...)
	return nil
}

// Run on workerPoolsList and create for each TerraformResource
func (g CloudRunGenerator) createWorkerPoolsResources(ctx context.Context, list *run.ProjectsLocationsWorkerPoolsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := list.Pages(ctx, func(page *run.GoogleCloudRunV2ListWorkerPoolsResponse) error {
		for _, obj := range page.WorkerPools {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_cloud_run_v2_worker_pool",
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
