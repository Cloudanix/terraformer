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
	"google.golang.org/api/osconfig/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var osConfigAllowEmptyValues = []string{""}

var osConfigAdditionalFields = map[string]interface{}{}

type OsConfigGenerator struct {
	GCPService
}

// Generate TerraformResources from GCP API,
func (g *OsConfigGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := osconfig.NewService(ctx)
	if err != nil {
		return err
	}
	location := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + g.GetArgs()["project"].(string) + "/locations/" + location

	project := g.GetArgs()["project"].(string)
	if err := svc.Projects.Locations.OsPolicyAssignments.List(parent).Pages(ctx, func(page *osconfig.ListOSPolicyAssignmentsResponse) error {
		for _, obj := range page.OsPolicyAssignments {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name, name, "google_os_config_os_policy_assignment", g.ProviderName,
				map[string]string{"name": name, "project": project, "location": location},
				osConfigAllowEmptyValues, osConfigAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := svc.Projects.PatchDeployments.List("projects/"+project).Pages(ctx, func(page *osconfig.ListPatchDeploymentsResponse) error {
		for _, obj := range page.PatchDeployments {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name, name, "google_os_config_patch_deployment", g.ProviderName,
				map[string]string{"patch_deployment_id": name, "project": project},
				osConfigAllowEmptyValues, osConfigAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}
