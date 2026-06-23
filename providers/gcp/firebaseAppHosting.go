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
	"google.golang.org/api/firebaseapphosting/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var firebaseAppHostingAllowEmptyValues = []string{""}

var firebaseAppHostingAdditionalFields = map[string]interface{}{}

type FirebaseAppHostingGenerator struct {
	GCPService
}

// Generate TerraformResources from GCP API,
func (g *FirebaseAppHostingGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := firebaseapphosting.NewService(ctx)
	if err != nil {
		return err
	}
	project := g.GetArgs()["project"].(string)
	location := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + location
	tail := func(s string) string { p := strings.Split(s, "/"); return p[len(p)-1] }

	if err := svc.Projects.Locations.Backends.List(parent).Pages(ctx, func(p *firebaseapphosting.ListBackendsResponse) error {
		for _, b := range p.Backends {
			backend := tail(b.Name)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				b.Name, backend, "google_firebase_app_hosting_backend", g.ProviderName,
				map[string]string{"backend_id": backend, "location": location, "project": project},
				firebaseAppHostingAllowEmptyValues, firebaseAppHostingAdditionalFields))
			if e := svc.Projects.Locations.Backends.Builds.List(b.Name).Pages(ctx, func(bp *firebaseapphosting.ListBuildsResponse) error {
				for _, o := range bp.Builds {
					g.Resources = append(g.Resources, terraformutils.NewResource(
						o.Name, backend+"_"+tail(o.Name), "google_firebase_app_hosting_build", g.ProviderName,
						map[string]string{"build_id": tail(o.Name), "backend": backend, "location": location, "project": project},
						firebaseAppHostingAllowEmptyValues, firebaseAppHostingAdditionalFields))
				}
				return nil
			}); e != nil {
				log.Println(e)
			}
			if e := svc.Projects.Locations.Backends.Domains.List(b.Name).Pages(ctx, func(dp *firebaseapphosting.ListDomainsResponse) error {
				for _, o := range dp.Domains {
					g.Resources = append(g.Resources, terraformutils.NewResource(
						o.Name, backend+"_"+tail(o.Name), "google_firebase_app_hosting_domain", g.ProviderName,
						map[string]string{"domain_id": tail(o.Name), "backend": backend, "location": location, "project": project},
						firebaseAppHostingAllowEmptyValues, firebaseAppHostingAdditionalFields))
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
