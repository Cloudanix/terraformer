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

// Generate TerraformResources from GCP API,
func (g *VmwareengineGenerator) InitResources() error {
	ctx := context.Background()
	vmwareengineService, err := vmwareengine.NewService(ctx)
	if err != nil {
		return err
	}

	project := g.GetArgs()["project"].(string)
	location := g.GetArgs()["region"].(compute.Region).Name

	pcNames := []string{}
	privateCloudsList := vmwareengineService.Projects.Locations.PrivateClouds.List("projects/" + project + "/locations/" + location)
	if err := privateCloudsList.Pages(ctx, func(page *vmwareengine.ListPrivateCloudsResponse) error {
		for _, obj := range page.PrivateClouds {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			pcNames = append(pcNames, name)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_vmwareengine_private_cloud",
				g.ProviderName,
				map[string]string{"name": name, "project": project, "location": location},
				vmwareengineAllowEmptyValues,
				vmwareengineAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	for _, pc := range pcNames {
		clustersList := vmwareengineService.Projects.Locations.PrivateClouds.Clusters.List(
			"projects/" + project + "/locations/" + location + "/privateClouds/" + pc)
		if err := clustersList.Pages(ctx, func(page *vmwareengine.ListClustersResponse) error {
			for _, obj := range page.Clusters {
				t := strings.Split(obj.Name, "/")
				name := t[len(t)-1]
				g.Resources = append(g.Resources, terraformutils.NewResource(
					obj.Name,
					name,
					"google_vmwareengine_cluster",
					g.ProviderName,
					map[string]string{
						"name":          name,
						"parent":        "projects/" + project + "/locations/" + location + "/privateClouds/" + pc,
						"project":       project,
						"location":      location,
						"private_cloud": pc,
					},
					vmwareengineAllowEmptyValues,
					vmwareengineAdditionalFields,
				))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
	}
	return nil
}
