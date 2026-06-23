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
	"strconv"

	"google.golang.org/api/appengine/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var appEngineAllowEmptyValues = []string{""}

var appEngineAdditionalFields = map[string]interface{}{}

type AppEngineGenerator struct {
	GCPService
}

// appEngineVersionType maps an App Engine version's environment to the Terraform
// version resource type. The default (standard) covers env values "standard"/"" ;
// only the flexible environment uses the flexible resource.
func appEngineVersionType(env string) string {
	if env == "flexible" || env == "flex" {
		return "google_app_engine_flexible_app_version"
	}
	return "google_app_engine_standard_app_version"
}

// Generate TerraformResources from GCP App Engine Admin API (appsId == project).
func (g *AppEngineGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := appengine.NewService(ctx)
	if err != nil {
		return err
	}
	project := g.GetArgs()["project"].(string)

	// Single App Engine application per project (emit only if one exists).
	if _, ae := svc.Apps.Get(project).Do(); ae == nil {
		g.Resources = append(g.Resources, terraformutils.NewResource(
			project, project, "google_app_engine_application", g.ProviderName,
			map[string]string{"project": project}, appEngineAllowEmptyValues, appEngineAdditionalFields))
	} else {
		log.Println(ae)
		return nil
	}

	if err := svc.Apps.Firewall.IngressRules.List(project).Pages(ctx, func(p *appengine.ListIngressRulesResponse) error {
		for _, o := range p.IngressRules {
			id := strconv.FormatInt(o.Priority, 10)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				"apps/"+project+"/firewall/ingressRules/"+id, id, "google_app_engine_firewall_rule", g.ProviderName,
				map[string]string{"priority": id, "project": project}, appEngineAllowEmptyValues, appEngineAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := svc.Apps.DomainMappings.List(project).Pages(ctx, func(p *appengine.ListDomainMappingsResponse) error {
		for _, o := range p.DomainMappings {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				"apps/"+project+"/domainMappings/"+o.Id, o.Id, "google_app_engine_domain_mapping", g.ProviderName,
				map[string]string{"domain_name": o.Id, "project": project}, appEngineAllowEmptyValues, appEngineAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := svc.Apps.Services.List(project).Pages(ctx, func(sp *appengine.ListServicesResponse) error {
		for _, s := range sp.Services {
			if e := svc.Apps.Services.Versions.List(project, s.Id).Pages(ctx, func(vp *appengine.ListVersionsResponse) error {
				for _, v := range vp.Versions {
					tfType := appEngineVersionType(v.Env)
					g.Resources = append(g.Resources, terraformutils.NewResource(
						"apps/"+project+"/services/"+s.Id+"/versions/"+v.Id, s.Id+"_"+v.Id, tfType, g.ProviderName,
						map[string]string{"version_id": v.Id, "service": s.Id, "project": project},
						appEngineAllowEmptyValues, appEngineAdditionalFields))
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
