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

	"google.golang.org/api/apigee/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var apigeeAllowEmptyValues = []string{""}

var apigeeAdditionalFields = map[string]interface{}{}

type ApigeeGenerator struct {
	GCPService
}

// Generate TerraformResources from GCP API,
// Apigee orgs map 1:1 to the GCP project (org name == project id).
func (g *ApigeeGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := apigee.NewService(ctx)
	if err != nil {
		return err
	}
	project := g.GetArgs()["project"].(string)
	org := "organizations/" + project

	add := func(id, name, tfType string, attrs map[string]string) {
		g.Resources = append(g.Resources, terraformutils.NewResource(
			id, name, tfType, g.ProviderName, attrs, apigeeAllowEmptyValues, apigeeAdditionalFields))
	}
	tail := func(s string) string { p := strings.Split(s, "/"); return p[len(p)-1] }

	if r, e := svc.Organizations.Apiproducts.List(org).Do(); e == nil {
		for _, o := range r.ApiProduct {
			add(org+"/apiproducts/"+o.Name, o.Name, "google_apigee_api_product", map[string]string{"name": o.Name, "org_id": org})
		}
	} else {
		log.Println(e)
	}
	if r, e := svc.Organizations.Developers.List(org).Do(); e == nil {
		for _, o := range r.Developer {
			add(org+"/developers/"+o.Email, o.Email, "google_apigee_developer", map[string]string{"email": o.Email, "org_id": org})
		}
	} else {
		log.Println(e)
	}
	if r, e := svc.Organizations.Datacollectors.List(org).Do(); e == nil {
		for _, o := range r.DataCollectors {
			add(o.Name, tail(o.Name), "google_apigee_data_collector", map[string]string{"name": tail(o.Name), "org_id": org})
		}
	} else {
		log.Println(e)
	}
	if r, e := svc.Organizations.Envgroups.List(org).Do(); e == nil {
		for _, o := range r.EnvironmentGroups {
			eg := tail(o.Name)
			add(o.Name, eg, "google_apigee_envgroup", map[string]string{"name": eg, "org_id": org})
			if ae := svc.Organizations.Envgroups.Attachments.List(o.Name).Pages(ctx, func(ap *apigee.GoogleCloudApigeeV1ListEnvironmentGroupAttachmentsResponse) error {
				for _, a := range ap.EnvironmentGroupAttachments {
					add(o.Name+"/attachments/"+a.Name, eg+"_"+a.Name, "google_apigee_envgroup_attachment",
						map[string]string{"envgroup_id": eg, "org_id": org})
				}
				return nil
			}); ae != nil {
				log.Println(ae)
			}
		}
	} else {
		log.Println(e)
	}
	if r, e := svc.Organizations.Instances.List(org).Do(); e == nil {
		for _, o := range r.Instances {
			inst := tail(o.Name)
			add(o.Name, inst, "google_apigee_instance", map[string]string{"name": inst, "org_id": org})
			if ae := svc.Organizations.Instances.Attachments.List(o.Name).Pages(ctx, func(ap *apigee.GoogleCloudApigeeV1ListInstanceAttachmentsResponse) error {
				for _, a := range ap.Attachments {
					add(o.Name+"/attachments/"+a.Name, inst+"_"+a.Name, "google_apigee_instance_attachment",
						map[string]string{"instance_id": inst, "org_id": org})
				}
				return nil
			}); ae != nil {
				log.Println(ae)
			}
		}
	} else {
		log.Println(e)
	}
	if r, e := svc.Organizations.EndpointAttachments.List(org).Do(); e == nil {
		for _, o := range r.EndpointAttachments {
			add(o.Name, tail(o.Name), "google_apigee_endpoint_attachment", map[string]string{"endpoint_attachment_id": tail(o.Name), "org_id": org})
		}
	} else {
		log.Println(e)
	}
	if r, e := svc.Organizations.DnsZones.List(org).Do(); e == nil {
		for _, o := range r.DnsZones {
			add(o.Name, tail(o.Name), "google_apigee_dns_zone", map[string]string{"dns_zone_id": tail(o.Name), "org_id": org})
		}
	} else {
		log.Println(e)
	}
	if r, e := svc.Organizations.Appgroups.List(org).Do(); e == nil {
		for _, o := range r.AppGroups {
			add(o.Name, o.Name, "google_apigee_app_group", map[string]string{"name": o.Name, "org_id": org})
		}
	} else {
		log.Println(e)
	}

	// Shared flows (org-level).
	if r, e := svc.Organizations.Sharedflows.List(org).Do(); e == nil {
		for _, o := range r.SharedFlows {
			add(o.Name, tail(o.Name), "google_apigee_sharedflow", map[string]string{"name": tail(o.Name), "org_id": org})
		}
	} else {
		log.Println(e)
	}

	// Environments come from the org resource as a list of names; walk each for security actions.
	if o, e := svc.Organizations.Get(org).Do(); e == nil {
		for _, env := range o.Environments {
			envParent := org + "/environments/" + env
			add(envParent, env, "google_apigee_environment", map[string]string{"name": env, "org_id": org})
			if se := svc.Organizations.Environments.SecurityActions.List(envParent).Pages(ctx, func(sp *apigee.GoogleCloudApigeeV1ListSecurityActionsResponse) error {
				for _, sa := range sp.SecurityActions {
					add(sa.Name, env+"_"+tail(sa.Name), "google_apigee_security_action",
						map[string]string{"security_action_id": tail(sa.Name), "env_id": env, "org_id": org})
				}
				return nil
			}); se != nil {
				log.Println(se)
			}
		}
	} else {
		log.Println(e)
	}
	return nil
}
