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

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/dataproc/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var dataprocAllowEmptyValues = []string{""}

var dataprocAdditionalFields = map[string]interface{}{}

type DataprocGenerator struct {
	GCPService
}

// Run on DataprocClusterList and create for each TerraformResource
func (g DataprocGenerator) createClusterResources(ctx context.Context, clusterList *dataproc.ProjectsRegionsClustersListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	if err := clusterList.Pages(ctx, func(page *dataproc.ListClustersResponse) error {
		for _, cluster := range page.Clusters {
			resource := terraformutils.NewResource(
				cluster.ClusterName,
				cluster.ClusterName,
				"google_dataproc_cluster",
				g.ProviderName,
				map[string]string{
					"name":    cluster.ClusterName,
					"project": g.GetArgs()["project"].(string),
					"region":  g.GetArgs()["region"].(compute.Region).Name,
				},
				dataprocAllowEmptyValues,
				dataprocAdditionalFields,
			)
			resource.IgnoreKeys = append(resource.IgnoreKeys, "^cluster_config.[0-9].delete_autogen_bucket$")
			resources = append(resources, resource)
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

/*
// Run on DataprocJobList and create for each TerraformResource
func (g DataprocGenerator) createJobResources(jobList *dataproc.ProjectsRegionsJobsListCall, ctx context.Context) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	if err := jobList.Pages(ctx, func(page *dataproc.ListJobsResponse) error {
		for _, job := range page.Jobs {
			resources = append(resources, terraformutils.NewResource(
				job.Reference.JobId,
				job.Reference.JobId,
				"google_dataproc_job",
				g.ProviderName,
				map[string]string{
					"project": g.GetArgs()["project"].(string),
					"region":  g.GetArgs()["region"].(compute.Region).Name,
				},
				dataprocAllowEmptyValues,
				dataprocAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}
	return resources
}
*/

// Run on autoscalingPolicyList and create for each TerraformResource
func (g DataprocGenerator) createAutoscalingPolicyResources(ctx context.Context, policyList *dataproc.ProjectsRegionsAutoscalingPoliciesListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	if err := policyList.Pages(ctx, func(page *dataproc.ListAutoscalingPoliciesResponse) error {
		for _, policy := range page.Policies {
			resources = append(resources, terraformutils.NewResource(
				policy.Name,
				policy.Id,
				"google_dataproc_autoscaling_policy",
				g.ProviderName,
				map[string]string{
					"policy_id": policy.Id,
					"project":   g.GetArgs()["project"].(string),
					"location":  g.GetArgs()["region"].(compute.Region).Name,
				},
				dataprocAllowEmptyValues,
				dataprocAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Run on workflowTemplateList and create for each TerraformResource
func (g DataprocGenerator) createWorkflowTemplateResources(ctx context.Context, wftList *dataproc.ProjectsRegionsWorkflowTemplatesListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	if err := wftList.Pages(ctx, func(page *dataproc.ListWorkflowTemplatesResponse) error {
		for _, tmpl := range page.Templates {
			resources = append(resources, terraformutils.NewResource(
				tmpl.Name,
				tmpl.Id,
				"google_dataproc_workflow_template",
				g.ProviderName,
				map[string]string{
					"name":     tmpl.Id,
					"project":  g.GetArgs()["project"].(string),
					"location": g.GetArgs()["region"].(compute.Region).Name,
				},
				dataprocAllowEmptyValues,
				dataprocAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
// from each DataprocGenerator create 1 TerraformResource
// Need DataprocGenerator name as ID for terraform resource
func (g *DataprocGenerator) InitResources() error {
	ctx := context.Background()
	dataprocService, err := dataproc.NewService(ctx)
	if err != nil {
		return err
	}

	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name

	clusterList := dataprocService.Projects.Regions.Clusters.List(project, region)
	g.Resources = g.createClusterResources(ctx, clusterList)

	policyList := dataprocService.Projects.Regions.AutoscalingPolicies.List("projects/" + project + "/regions/" + region)
	g.Resources = append(g.Resources, g.createAutoscalingPolicyResources(ctx, policyList)...)

	wftList := dataprocService.Projects.Regions.WorkflowTemplates.List("projects/" + project + "/regions/" + region)
	g.Resources = append(g.Resources, g.createWorkflowTemplateResources(ctx, wftList)...)

	// jobList := dataprocService.Projects.Regions.Jobs.List(g.GetArgs()["project"].(string), g.GetArgs()["region"])
	// g.Resources = append(g.Resources, g.createJobResources(jobList, ctx)...)

	return nil
}
