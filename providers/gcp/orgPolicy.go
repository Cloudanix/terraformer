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

	"google.golang.org/api/orgpolicy/v2"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var orgPolicyAllowEmptyValues = []string{""}

var orgPolicyAdditionalFields = map[string]interface{}{}

type OrgPolicyGenerator struct {
	GCPService
}

// Generate TerraformResources from GCP API,
func (g *OrgPolicyGenerator) InitResources() error {
	ctx := context.Background()
	orgPolicyService, err := orgpolicy.NewService(ctx)
	if err != nil {
		return err
	}
	project := g.GetArgs()["project"].(string)
	parent := "projects/" + project

	policiesList := orgPolicyService.Projects.Policies.List(parent)
	if err := policiesList.Pages(ctx, func(page *orgpolicy.GoogleCloudOrgpolicyV2ListPoliciesResponse) error {
		for _, obj := range page.Policies {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_org_policy_policy",
				g.ProviderName,
				map[string]string{
					"name":   obj.Name,
					"parent": parent,
				},
				orgPolicyAllowEmptyValues,
				orgPolicyAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}
