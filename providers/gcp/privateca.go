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
	"google.golang.org/api/privateca/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var privatecaAllowEmptyValues = []string{""}

var privatecaAdditionalFields = map[string]interface{}{}

type PrivatecaGenerator struct {
	GCPService
}

// Generate TerraformResources from GCP API,
func (g *PrivatecaGenerator) InitResources() error {
	ctx := context.Background()
	privatecaService, err := privateca.NewService(ctx)
	if err != nil {
		return err
	}
	project := g.GetArgs()["project"].(string)
	location := g.GetArgs()["region"].(compute.Region).Name

	poolNames := []string{}
	caPoolsList := privatecaService.Projects.Locations.CaPools.List("projects/" + project + "/locations/" + location)
	if err := caPoolsList.Pages(ctx, func(page *privateca.ListCaPoolsResponse) error {
		for _, obj := range page.CaPools {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			poolNames = append(poolNames, name)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_privateca_ca_pool",
				g.ProviderName,
				map[string]string{"name": name, "project": project, "location": location},
				privatecaAllowEmptyValues,
				privatecaAdditionalFields,
			))
			if policy, perr := privatecaService.Projects.Locations.CaPools.GetIamPolicy(obj.Name).Do(); perr == nil {
				for _, b := range policy.Bindings {
					for _, m := range b.Members {
						g.Resources = append(g.Resources, terraformutils.NewResource(
							obj.Name+" "+b.Role+" "+m, name+"_"+b.Role+"_"+m,
							"google_privateca_ca_pool_iam_member", g.ProviderName,
							map[string]string{"ca_pool": name, "role": b.Role, "member": m, "project": project, "location": location},
							privatecaAllowEmptyValues, privatecaAdditionalFields))
					}
				}
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	if err := privatecaService.Projects.Locations.CertificateTemplates.List("projects/"+project+"/locations/"+location).Pages(ctx, func(p *privateca.ListCertificateTemplatesResponse) error {
		for _, o := range p.CertificateTemplates {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_privateca_certificate_template", g.ProviderName,
				map[string]string{"name": name, "project": project, "location": location},
				privatecaAllowEmptyValues, privatecaAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	// Walk each CA pool for its certificate authorities.
	for _, pool := range poolNames {
		caList := privatecaService.Projects.Locations.CaPools.CertificateAuthorities.List(
			"projects/" + project + "/locations/" + location + "/caPools/" + pool)
		if err := caList.Pages(ctx, func(page *privateca.ListCertificateAuthoritiesResponse) error {
			for _, obj := range page.CertificateAuthorities {
				t := strings.Split(obj.Name, "/")
				name := t[len(t)-1]
				g.Resources = append(g.Resources, terraformutils.NewResource(
					obj.Name,
					name,
					"google_privateca_certificate_authority",
					g.ProviderName,
					map[string]string{
						"certificate_authority_id": name,
						"pool":                     pool,
						"project":                  project,
						"location":                 location,
					},
					privatecaAllowEmptyValues,
					privatecaAdditionalFields,
				))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
		if err := privatecaService.Projects.Locations.CaPools.Certificates.List(
			"projects/"+project+"/locations/"+location+"/caPools/"+pool).Pages(ctx, func(page *privateca.ListCertificatesResponse) error {
			for _, obj := range page.Certificates {
				t := strings.Split(obj.Name, "/")
				name := t[len(t)-1]
				g.Resources = append(g.Resources, terraformutils.NewResource(
					obj.Name, name, "google_privateca_certificate", g.ProviderName,
					map[string]string{"name": name, "pool": pool, "project": project, "location": location},
					privatecaAllowEmptyValues, privatecaAdditionalFields))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
	}
	return nil
}
