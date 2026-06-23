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
	"google.golang.org/api/contactcenterinsights/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var contactCenterInsightsAllowEmptyValues = []string{""}

var contactCenterInsightsAdditionalFields = map[string]interface{}{}

type ContactCenterInsightsGenerator struct {
	GCPService
}

// Generate TerraformResources from GCP API,
func (g *ContactCenterInsightsGenerator) InitResources() error {
	ctx := context.Background()
	cciService, err := contactcenterinsights.NewService(ctx)
	if err != nil {
		return err
	}
	location := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + g.GetArgs()["project"].(string) + "/locations/" + location

	viewsList := cciService.Projects.Locations.Views.List(parent)
	if err := viewsList.Pages(ctx, func(page *contactcenterinsights.GoogleCloudContactcenterinsightsV1ListViewsResponse) error {
		for _, obj := range page.Views {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name, name, "google_contact_center_insights_view", g.ProviderName,
				map[string]string{"name": name, "project": g.GetArgs()["project"].(string), "location": location},
				contactCenterInsightsAllowEmptyValues, contactCenterInsightsAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}
