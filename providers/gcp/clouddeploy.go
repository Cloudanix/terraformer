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

	"google.golang.org/api/clouddeploy/v1"
	"google.golang.org/api/compute/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var clouddeployAllowEmptyValues = []string{""}

var clouddeployAdditionalFields = map[string]interface{}{}

type ClouddeployGenerator struct {
	GCPService
}

// Run on deliveryPipelinesList and create for each TerraformResource
func (g ClouddeployGenerator) createResources(ctx context.Context, deliveryPipelinesList *clouddeploy.ProjectsLocationsDeliveryPipelinesListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := deliveryPipelinesList.Pages(ctx, func(page *clouddeploy.ListDeliveryPipelinesResponse) error {
		for _, obj := range page.DeliveryPipelines {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_clouddeploy_delivery_pipeline",
				g.ProviderName,
				map[string]string{
					"name":     name,
					"project":  g.GetArgs()["project"].(string),
					"location": location,
				},
				clouddeployAllowEmptyValues,
				clouddeployAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *ClouddeployGenerator) InitResources() error {
	ctx := context.Background()
	clouddeployService, err := clouddeploy.NewService(ctx)
	if err != nil {
		return err
	}

	parent := "projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name
	loc := g.GetArgs()["region"].(compute.Region).Name
	proj := g.GetArgs()["project"].(string)

	deliveryPipelinesList := clouddeployService.Projects.Locations.DeliveryPipelines.List(parent)
	pipelineRes := g.createResources(ctx, deliveryPipelinesList)
	g.Resources = append(g.Resources, pipelineRes...)
	for _, r := range pipelineRes {
		res := r.InstanceState.ID
		if policy, perr := clouddeployService.Projects.Locations.DeliveryPipelines.GetIamPolicy(res).Do(); perr == nil {
			short := strings.Split(res, "/")[len(strings.Split(res, "/"))-1]
			for _, b := range policy.Bindings {
				for _, m := range b.Members {
					g.Resources = append(g.Resources, terraformutils.NewResource(
						res+" "+b.Role+" "+m, short+"_"+b.Role+"_"+m,
						"google_clouddeploy_delivery_pipeline_iam_member", g.ProviderName,
						map[string]string{"name": short, "role": b.Role, "member": m, "project": proj, "location": loc},
						clouddeployAllowEmptyValues, clouddeployAdditionalFields))
				}
			}
		}
	}

	targetsList := clouddeployService.Projects.Locations.Targets.List(parent)
	g.Resources = append(g.Resources, g.createTargetsResources(ctx, targetsList)...)

	if err := clouddeployService.Projects.Locations.CustomTargetTypes.List(parent).Pages(ctx, func(p *clouddeploy.ListCustomTargetTypesResponse) error {
		for _, o := range p.CustomTargetTypes {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_clouddeploy_custom_target_type", g.ProviderName,
				map[string]string{"name": name, "project": proj, "location": loc},
				clouddeployAllowEmptyValues, clouddeployAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := clouddeployService.Projects.Locations.DeployPolicies.List(parent).Pages(ctx, func(p *clouddeploy.ListDeployPoliciesResponse) error {
		for _, o := range p.DeployPolicies {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_clouddeploy_deploy_policy", g.ProviderName,
				map[string]string{"name": name, "project": proj, "location": loc},
				clouddeployAllowEmptyValues, clouddeployAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}

// Run on targetsList and create for each TerraformResource
func (g ClouddeployGenerator) createTargetsResources(ctx context.Context, list *clouddeploy.ProjectsLocationsTargetsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := list.Pages(ctx, func(page *clouddeploy.ListTargetsResponse) error {
		for _, obj := range page.Targets {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_clouddeploy_target",
				g.ProviderName,
				map[string]string{
					"name":     name,
					"project":  g.GetArgs()["project"].(string),
					"location": location,
				},
				clouddeployAllowEmptyValues,
				clouddeployAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}
