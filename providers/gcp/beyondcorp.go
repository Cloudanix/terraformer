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

	"google.golang.org/api/beyondcorp/v1"
	"google.golang.org/api/compute/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var beyondcorpAllowEmptyValues = []string{""}

var beyondcorpAdditionalFields = map[string]interface{}{}

type BeyondcorpGenerator struct {
	GCPService
}

// Generate TerraformResources from GCP API,
func (g *BeyondcorpGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := beyondcorp.NewService(ctx)
	if err != nil {
		return err
	}
	project := g.GetArgs()["project"].(string)
	location := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + location

	if err := svc.Projects.Locations.AppConnections.List(parent).Pages(ctx, func(p *beyondcorp.GoogleCloudBeyondcorpAppconnectionsV1ListAppConnectionsResponse) error {
		for _, o := range p.AppConnections {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_beyondcorp_app_connection", g.ProviderName,
				map[string]string{"name": name, "project": project, "region": location},
				beyondcorpAllowEmptyValues, beyondcorpAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := svc.Projects.Locations.AppConnectors.List(parent).Pages(ctx, func(p *beyondcorp.GoogleCloudBeyondcorpAppconnectorsV1ListAppConnectorsResponse) error {
		for _, o := range p.AppConnectors {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_beyondcorp_app_connector", g.ProviderName,
				map[string]string{"name": name, "project": project, "region": location},
				beyondcorpAllowEmptyValues, beyondcorpAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := svc.Projects.Locations.AppGateways.List(parent).Pages(ctx, func(p *beyondcorp.ListAppGatewaysResponse) error {
		for _, o := range p.AppGateways {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_beyondcorp_app_gateway", g.ProviderName,
				map[string]string{"name": name, "project": project, "region": location},
				beyondcorpAllowEmptyValues, beyondcorpAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	// Security gateways are global.
	globalParent := "projects/" + project + "/locations/global"
	if err := svc.Projects.Locations.SecurityGateways.List(globalParent).Pages(ctx, func(p *beyondcorp.GoogleCloudBeyondcorpSecuritygatewaysV1ListSecurityGatewaysResponse) error {
		for _, o := range p.SecurityGateways {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_beyondcorp_security_gateway", g.ProviderName,
				map[string]string{"security_gateway_id": name, "project": project},
				beyondcorpAllowEmptyValues, beyondcorpAdditionalFields))
			if aerr := svc.Projects.Locations.SecurityGateways.Applications.List(o.Name).Pages(ctx, func(ap *beyondcorp.GoogleCloudBeyondcorpSecuritygatewaysV1ListApplicationsResponse) error {
				for _, a := range ap.Applications {
					at := strings.Split(a.Name, "/")
					g.Resources = append(g.Resources, terraformutils.NewResource(
						a.Name, name+"_"+at[len(at)-1], "google_beyondcorp_security_gateway_application", g.ProviderName,
						map[string]string{"application_id": at[len(at)-1], "security_gateway_id": name, "project": project},
						beyondcorpAllowEmptyValues, beyondcorpAdditionalFields))
				}
				return nil
			}); aerr != nil {
				log.Println(aerr)
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}
