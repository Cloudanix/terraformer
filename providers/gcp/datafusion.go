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
	"google.golang.org/api/datafusion/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var datafusionAllowEmptyValues = []string{""}

var datafusionAdditionalFields = map[string]interface{}{}

type DatafusionGenerator struct {
	GCPService
}

// Run on instancesList and create for each TerraformResource
func (g DatafusionGenerator) createResources(ctx context.Context, instancesList *datafusion.ProjectsLocationsInstancesListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	region := g.GetArgs()["region"].(compute.Region).Name
	if err := instancesList.Pages(ctx, func(page *datafusion.ListInstancesResponse) error {
		for _, obj := range page.Instances {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_data_fusion_instance",
				g.ProviderName,
				map[string]string{
					"name":    name,
					"project": g.GetArgs()["project"].(string),
					"region":  region,
				},
				datafusionAllowEmptyValues,
				datafusionAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *DatafusionGenerator) InitResources() error {
	ctx := context.Background()
	datafusionService, err := datafusion.NewService(ctx)
	if err != nil {
		return err
	}

	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	instancesList := datafusionService.Projects.Locations.Instances.List(
		"projects/" + project + "/locations/" + region)
	instanceRes := g.createResources(ctx, instancesList)
	g.Resources = instanceRes
	for _, r := range instanceRes {
		res := r.InstanceState.ID
		short := strings.Split(res, "/")[len(strings.Split(res, "/"))-1]
		if policy, perr := datafusionService.Projects.Locations.Instances.GetIamPolicy(res).Do(); perr == nil {
			for _, b := range policy.Bindings {
				for _, m := range b.Members {
					g.Resources = append(g.Resources, terraformutils.NewResource(
						res+" "+b.Role+" "+m, short+"_"+b.Role+"_"+m,
						"google_data_fusion_instance_iam_member", g.ProviderName,
						map[string]string{"name": short, "role": b.Role, "member": m, "project": project, "region": region},
						datafusionAllowEmptyValues, datafusionAdditionalFields))
				}
			}
		}
	}
	return nil
}
