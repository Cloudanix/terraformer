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

	"google.golang.org/api/accesscontextmanager/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var accessContextManagerAllowEmptyValues = []string{""}

var accessContextManagerAdditionalFields = map[string]interface{}{}

type AccessContextManagerGenerator struct {
	GCPService
}

// Generate TerraformResources from GCP API,
// Access policies are organization-scoped; requires GOOGLE_ORGANIZATION.
func (g *AccessContextManagerGenerator) InitResources() error {
	org, _ := g.GetArgs()["organization"].(string)
	if org == "" {
		log.Println("accessContextManager: GOOGLE_ORGANIZATION not set; skipping org-scoped access policies")
		return nil
	}
	ctx := context.Background()
	acmService, err := accesscontextmanager.NewService(ctx)
	if err != nil {
		return err
	}

	policyNames := []string{}
	policiesList := acmService.AccessPolicies.List().Parent("organizations/" + org)
	if err := policiesList.Pages(ctx, func(page *accesscontextmanager.ListAccessPoliciesResponse) error {
		for _, obj := range page.AccessPolicies {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			policyNames = append(policyNames, obj.Name)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_access_context_manager_access_policy",
				g.ProviderName,
				map[string]string{
					"parent": "organizations/" + org,
					"title":  obj.Title,
				},
				accessContextManagerAllowEmptyValues,
				accessContextManagerAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	for _, policy := range policyNames {
		if err := acmService.AccessPolicies.ServicePerimeters.List(policy).Pages(ctx, func(page *accesscontextmanager.ListServicePerimetersResponse) error {
			for _, obj := range page.ServicePerimeters {
				g.Resources = append(g.Resources, terraformutils.NewResource(
					obj.Name, obj.Name, "google_access_context_manager_service_perimeter", g.ProviderName,
					map[string]string{"name": obj.Name, "parent": policy},
					accessContextManagerAllowEmptyValues, accessContextManagerAdditionalFields))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
		if err := acmService.AccessPolicies.AccessLevels.List(policy).Pages(ctx, func(page *accesscontextmanager.ListAccessLevelsResponse) error {
			for _, obj := range page.AccessLevels {
				g.Resources = append(g.Resources, terraformutils.NewResource(
					obj.Name, obj.Name, "google_access_context_manager_access_level", g.ProviderName,
					map[string]string{"name": obj.Name, "parent": policy},
					accessContextManagerAllowEmptyValues, accessContextManagerAdditionalFields))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
	}
	return nil
}
