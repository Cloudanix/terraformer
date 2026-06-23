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
	"google.golang.org/api/dlp/v2"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var dlpAllowEmptyValues = []string{""}

var dlpAdditionalFields = map[string]interface{}{}

type DlpGenerator struct {
	GCPService
}

// Generate TerraformResources from GCP API,
func (g *DlpGenerator) InitResources() error {
	ctx := context.Background()
	dlpService, err := dlp.NewService(ctx)
	if err != nil {
		return err
	}
	project := g.GetArgs()["project"].(string)
	parent := "projects/" + project

	templatesList := dlpService.Projects.InspectTemplates.List(parent)
	if err := templatesList.Pages(ctx, func(page *dlp.GooglePrivacyDlpV2ListInspectTemplatesResponse) error {
		for _, obj := range page.InspectTemplates {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_data_loss_prevention_inspect_template",
				g.ProviderName,
				map[string]string{
					"parent": parent,
				},
				dlpAllowEmptyValues,
				dlpAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	location := g.GetArgs()["region"].(compute.Region).Name
	locParent := "projects/" + project + "/locations/" + location
	tail := func(s string) string { p := strings.Split(s, "/"); return p[len(p)-1] }
	if err := dlpService.Projects.Locations.DeidentifyTemplates.List(locParent).Pages(ctx, func(p *dlp.GooglePrivacyDlpV2ListDeidentifyTemplatesResponse) error {
		for _, o := range p.DeidentifyTemplates {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, tail(o.Name), "google_data_loss_prevention_deidentify_template", g.ProviderName,
				map[string]string{"parent": locParent}, dlpAllowEmptyValues, dlpAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := dlpService.Projects.Locations.JobTriggers.List(locParent).Pages(ctx, func(p *dlp.GooglePrivacyDlpV2ListJobTriggersResponse) error {
		for _, o := range p.JobTriggers {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, tail(o.Name), "google_data_loss_prevention_job_trigger", g.ProviderName,
				map[string]string{"parent": locParent}, dlpAllowEmptyValues, dlpAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := dlpService.Projects.Locations.StoredInfoTypes.List(locParent).Pages(ctx, func(p *dlp.GooglePrivacyDlpV2ListStoredInfoTypesResponse) error {
		for _, o := range p.StoredInfoTypes {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, tail(o.Name), "google_data_loss_prevention_stored_info_type", g.ProviderName,
				map[string]string{"parent": locParent}, dlpAllowEmptyValues, dlpAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := dlpService.Projects.Locations.DiscoveryConfigs.List(locParent).Pages(ctx, func(p *dlp.GooglePrivacyDlpV2ListDiscoveryConfigsResponse) error {
		for _, o := range p.DiscoveryConfigs {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, tail(o.Name), "google_data_loss_prevention_discovery_config", g.ProviderName,
				map[string]string{"parent": locParent, "location": location}, dlpAllowEmptyValues, dlpAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}
