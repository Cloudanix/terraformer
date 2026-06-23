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
	"google.golang.org/api/networksecurity/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var networkSecurityAllowEmptyValues = []string{""}

var networkSecurityAdditionalFields = map[string]interface{}{}

type NetworkSecurityGenerator struct {
	GCPService
}

func (g NetworkSecurityGenerator) createServerTLSResources(ctx context.Context, list *networksecurity.ProjectsLocationsServerTlsPoliciesListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := list.Pages(ctx, func(page *networksecurity.ListServerTlsPoliciesResponse) error {
		for _, obj := range page.ServerTlsPolicies {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name, name, "google_network_security_server_tls_policy", g.ProviderName,
				map[string]string{"name": name, "project": g.GetArgs()["project"].(string), "location": location},
				networkSecurityAllowEmptyValues, networkSecurityAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

func (g NetworkSecurityGenerator) createClientTLSResources(ctx context.Context, list *networksecurity.ProjectsLocationsClientTlsPoliciesListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := list.Pages(ctx, func(page *networksecurity.ListClientTlsPoliciesResponse) error {
		for _, obj := range page.ClientTlsPolicies {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name, name, "google_network_security_client_tls_policy", g.ProviderName,
				map[string]string{"name": name, "project": g.GetArgs()["project"].(string), "location": location},
				networkSecurityAllowEmptyValues, networkSecurityAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *NetworkSecurityGenerator) InitResources() error {
	ctx := context.Background()
	nsService, err := networksecurity.NewService(ctx)
	if err != nil {
		return err
	}
	parent := "projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name
	g.Resources = append(g.Resources, g.createServerTLSResources(ctx, nsService.Projects.Locations.ServerTlsPolicies.List(parent))...)
	g.Resources = append(g.Resources, g.createClientTLSResources(ctx, nsService.Projects.Locations.ClientTlsPolicies.List(parent))...)

	loc := g.GetArgs()["region"].(compute.Region).Name
	proj := g.GetArgs()["project"].(string)
	if err := nsService.Projects.Locations.AddressGroups.List(parent).Pages(ctx, func(p *networksecurity.ListAddressGroupsResponse) error {
		for _, o := range p.AddressGroups {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_network_security_address_group", g.ProviderName,
				map[string]string{"name": name, "project": proj, "location": loc},
				networkSecurityAllowEmptyValues, networkSecurityAdditionalFields))
			if policy, perr := nsService.Projects.Locations.AddressGroups.GetIamPolicy(o.Name).Do(); perr == nil {
				for _, b := range policy.Bindings {
					for _, m := range b.Members {
						g.Resources = append(g.Resources, terraformutils.NewResource(
							o.Name+" "+b.Role+" "+m, name+"_"+b.Role+"_"+m,
							"google_network_security_address_group_iam_member", g.ProviderName,
							map[string]string{"name": name, "role": b.Role, "member": m, "project": proj, "location": loc},
							networkSecurityAllowEmptyValues, networkSecurityAdditionalFields))
					}
				}
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := nsService.Projects.Locations.AuthzPolicies.List(parent).Pages(ctx, func(p *networksecurity.ListAuthzPoliciesResponse) error {
		for _, o := range p.AuthzPolicies {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_network_security_authz_policy", g.ProviderName,
				map[string]string{"name": name, "project": proj, "location": loc},
				networkSecurityAllowEmptyValues, networkSecurityAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := nsService.Projects.Locations.BackendAuthenticationConfigs.List(parent).Pages(ctx, func(p *networksecurity.ListBackendAuthenticationConfigsResponse) error {
		for _, o := range p.BackendAuthenticationConfigs {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_network_security_backend_authentication_config", g.ProviderName,
				map[string]string{"name": name, "project": proj, "location": loc},
				networkSecurityAllowEmptyValues, networkSecurityAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := nsService.Projects.Locations.DnsThreatDetectors.List(parent).Pages(ctx, func(p *networksecurity.ListDnsThreatDetectorsResponse) error {
		for _, o := range p.DnsThreatDetectors {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_network_security_dns_threat_detector", g.ProviderName,
				map[string]string{"name": name, "project": proj, "location": loc},
				networkSecurityAllowEmptyValues, networkSecurityAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := nsService.Projects.Locations.TlsInspectionPolicies.List(parent).Pages(ctx, func(p *networksecurity.ListTlsInspectionPoliciesResponse) error {
		for _, o := range p.TlsInspectionPolicies {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_network_security_tls_inspection_policy", g.ProviderName,
				map[string]string{"name": name, "project": proj, "location": loc},
				networkSecurityAllowEmptyValues, networkSecurityAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := nsService.Projects.Locations.UrlLists.List(parent).Pages(ctx, func(p *networksecurity.ListUrlListsResponse) error {
		for _, o := range p.UrlLists {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_network_security_url_lists", g.ProviderName,
				map[string]string{"name": name, "project": proj, "location": loc},
				networkSecurityAllowEmptyValues, networkSecurityAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := nsService.Projects.Locations.FirewallEndpointAssociations.List(parent).Pages(ctx, func(p *networksecurity.ListFirewallEndpointAssociationsResponse) error {
		for _, o := range p.FirewallEndpointAssociations {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_network_security_firewall_endpoint_association", g.ProviderName,
				map[string]string{"name": name, "project": proj, "location": loc},
				networkSecurityAllowEmptyValues, networkSecurityAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := nsService.Projects.Locations.InterceptDeploymentGroups.List(parent).Pages(ctx, func(p *networksecurity.ListInterceptDeploymentGroupsResponse) error {
		for _, o := range p.InterceptDeploymentGroups {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_network_security_intercept_deployment_group", g.ProviderName,
				map[string]string{"name": name, "project": proj, "location": loc},
				networkSecurityAllowEmptyValues, networkSecurityAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := nsService.Projects.Locations.InterceptDeployments.List(parent).Pages(ctx, func(p *networksecurity.ListInterceptDeploymentsResponse) error {
		for _, o := range p.InterceptDeployments {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_network_security_intercept_deployment", g.ProviderName,
				map[string]string{"name": name, "project": proj, "location": loc},
				networkSecurityAllowEmptyValues, networkSecurityAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := nsService.Projects.Locations.InterceptEndpointGroups.List(parent).Pages(ctx, func(p *networksecurity.ListInterceptEndpointGroupsResponse) error {
		for _, o := range p.InterceptEndpointGroups {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_network_security_intercept_endpoint_group", g.ProviderName,
				map[string]string{"name": name, "project": proj, "location": loc},
				networkSecurityAllowEmptyValues, networkSecurityAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := nsService.Projects.Locations.InterceptEndpointGroupAssociations.List(parent).Pages(ctx, func(p *networksecurity.ListInterceptEndpointGroupAssociationsResponse) error {
		for _, o := range p.InterceptEndpointGroupAssociations {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_network_security_intercept_endpoint_group_association", g.ProviderName,
				map[string]string{"name": name, "project": proj, "location": loc},
				networkSecurityAllowEmptyValues, networkSecurityAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := nsService.Projects.Locations.MirroringDeploymentGroups.List(parent).Pages(ctx, func(p *networksecurity.ListMirroringDeploymentGroupsResponse) error {
		for _, o := range p.MirroringDeploymentGroups {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_network_security_mirroring_deployment_group", g.ProviderName,
				map[string]string{"name": name, "project": proj, "location": loc},
				networkSecurityAllowEmptyValues, networkSecurityAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := nsService.Projects.Locations.MirroringDeployments.List(parent).Pages(ctx, func(p *networksecurity.ListMirroringDeploymentsResponse) error {
		for _, o := range p.MirroringDeployments {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_network_security_mirroring_deployment", g.ProviderName,
				map[string]string{"name": name, "project": proj, "location": loc},
				networkSecurityAllowEmptyValues, networkSecurityAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := nsService.Projects.Locations.MirroringEndpointGroups.List(parent).Pages(ctx, func(p *networksecurity.ListMirroringEndpointGroupsResponse) error {
		for _, o := range p.MirroringEndpointGroups {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_network_security_mirroring_endpoint_group", g.ProviderName,
				map[string]string{"name": name, "project": proj, "location": loc},
				networkSecurityAllowEmptyValues, networkSecurityAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := nsService.Projects.Locations.MirroringEndpointGroupAssociations.List(parent).Pages(ctx, func(p *networksecurity.ListMirroringEndpointGroupAssociationsResponse) error {
		for _, o := range p.MirroringEndpointGroupAssociations {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_network_security_mirroring_endpoint_group_association", g.ProviderName,
				map[string]string{"name": name, "project": proj, "location": loc},
				networkSecurityAllowEmptyValues, networkSecurityAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := nsService.Projects.Locations.GatewaySecurityPolicies.List(parent).Pages(ctx, func(p *networksecurity.ListGatewaySecurityPoliciesResponse) error {
		for _, o := range p.GatewaySecurityPolicies {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_network_security_gateway_security_policy", g.ProviderName,
				map[string]string{"name": name, "project": proj, "location": loc},
				networkSecurityAllowEmptyValues, networkSecurityAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}
