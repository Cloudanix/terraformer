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

	"google.golang.org/api/apihub/v1"
	"google.golang.org/api/compute/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var apihubAllowEmptyValues = []string{""}

var apihubAdditionalFields = map[string]interface{}{}

type ApihubGenerator struct {
	GCPService
}

// Generate TerraformResources from GCP API,
func (g *ApihubGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := apihub.NewService(ctx)
	if err != nil {
		return err
	}
	project := g.GetArgs()["project"].(string)
	location := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + location

	if err := svc.Projects.Locations.Curations.List(parent).Pages(ctx, func(p *apihub.GoogleCloudApihubV1ListCurationsResponse) error {
		for _, o := range p.Curations {
			t := strings.Split(o.Name, "/")
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, t[len(t)-1], "google_apihub_curation", g.ProviderName,
				map[string]string{"curation_id": t[len(t)-1], "location": location, "project": project},
				apihubAllowEmptyValues, apihubAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := svc.Projects.Locations.HostProjectRegistrations.List(parent).Pages(ctx, func(p *apihub.GoogleCloudApihubV1ListHostProjectRegistrationsResponse) error {
		for _, o := range p.HostProjectRegistrations {
			t := strings.Split(o.Name, "/")
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, t[len(t)-1], "google_apihub_host_project_registration", g.ProviderName,
				map[string]string{"host_project_registration_id": t[len(t)-1], "location": location, "project": project},
				apihubAllowEmptyValues, apihubAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := svc.Projects.Locations.Plugins.List(parent).Pages(ctx, func(p *apihub.GoogleCloudApihubV1ListPluginsResponse) error {
		for _, o := range p.Plugins {
			t := strings.Split(o.Name, "/")
			pluginID := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, pluginID, "google_apihub_plugin", g.ProviderName,
				map[string]string{"plugin_id": pluginID, "location": location, "project": project},
				apihubAllowEmptyValues, apihubAdditionalFields))
			if ierr := svc.Projects.Locations.Plugins.Instances.List(o.Name).Pages(ctx, func(ip *apihub.GoogleCloudApihubV1ListPluginInstancesResponse) error {
				for _, pi := range ip.PluginInstances {
					it := strings.Split(pi.Name, "/")
					g.Resources = append(g.Resources, terraformutils.NewResource(
						pi.Name, pluginID+"_"+it[len(it)-1], "google_apihub_plugin_instance", g.ProviderName,
						map[string]string{"plugin_instance_id": it[len(it)-1], "plugin": pluginID, "location": location, "project": project},
						apihubAllowEmptyValues, apihubAdditionalFields))
				}
				return nil
			}); ierr != nil {
				log.Println(ierr)
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}
