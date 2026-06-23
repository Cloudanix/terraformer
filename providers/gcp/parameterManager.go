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
	"google.golang.org/api/parametermanager/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var parameterManagerAllowEmptyValues = []string{""}

var parameterManagerAdditionalFields = map[string]interface{}{}

type ParameterManagerGenerator struct {
	GCPService
}

// walkParameters enumerates parameters + versions at a given location, emitting the
// supplied tf types (global vs regional share the same API shape).
func (g *ParameterManagerGenerator) walkParameters(ctx context.Context, svc *parametermanager.Service, project, location, paramType, versionType string) {
	parent := "projects/" + project + "/locations/" + location
	if err := svc.Projects.Locations.Parameters.List(parent).Pages(ctx, func(p *parametermanager.ListParametersResponse) error {
		for _, o := range p.Parameters {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			attrs := map[string]string{"parameter_id": name, "project": project}
			if location != "global" {
				attrs["location"] = location
			}
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, paramType, g.ProviderName, attrs,
				parameterManagerAllowEmptyValues, parameterManagerAdditionalFields))
			if verr := svc.Projects.Locations.Parameters.Versions.List(o.Name).Pages(ctx, func(vp *parametermanager.ListParameterVersionsResponse) error {
				for _, v := range vp.ParameterVersions {
					vt := strings.Split(v.Name, "/")
					vAttrs := map[string]string{"parameter_version_id": vt[len(vt)-1], "parameter": name, "project": project}
					if location != "global" {
						vAttrs["location"] = location
					}
					g.Resources = append(g.Resources, terraformutils.NewResource(
						v.Name, name+"_"+vt[len(vt)-1], versionType, g.ProviderName, vAttrs,
						parameterManagerAllowEmptyValues, parameterManagerAdditionalFields))
				}
				return nil
			}); verr != nil {
				log.Println(verr)
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
}

// Generate TerraformResources from GCP API,
func (g *ParameterManagerGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := parametermanager.NewService(ctx)
	if err != nil {
		return err
	}
	project := g.GetArgs()["project"].(string)
	location := g.GetArgs()["region"].(compute.Region).Name

	g.walkParameters(ctx, svc, project, "global", "google_parameter_manager_parameter", "google_parameter_manager_parameter_version")
	g.walkParameters(ctx, svc, project, location, "google_parameter_manager_regional_parameter", "google_parameter_manager_regional_parameter_version")
	return nil
}
