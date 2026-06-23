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
	"google.golang.org/api/vmwareengine/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var vmwareengineAllowEmptyValues = []string{""}

var vmwareengineAdditionalFields = map[string]interface{}{}

type VmwareengineGenerator struct {
	GCPService
}

// Run on privateCloudsList and create for each TerraformResource
func (g VmwareengineGenerator) createResources(ctx context.Context, privateCloudsList *vmwareengine.ProjectsLocationsPrivateCloudsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := privateCloudsList.Pages(ctx, func(page *vmwareengine.ListPrivateCloudsResponse) error {
		for _, obj := range page.PrivateClouds {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_vmwareengine_private_cloud",
				g.ProviderName,
				map[string]string{
					"name":     name,
					"project":  g.GetArgs()["project"].(string),
					"location": location,
				},
				vmwareengineAllowEmptyValues,
				vmwareengineAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *VmwareengineGenerator) InitResources() error {
	ctx := context.Background()
	vmwareengineService, err := vmwareengine.NewService(ctx)
	if err != nil {
		return err
	}

	privateCloudsList := vmwareengineService.Projects.Locations.PrivateClouds.List(
		"projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name)
	g.Resources = g.createResources(ctx, privateCloudsList)
	return nil
}
