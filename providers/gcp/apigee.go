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
			add(o.Name, tail(o.Name), "google_apigee_envgroup", map[string]string{"name": tail(o.Name), "org_id": org})
		}
	} else {
		log.Println(e)
	}
	if r, e := svc.Organizations.Instances.List(org).Do(); e == nil {
		for _, o := range r.Instances {
			add(o.Name, tail(o.Name), "google_apigee_instance", map[string]string{"name": tail(o.Name), "org_id": org})
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
	return nil
}
