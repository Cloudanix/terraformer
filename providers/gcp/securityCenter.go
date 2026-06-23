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

	"google.golang.org/api/securitycenter/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var securityCenterAllowEmptyValues = []string{""}

var securityCenterAdditionalFields = map[string]interface{}{}

type SecurityCenterGenerator struct {
	GCPService
}

// Generate TerraformResources from GCP API,
// SCC sources are organization-scoped; requires GOOGLE_ORGANIZATION to be set.
func (g *SecurityCenterGenerator) InitResources() error {
	org, _ := g.GetArgs()["organization"].(string)
	if org == "" {
		log.Println("securityCenter: GOOGLE_ORGANIZATION not set; skipping org-scoped SCC sources")
		return nil
	}
	ctx := context.Background()
	sccService, err := securitycenter.NewService(ctx)
	if err != nil {
		return err
	}

	sourcesList := sccService.Organizations.Sources.List("organizations/" + org)
	if err := sourcesList.Pages(ctx, func(page *securitycenter.ListSourcesResponse) error {
		for _, obj := range page.Sources {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_scc_source",
				g.ProviderName,
				map[string]string{
					"organization": org,
				},
				securityCenterAllowEmptyValues,
				securityCenterAdditionalFields,
			))
			if policy, perr := sccService.Organizations.Sources.GetIamPolicy(obj.Name, &securitycenter.GetIamPolicyRequest{}).Do(); perr == nil {
				for _, b := range policy.Bindings {
					for _, m := range b.Members {
						g.Resources = append(g.Resources, terraformutils.NewResource(
							obj.Name+" "+b.Role+" "+m, name+"_"+b.Role+"_"+m,
							"google_scc_source_iam_member", g.ProviderName,
							map[string]string{"source": obj.Name, "organization": org, "role": b.Role, "member": m},
							securityCenterAllowEmptyValues, securityCenterAdditionalFields))
					}
				}
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	ncList := sccService.Organizations.NotificationConfigs.List("organizations/" + org)
	if err := ncList.Pages(ctx, func(page *securitycenter.ListNotificationConfigsResponse) error {
		for _, obj := range page.NotificationConfigs {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_scc_notification_config",
				g.ProviderName,
				map[string]string{
					"organization": org,
					"config_id":    name,
				},
				securityCenterAllowEmptyValues,
				securityCenterAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}
