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

// Run on batchesList and create for each TerraformResource
func (g DataprocGenerator) createBatchResources(ctx context.Context, list *dataproc.ProjectsLocationsBatchesListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	if err := list.Pages(ctx, func(page *dataproc.ListBatchesResponse) error {
		for _, obj := range page.Batches {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_dataproc_batch",
				g.ProviderName,
				map[string]string{
					"batch_id": name,
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
	clusterResources := g.createClusterResources(ctx, clusterList)
	g.Resources = clusterResources
	for _, r := range clusterResources {
		cluster := r.InstanceState.ID
		clusterPath := "projects/" + project + "/regions/" + region + "/clusters/" + cluster
		if policy, perr := dataprocService.Projects.Regions.Clusters.GetIamPolicy(clusterPath, &dataproc.GetIamPolicyRequest{}).Do(); perr == nil {
			for _, b := range policy.Bindings {
				for _, m := range b.Members {
					g.Resources = append(g.Resources, terraformutils.NewResource(
						clusterPath+" "+b.Role+" "+m, cluster+"_"+b.Role+"_"+m,
						"google_dataproc_cluster_iam_member", g.ProviderName,
						map[string]string{"cluster": cluster, "role": b.Role, "member": m, "project": project, "region": region},
						dataprocAllowEmptyValues, dataprocAdditionalFields))
				}
			}
		}
	}

	policyList := dataprocService.Projects.Regions.AutoscalingPolicies.List("projects/" + project + "/regions/" + region)
	autoscalingResources := g.createAutoscalingPolicyResources(ctx, policyList)
	g.Resources = append(g.Resources, autoscalingResources...)
	for _, r := range autoscalingResources {
		res := r.InstanceState.ID
		short := strings.Split(res, "/")[len(strings.Split(res, "/"))-1]
		if policy, perr := dataprocService.Projects.Regions.AutoscalingPolicies.GetIamPolicy(res, &dataproc.GetIamPolicyRequest{}).Do(); perr == nil {
			for _, b := range policy.Bindings {
				for _, m := range b.Members {
					g.Resources = append(g.Resources, terraformutils.NewResource(
						res+" "+b.Role+" "+m, short+"_"+b.Role+"_"+m,
						"google_dataproc_autoscaling_policy_iam_member", g.ProviderName,
						map[string]string{"policy_id": short, "role": b.Role, "member": m, "project": project, "location": region},
						dataprocAllowEmptyValues, dataprocAdditionalFields))
				}
			}
		}
	}

	wftList := dataprocService.Projects.Regions.WorkflowTemplates.List("projects/" + project + "/regions/" + region)
	g.Resources = append(g.Resources, g.createWorkflowTemplateResources(ctx, wftList)...)

	batchesList := dataprocService.Projects.Locations.Batches.List("projects/" + project + "/locations/" + region)
	g.Resources = append(g.Resources, g.createBatchResources(ctx, batchesList)...)

	if err := dataprocService.Projects.Locations.SessionTemplates.List("projects/"+project+"/locations/"+region).Pages(ctx, func(p *dataproc.ListSessionTemplatesResponse) error {
		for _, o := range p.SessionTemplates {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_dataproc_session_template", g.ProviderName,
				map[string]string{"name": name, "project": project, "location": region},
				dataprocAllowEmptyValues, dataprocAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	// jobList := dataprocService.Projects.Regions.Jobs.List(g.GetArgs()["project"].(string), g.GetArgs()["region"])
	// g.Resources = append(g.Resources, g.createJobResources(jobList, ctx)...)

	return nil
}
