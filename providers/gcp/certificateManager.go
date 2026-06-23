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

	"google.golang.org/api/certificatemanager/v1"
	"google.golang.org/api/compute/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var certificateManagerAllowEmptyValues = []string{""}

var certificateManagerAdditionalFields = map[string]interface{}{}

type CertificateManagerGenerator struct {
	GCPService
}

// Run on certificatesList and create for each TerraformResource
func (g CertificateManagerGenerator) createResources(ctx context.Context, certificatesList *certificatemanager.ProjectsLocationsCertificatesListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := certificatesList.Pages(ctx, func(page *certificatemanager.ListCertificatesResponse) error {
		for _, obj := range page.Certificates {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_certificate_manager_certificate",
				g.ProviderName,
				map[string]string{
					"name":     name,
					"project":  g.GetArgs()["project"].(string),
					"location": location,
				},
				certificateManagerAllowEmptyValues,
				certificateManagerAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *CertificateManagerGenerator) InitResources() error {
	ctx := context.Background()
	certificateManagerService, err := certificatemanager.NewService(ctx)
	if err != nil {
		return err
	}

	parent := "projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name
	certificatesList := certificateManagerService.Projects.Locations.Certificates.List(parent)
	g.Resources = append(g.Resources, g.createResources(ctx, certificatesList)...)

	mapsList := certificateManagerService.Projects.Locations.CertificateMaps.List(parent)
	g.Resources = append(g.Resources, g.createMapsResources(ctx, mapsList)...)

	authsList := certificateManagerService.Projects.Locations.DnsAuthorizations.List(parent)
	g.Resources = append(g.Resources, g.createDNSAuthResources(ctx, authsList)...)

	trustList := certificateManagerService.Projects.Locations.TrustConfigs.List(parent)
	g.Resources = append(g.Resources, g.createTrustConfigResources(ctx, trustList)...)

	cicList := certificateManagerService.Projects.Locations.CertificateIssuanceConfigs.List(parent)
	if err := cicList.Pages(ctx, func(page *certificatemanager.ListCertificateIssuanceConfigsResponse) error {
		for _, obj := range page.CertificateIssuanceConfigs {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name, name, "google_certificate_manager_certificate_issuance_config", g.ProviderName,
				map[string]string{"name": name, "project": g.GetArgs()["project"].(string), "location": g.GetArgs()["region"].(compute.Region).Name},
				certificateManagerAllowEmptyValues, certificateManagerAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}

func (g CertificateManagerGenerator) createTrustConfigResources(ctx context.Context, list *certificatemanager.ProjectsLocationsTrustConfigsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := list.Pages(ctx, func(page *certificatemanager.ListTrustConfigsResponse) error {
		for _, obj := range page.TrustConfigs {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name, name, "google_certificate_manager_trust_config", g.ProviderName,
				map[string]string{"name": name, "project": g.GetArgs()["project"].(string), "location": location},
				certificateManagerAllowEmptyValues, certificateManagerAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

func (g CertificateManagerGenerator) createMapsResources(ctx context.Context, list *certificatemanager.ProjectsLocationsCertificateMapsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := list.Pages(ctx, func(page *certificatemanager.ListCertificateMapsResponse) error {
		for _, obj := range page.CertificateMaps {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name, name, "google_certificate_manager_certificate_map", g.ProviderName,
				map[string]string{"name": name, "project": g.GetArgs()["project"].(string), "location": location},
				certificateManagerAllowEmptyValues, certificateManagerAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

func (g CertificateManagerGenerator) createDNSAuthResources(ctx context.Context, list *certificatemanager.ProjectsLocationsDnsAuthorizationsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := list.Pages(ctx, func(page *certificatemanager.ListDnsAuthorizationsResponse) error {
		for _, obj := range page.DnsAuthorizations {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name, name, "google_certificate_manager_dns_authorization", g.ProviderName,
				map[string]string{"name": name, "project": g.GetArgs()["project"].(string), "location": location},
				certificateManagerAllowEmptyValues, certificateManagerAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}
