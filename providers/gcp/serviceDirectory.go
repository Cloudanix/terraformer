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
	"google.golang.org/api/servicedirectory/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var serviceDirectoryAllowEmptyValues = []string{""}

var serviceDirectoryAdditionalFields = map[string]interface{}{}

type ServiceDirectoryGenerator struct {
	GCPService
}

// Generate TerraformResources from GCP API,
func (g *ServiceDirectoryGenerator) InitResources() error {
	ctx := context.Background()
	serviceDirectoryService, err := servicedirectory.NewService(ctx)
	if err != nil {
		return err
	}

	project := g.GetArgs()["project"].(string)
	location := g.GetArgs()["region"].(compute.Region).Name
	nsNames := []string{}
	namespacesList := serviceDirectoryService.Projects.Locations.Namespaces.List("projects/" + project + "/locations/" + location)
	if err := namespacesList.Pages(ctx, func(page *servicedirectory.ListNamespacesResponse) error {
		for _, obj := range page.Namespaces {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			nsNames = append(nsNames, obj.Name)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name, name, "google_service_directory_namespace", g.ProviderName,
				map[string]string{"name": name, "project": project, "location": location},
				serviceDirectoryAllowEmptyValues, serviceDirectoryAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	for _, ns := range nsNames {
		if err := serviceDirectoryService.Projects.Locations.Namespaces.Services.List(ns).Pages(ctx, func(p *servicedirectory.ListServicesResponse) error {
			for _, o := range p.Services {
				t := strings.Split(o.Name, "/")
				g.Resources = append(g.Resources, terraformutils.NewResource(
					o.Name, t[len(t)-1], "google_service_directory_service", g.ProviderName,
					map[string]string{"namespace": ns, "project": project},
					serviceDirectoryAllowEmptyValues, serviceDirectoryAdditionalFields))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
	}
	return nil
}
