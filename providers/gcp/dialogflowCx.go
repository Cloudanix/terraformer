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
	"google.golang.org/api/dialogflow/v3"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var dialogflowCxAllowEmptyValues = []string{""}

var dialogflowCxAdditionalFields = map[string]interface{}{}

type DialogflowCxGenerator struct {
	GCPService
}

// Generate TerraformResources from GCP Dialogflow CX (v3) API.
func (g *DialogflowCxGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := dialogflow.NewService(ctx)
	if err != nil {
		return err
	}
	project := g.GetArgs()["project"].(string)
	location := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + location
	tail := func(s string) string { p := strings.Split(s, "/"); return p[len(p)-1] }
	add := func(id, name, tfType string) {
		g.Resources = append(g.Resources, terraformutils.NewResource(
			id, name, tfType, g.ProviderName, map[string]string{"project": project, "location": location},
			dialogflowCxAllowEmptyValues, dialogflowCxAdditionalFields))
	}

	if err := svc.Projects.Locations.SecuritySettings.List(parent).Pages(ctx, func(p *dialogflow.GoogleCloudDialogflowCxV3ListSecuritySettingsResponse) error {
		for _, o := range p.SecuritySettings {
			add(o.Name, tail(o.Name), "google_dialogflow_cx_security_settings")
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	if err := svc.Projects.Locations.Agents.List(parent).Pages(ctx, func(p *dialogflow.GoogleCloudDialogflowCxV3ListAgentsResponse) error {
		for _, a := range p.Agents {
			agent := a.Name
			if e := svc.Projects.Locations.Agents.Environments.List(agent).Pages(ctx, func(r *dialogflow.GoogleCloudDialogflowCxV3ListEnvironmentsResponse) error {
				for _, o := range r.Environments {
					add(o.Name, tail(o.Name), "google_dialogflow_cx_environment")
				}
				return nil
			}); e != nil {
				log.Println(e)
			}
			if e := svc.Projects.Locations.Agents.Generators.List(agent).Pages(ctx, func(r *dialogflow.GoogleCloudDialogflowCxV3ListGeneratorsResponse) error {
				for _, o := range r.Generators {
					add(o.Name, tail(o.Name), "google_dialogflow_cx_generator")
				}
				return nil
			}); e != nil {
				log.Println(e)
			}
			if e := svc.Projects.Locations.Agents.Playbooks.List(agent).Pages(ctx, func(r *dialogflow.GoogleCloudDialogflowCxV3ListPlaybooksResponse) error {
				for _, o := range r.Playbooks {
					add(o.Name, tail(o.Name), "google_dialogflow_cx_playbook")
				}
				return nil
			}); e != nil {
				log.Println(e)
			}
			if e := svc.Projects.Locations.Agents.TestCases.List(agent).Pages(ctx, func(r *dialogflow.GoogleCloudDialogflowCxV3ListTestCasesResponse) error {
				for _, o := range r.TestCases {
					add(o.Name, tail(o.Name), "google_dialogflow_cx_test_case")
				}
				return nil
			}); e != nil {
				log.Println(e)
			}
			if e := svc.Projects.Locations.Agents.Tools.List(agent).Pages(ctx, func(r *dialogflow.GoogleCloudDialogflowCxV3ListToolsResponse) error {
				for _, o := range r.Tools {
					add(o.Name, tail(o.Name), "google_dialogflow_cx_tool")
					if ve := svc.Projects.Locations.Agents.Tools.Versions.List(o.Name).Pages(ctx, func(vr *dialogflow.GoogleCloudDialogflowCxV3ListToolVersionsResponse) error {
						for _, v := range vr.ToolVersions {
							add(v.Name, tail(o.Name)+"_"+tail(v.Name), "google_dialogflow_cx_tool_version")
						}
						return nil
					}); ve != nil {
						log.Println(ve)
					}
				}
				return nil
			}); e != nil {
				log.Println(e)
			}
			if e := svc.Projects.Locations.Agents.Flows.List(agent).Pages(ctx, func(r *dialogflow.GoogleCloudDialogflowCxV3ListFlowsResponse) error {
				for _, f := range r.Flows {
					if pe := svc.Projects.Locations.Agents.Flows.Pages.List(f.Name).Pages(ctx, func(pr *dialogflow.GoogleCloudDialogflowCxV3ListPagesResponse) error {
						for _, o := range pr.Pages {
							add(o.Name, tail(f.Name)+"_"+tail(o.Name), "google_dialogflow_cx_page")
						}
						return nil
					}); pe != nil {
						log.Println(pe)
					}
					if ve := svc.Projects.Locations.Agents.Flows.Versions.List(f.Name).Pages(ctx, func(vr *dialogflow.GoogleCloudDialogflowCxV3ListVersionsResponse) error {
						for _, o := range vr.Versions {
							add(o.Name, tail(f.Name)+"_"+tail(o.Name), "google_dialogflow_cx_version")
						}
						return nil
					}); ve != nil {
						log.Println(ve)
					}
				}
				return nil
			}); e != nil {
				log.Println(e)
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}
