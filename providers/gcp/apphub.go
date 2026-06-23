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

	"google.golang.org/api/apphub/v1"
	"google.golang.org/api/compute/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var apphubAllowEmptyValues = []string{""}

var apphubAdditionalFields = map[string]interface{}{}

type ApphubGenerator struct {
	GCPService
}

// Run on applicationsList and create for each TerraformResource
func (g ApphubGenerator) createResources(ctx context.Context, applicationsList *apphub.ProjectsLocationsApplicationsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := applicationsList.Pages(ctx, func(page *apphub.ListApplicationsResponse) error {
		for _, obj := range page.Applications {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_apphub_application",
				g.ProviderName,
				map[string]string{
					"name":     name,
					"project":  g.GetArgs()["project"].(string),
					"location": location,
				},
				apphubAllowEmptyValues,
				apphubAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *ApphubGenerator) InitResources() error {
	ctx := context.Background()
	apphubService, err := apphub.NewService(ctx)
	if err != nil {
		return err
	}

	parent := "projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name
	applicationsList := apphubService.Projects.Locations.Applications.List(parent)
	g.Resources = append(g.Resources, g.createResources(ctx, applicationsList)...)

	if err := apphubService.Projects.Locations.ServiceProjectAttachments.List(parent).Pages(ctx, func(p *apphub.ListServiceProjectAttachmentsResponse) error {
		for _, o := range p.ServiceProjectAttachments {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_apphub_service_project_attachment", g.ProviderName,
				map[string]string{"service_project_attachment_id": name, "project": g.GetArgs()["project"].(string), "location": g.GetArgs()["region"].(compute.Region).Name},
				apphubAllowEmptyValues, apphubAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}
