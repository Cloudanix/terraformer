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
	"google.golang.org/api/gkehub/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var gkeHubAllowEmptyValues = []string{""}

var gkeHubAdditionalFields = map[string]interface{}{}

type GkeHubGenerator struct {
	GCPService
}

// Run on membershipsList and create for each TerraformResource
func (g GkeHubGenerator) createResources(ctx context.Context, membershipsList *gkehub.ProjectsLocationsMembershipsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := membershipsList.Pages(ctx, func(page *gkehub.ListMembershipsResponse) error {
		for _, obj := range page.Resources {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_gke_hub_membership",
				g.ProviderName,
				map[string]string{
					"name":          name,
					"membership_id": name,
					"project":       g.GetArgs()["project"].(string),
					"location":      location,
				},
				gkeHubAllowEmptyValues,
				gkeHubAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *GkeHubGenerator) InitResources() error {
	ctx := context.Background()
	gkeHubService, err := gkehub.NewService(ctx)
	if err != nil {
		return err
	}

	parent := "projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name
	project := g.GetArgs()["project"].(string)
	loc := g.GetArgs()["region"].(compute.Region).Name

	membershipsList := gkeHubService.Projects.Locations.Memberships.List(parent)
	membershipRes := g.createResources(ctx, membershipsList)
	g.Resources = append(g.Resources, membershipRes...)
	for _, r := range membershipRes {
		res := r.InstanceState.ID
		short := strings.Split(res, "/")[len(strings.Split(res, "/"))-1]
		if policy, perr := gkeHubService.Projects.Locations.Memberships.GetIamPolicy(res).Do(); perr == nil {
			for _, b := range policy.Bindings {
				for _, m := range b.Members {
					g.Resources = append(g.Resources, terraformutils.NewResource(
						res+" "+b.Role+" "+m, short+"_"+b.Role+"_"+m,
						"google_gke_hub_membership_iam_member", g.ProviderName,
						map[string]string{"membership_id": short, "role": b.Role, "member": m, "project": project, "location": loc},
						gkeHubAllowEmptyValues, gkeHubAdditionalFields))
				}
			}
		}
	}

	featuresList := gkeHubService.Projects.Locations.Features.List(parent)
	featureRes := g.createFeaturesResources(ctx, featuresList)
	g.Resources = append(g.Resources, featureRes...)
	for _, r := range featureRes {
		res := r.InstanceState.ID
		short := strings.Split(res, "/")[len(strings.Split(res, "/"))-1]
		if policy, perr := gkeHubService.Projects.Locations.Features.GetIamPolicy(res).Do(); perr == nil {
			for _, b := range policy.Bindings {
				for _, m := range b.Members {
					g.Resources = append(g.Resources, terraformutils.NewResource(
						res+" "+b.Role+" "+m, short+"_"+b.Role+"_"+m,
						"google_gke_hub_feature_iam_member", g.ProviderName,
						map[string]string{"name": short, "role": b.Role, "member": m, "project": project, "location": loc},
						gkeHubAllowEmptyValues, gkeHubAdditionalFields))
				}
			}
		}
	}

	scopeNames := []string{}
	scopesList := gkeHubService.Projects.Locations.Scopes.List(parent)
	if err := scopesList.Pages(ctx, func(page *gkehub.ListScopesResponse) error {
		for _, obj := range page.Scopes {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			scopeNames = append(scopeNames, obj.Name)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name, name, "google_gke_hub_scope", g.ProviderName,
				map[string]string{"scope_id": name, "project": g.GetArgs()["project"].(string)},
				gkeHubAllowEmptyValues, gkeHubAdditionalFields,
			))
			if policy, perr := gkeHubService.Projects.Locations.Scopes.GetIamPolicy(obj.Name).Do(); perr == nil {
				for _, b := range policy.Bindings {
					for _, m := range b.Members {
						g.Resources = append(g.Resources, terraformutils.NewResource(
							obj.Name+" "+b.Role+" "+m, name+"_"+b.Role+"_"+m,
							"google_gke_hub_scope_iam_member", g.ProviderName,
							map[string]string{"scope_id": name, "role": b.Role, "member": m, "project": g.GetArgs()["project"].(string)},
							gkeHubAllowEmptyValues, gkeHubAdditionalFields))
					}
				}
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	for _, scope := range scopeNames {
		scopeID := strings.Split(scope, "/")[len(strings.Split(scope, "/"))-1]
		if err := gkeHubService.Projects.Locations.Scopes.Namespaces.List(scope).Pages(ctx, func(p *gkehub.ListScopeNamespacesResponse) error {
			for _, o := range p.ScopeNamespaces {
				t := strings.Split(o.Name, "/")
				g.Resources = append(g.Resources, terraformutils.NewResource(
					o.Name, t[len(t)-1], "google_gke_hub_namespace", g.ProviderName,
					map[string]string{"scope_id": scopeID, "project": g.GetArgs()["project"].(string)},
					gkeHubAllowEmptyValues, gkeHubAdditionalFields))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
		if err := gkeHubService.Projects.Locations.Scopes.Rbacrolebindings.List(scope).Pages(ctx, func(p *gkehub.ListScopeRBACRoleBindingsResponse) error {
			for _, o := range p.Rbacrolebindings {
				t := strings.Split(o.Name, "/")
				g.Resources = append(g.Resources, terraformutils.NewResource(
					o.Name, t[len(t)-1], "google_gke_hub_scope_rbac_role_binding", g.ProviderName,
					map[string]string{"scope_id": scopeID, "project": g.GetArgs()["project"].(string)},
					gkeHubAllowEmptyValues, gkeHubAdditionalFields))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
	}

	fleetsList := gkeHubService.Projects.Locations.Fleets.List(parent)
	if err := fleetsList.Pages(ctx, func(page *gkehub.ListFleetsResponse) error {
		for _, obj := range page.Fleets {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name, name, "google_gke_hub_fleet", g.ProviderName,
				map[string]string{"project": g.GetArgs()["project"].(string)},
				gkeHubAllowEmptyValues, gkeHubAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}

// Run on featuresList and create for each TerraformResource (response field is Resources)
func (g GkeHubGenerator) createFeaturesResources(ctx context.Context, list *gkehub.ProjectsLocationsFeaturesListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := list.Pages(ctx, func(page *gkehub.ListFeaturesResponse) error {
		for _, obj := range page.Resources {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_gke_hub_feature",
				g.ProviderName,
				map[string]string{
					"name":     name,
					"project":  g.GetArgs()["project"].(string),
					"location": location,
				},
				gkeHubAllowEmptyValues,
				gkeHubAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}
