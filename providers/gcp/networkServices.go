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
	"google.golang.org/api/networkservices/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var networkServicesAllowEmptyValues = []string{""}

var networkServicesAdditionalFields = map[string]interface{}{}

type NetworkServicesGenerator struct {
	GCPService
}

// Run on meshesList and create for each TerraformResource
func (g NetworkServicesGenerator) createResources(ctx context.Context, meshesList *networkservices.ProjectsLocationsMeshesListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := meshesList.Pages(ctx, func(page *networkservices.ListMeshesResponse) error {
		for _, obj := range page.Meshes {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_network_services_mesh",
				g.ProviderName,
				map[string]string{
					"name":     name,
					"project":  g.GetArgs()["project"].(string),
					"location": location,
				},
				networkServicesAllowEmptyValues,
				networkServicesAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Run on gatewaysList and create for each TerraformResource
func (g NetworkServicesGenerator) createGatewaysResources(ctx context.Context, list *networkservices.ProjectsLocationsGatewaysListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := list.Pages(ctx, func(page *networkservices.ListGatewaysResponse) error {
		for _, obj := range page.Gateways {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_network_services_gateway",
				g.ProviderName,
				map[string]string{
					"name":     name,
					"project":  g.GetArgs()["project"].(string),
					"location": location,
				},
				networkServicesAllowEmptyValues,
				networkServicesAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *NetworkServicesGenerator) InitResources() error {
	ctx := context.Background()
	networkServicesService, err := networkservices.NewService(ctx)
	if err != nil {
		return err
	}

	parent := "projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name
	meshesList := networkServicesService.Projects.Locations.Meshes.List(parent)
	g.Resources = append(g.Resources, g.createResources(ctx, meshesList)...)

	gatewaysList := networkServicesService.Projects.Locations.Gateways.List(parent)
	g.Resources = append(g.Resources, g.createGatewaysResources(ctx, gatewaysList)...)

	httpRoutesList := networkServicesService.Projects.Locations.HttpRoutes.List(parent)
	g.Resources = append(g.Resources, g.createHTTPRoutesResources(ctx, httpRoutesList)...)

	tcpRoutesList := networkServicesService.Projects.Locations.TcpRoutes.List(parent)
	g.Resources = append(g.Resources, g.createTCPRoutesResources(ctx, tcpRoutesList)...)

	grpcRoutesList := networkServicesService.Projects.Locations.GrpcRoutes.List(parent)
	g.Resources = append(g.Resources, g.createGRPCRoutesResources(ctx, grpcRoutesList)...)

	tlsRoutesList := networkServicesService.Projects.Locations.TlsRoutes.List(parent)
	g.Resources = append(g.Resources, g.createTLSRoutesResources(ctx, tlsRoutesList)...)

	bindingsList := networkServicesService.Projects.Locations.ServiceBindings.List(parent)
	g.Resources = append(g.Resources, g.createServiceBindingsResources(ctx, bindingsList)...)

	endpointPoliciesList := networkServicesService.Projects.Locations.EndpointPolicies.List(parent)
	g.Resources = append(g.Resources, g.createEndpointPoliciesResources(ctx, endpointPoliciesList)...)

	loc := g.GetArgs()["region"].(compute.Region).Name
	if err := networkServicesService.Projects.Locations.WasmPlugins.List(parent).Pages(ctx, func(p *networkservices.ListWasmPluginsResponse) error {
		for _, o := range p.WasmPlugins {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_network_services_wasm_plugin", g.ProviderName,
				map[string]string{"name": name, "project": g.GetArgs()["project"].(string), "location": loc},
				networkServicesAllowEmptyValues, networkServicesAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}

func (g NetworkServicesGenerator) createTLSRoutesResources(ctx context.Context, list *networkservices.ProjectsLocationsTlsRoutesListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := list.Pages(ctx, func(page *networkservices.ListTlsRoutesResponse) error {
		for _, obj := range page.TlsRoutes {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name, name, "google_network_services_tls_route", g.ProviderName,
				map[string]string{"name": name, "project": g.GetArgs()["project"].(string), "location": location},
				networkServicesAllowEmptyValues, networkServicesAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

func (g NetworkServicesGenerator) createServiceBindingsResources(ctx context.Context, list *networkservices.ProjectsLocationsServiceBindingsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := list.Pages(ctx, func(page *networkservices.ListServiceBindingsResponse) error {
		for _, obj := range page.ServiceBindings {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name, name, "google_network_services_service_binding", g.ProviderName,
				map[string]string{"name": name, "project": g.GetArgs()["project"].(string), "location": location},
				networkServicesAllowEmptyValues, networkServicesAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

func (g NetworkServicesGenerator) createEndpointPoliciesResources(ctx context.Context, list *networkservices.ProjectsLocationsEndpointPoliciesListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := list.Pages(ctx, func(page *networkservices.ListEndpointPoliciesResponse) error {
		for _, obj := range page.EndpointPolicies {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name, name, "google_network_services_endpoint_policy", g.ProviderName,
				map[string]string{"name": name, "project": g.GetArgs()["project"].(string), "location": location},
				networkServicesAllowEmptyValues, networkServicesAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Run on grpcRoutesList and create for each TerraformResource
func (g NetworkServicesGenerator) createGRPCRoutesResources(ctx context.Context, list *networkservices.ProjectsLocationsGrpcRoutesListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := list.Pages(ctx, func(page *networkservices.ListGrpcRoutesResponse) error {
		for _, obj := range page.GrpcRoutes {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name, name, "google_network_services_grpc_route", g.ProviderName,
				map[string]string{"name": name, "project": g.GetArgs()["project"].(string), "location": location},
				networkServicesAllowEmptyValues, networkServicesAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Run on tcpRoutesList and create for each TerraformResource
func (g NetworkServicesGenerator) createTCPRoutesResources(ctx context.Context, list *networkservices.ProjectsLocationsTcpRoutesListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := list.Pages(ctx, func(page *networkservices.ListTcpRoutesResponse) error {
		for _, obj := range page.TcpRoutes {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name, name, "google_network_services_tcp_route", g.ProviderName,
				map[string]string{"name": name, "project": g.GetArgs()["project"].(string), "location": location},
				networkServicesAllowEmptyValues, networkServicesAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Run on httpRoutesList and create for each TerraformResource
func (g NetworkServicesGenerator) createHTTPRoutesResources(ctx context.Context, list *networkservices.ProjectsLocationsHttpRoutesListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := list.Pages(ctx, func(page *networkservices.ListHttpRoutesResponse) error {
		for _, obj := range page.HttpRoutes {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_network_services_http_route",
				g.ProviderName,
				map[string]string{
					"name":     name,
					"project":  g.GetArgs()["project"].(string),
					"location": location,
				},
				networkServicesAllowEmptyValues,
				networkServicesAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}
