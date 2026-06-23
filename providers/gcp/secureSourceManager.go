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
	"google.golang.org/api/securesourcemanager/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var secureSourceManagerAllowEmptyValues = []string{""}

var secureSourceManagerAdditionalFields = map[string]interface{}{}

type SecureSourceManagerGenerator struct {
	GCPService
}

// Run on instancesList and create for each TerraformResource
func (g SecureSourceManagerGenerator) createResources(ctx context.Context, list *securesourcemanager.ProjectsLocationsInstancesListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := list.Pages(ctx, func(page *securesourcemanager.ListInstancesResponse) error {
		for _, obj := range page.Instances {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_secure_source_manager_instance",
				g.ProviderName,
				map[string]string{
					"instance_id": name,
					"project":     g.GetArgs()["project"].(string),
					"location":    location,
				},
				secureSourceManagerAllowEmptyValues,
				secureSourceManagerAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *SecureSourceManagerGenerator) InitResources() error {
	ctx := context.Background()
	ssmService, err := securesourcemanager.NewService(ctx)
	if err != nil {
		return err
	}

	project := g.GetArgs()["project"].(string)
	location := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + location
	instancesList := ssmService.Projects.Locations.Instances.List(parent)
	instanceRes := g.createResources(ctx, instancesList)
	g.Resources = append(g.Resources, instanceRes...)
	for _, r := range instanceRes {
		res := r.InstanceState.ID
		short := strings.Split(res, "/")[len(strings.Split(res, "/"))-1]
		if policy, perr := ssmService.Projects.Locations.Instances.GetIamPolicy(res).Do(); perr == nil {
			for _, b := range policy.Bindings {
				for _, m := range b.Members {
					g.Resources = append(g.Resources, terraformutils.NewResource(
						res+" "+b.Role+" "+m, short+"_"+b.Role+"_"+m,
						"google_secure_source_manager_instance_iam_member", g.ProviderName,
						map[string]string{"instance_id": short, "role": b.Role, "member": m, "project": project, "location": location},
						secureSourceManagerAllowEmptyValues, secureSourceManagerAdditionalFields))
				}
			}
		}
	}

	if err := ssmService.Projects.Locations.Repositories.List(parent).Pages(ctx, func(p *securesourcemanager.ListRepositoriesResponse) error {
		for _, o := range p.Repositories {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_secure_source_manager_repository", g.ProviderName,
				map[string]string{"repository_id": name, "project": g.GetArgs()["project"].(string), "location": g.GetArgs()["region"].(compute.Region).Name},
				secureSourceManagerAllowEmptyValues, secureSourceManagerAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}
