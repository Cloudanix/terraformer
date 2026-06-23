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
	"google.golang.org/api/vmwareengine/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var vmwareengineAllowEmptyValues = []string{""}

var vmwareengineAdditionalFields = map[string]interface{}{}

type VmwareengineGenerator struct {
	GCPService
}

// Generate TerraformResources from GCP API,
func (g *VmwareengineGenerator) InitResources() error {
	ctx := context.Background()
	vmwareengineService, err := vmwareengine.NewService(ctx)
	if err != nil {
		return err
	}

	project := g.GetArgs()["project"].(string)
	location := g.GetArgs()["region"].(compute.Region).Name

	pcNames := []string{}
	privateCloudsList := vmwareengineService.Projects.Locations.PrivateClouds.List("projects/" + project + "/locations/" + location)
	if err := privateCloudsList.Pages(ctx, func(page *vmwareengine.ListPrivateCloudsResponse) error {
		for _, obj := range page.PrivateClouds {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			pcNames = append(pcNames, name)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_vmwareengine_private_cloud",
				g.ProviderName,
				map[string]string{"name": name, "project": project, "location": location},
				vmwareengineAllowEmptyValues,
				vmwareengineAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	peeringsList := vmwareengineService.Projects.Locations.NetworkPeerings.List("projects/" + project + "/locations/" + location)
	if err := peeringsList.Pages(ctx, func(page *vmwareengine.ListNetworkPeeringsResponse) error {
		for _, obj := range page.NetworkPeerings {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name, name, "google_vmwareengine_network_peering", g.ProviderName,
				map[string]string{"name": name, "project": project, "location": location},
				vmwareengineAllowEmptyValues, vmwareengineAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	venList := vmwareengineService.Projects.Locations.VmwareEngineNetworks.List("projects/" + project + "/locations/" + location)
	if err := venList.Pages(ctx, func(page *vmwareengine.ListVmwareEngineNetworksResponse) error {
		for _, obj := range page.VmwareEngineNetworks {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name, name, "google_vmwareengine_network", g.ProviderName,
				map[string]string{"name": name, "project": project, "location": location},
				vmwareengineAllowEmptyValues, vmwareengineAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	npNames := []string{}
	npList := vmwareengineService.Projects.Locations.NetworkPolicies.List("projects/" + project + "/locations/" + location)
	if err := npList.Pages(ctx, func(page *vmwareengine.ListNetworkPoliciesResponse) error {
		for _, obj := range page.NetworkPolicies {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			npNames = append(npNames, obj.Name)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name, name, "google_vmwareengine_network_policy", g.ProviderName,
				map[string]string{"name": name, "project": project, "location": location},
				vmwareengineAllowEmptyValues, vmwareengineAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	if err := vmwareengineService.Projects.Locations.Datastores.List("projects/"+project+"/locations/"+location).Pages(ctx, func(p *vmwareengine.ListDatastoresResponse) error {
		for _, o := range p.Datastores {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_vmwareengine_datastore", g.ProviderName,
				map[string]string{"name": name, "project": project, "location": location},
				vmwareengineAllowEmptyValues, vmwareengineAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	for _, np := range npNames {
		if err := vmwareengineService.Projects.Locations.NetworkPolicies.ExternalAccessRules.List(np).Pages(ctx, func(p *vmwareengine.ListExternalAccessRulesResponse) error {
			for _, o := range p.ExternalAccessRules {
				t := strings.Split(o.Name, "/")
				name := t[len(t)-1]
				g.Resources = append(g.Resources, terraformutils.NewResource(
					o.Name, name, "google_vmwareengine_external_access_rule", g.ProviderName,
					map[string]string{"name": name, "parent": np, "project": project, "location": location},
					vmwareengineAllowEmptyValues, vmwareengineAdditionalFields))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
	}

	for _, pc := range pcNames {
		pcParent := "projects/" + project + "/locations/" + location + "/privateClouds/" + pc
		if err := vmwareengineService.Projects.Locations.PrivateClouds.ExternalAddresses.List(pcParent).Pages(ctx, func(p *vmwareengine.ListExternalAddressesResponse) error {
			for _, o := range p.ExternalAddresses {
				t := strings.Split(o.Name, "/")
				name := t[len(t)-1]
				g.Resources = append(g.Resources, terraformutils.NewResource(
					o.Name, name, "google_vmwareengine_external_address", g.ProviderName,
					map[string]string{"name": name, "parent": pcParent, "project": project, "location": location},
					vmwareengineAllowEmptyValues, vmwareengineAdditionalFields))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}

		if err := vmwareengineService.Projects.Locations.PrivateClouds.Subnets.List(pcParent).Pages(ctx, func(p *vmwareengine.ListSubnetsResponse) error {
			for _, o := range p.Subnets {
				t := strings.Split(o.Name, "/")
				name := t[len(t)-1]
				g.Resources = append(g.Resources, terraformutils.NewResource(
					o.Name, name, "google_vmwareengine_subnet", g.ProviderName,
					map[string]string{"name": name, "parent": pcParent, "project": project, "location": location},
					vmwareengineAllowEmptyValues, vmwareengineAdditionalFields))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}

		clustersList := vmwareengineService.Projects.Locations.PrivateClouds.Clusters.List(
			"projects/" + project + "/locations/" + location + "/privateClouds/" + pc)
		if err := clustersList.Pages(ctx, func(page *vmwareengine.ListClustersResponse) error {
			for _, obj := range page.Clusters {
				t := strings.Split(obj.Name, "/")
				name := t[len(t)-1]
				g.Resources = append(g.Resources, terraformutils.NewResource(
					obj.Name,
					name,
					"google_vmwareengine_cluster",
					g.ProviderName,
					map[string]string{
						"name":          name,
						"parent":        "projects/" + project + "/locations/" + location + "/privateClouds/" + pc,
						"project":       project,
						"location":      location,
						"private_cloud": pc,
					},
					vmwareengineAllowEmptyValues,
					vmwareengineAdditionalFields,
				))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
	}
	return nil
}
