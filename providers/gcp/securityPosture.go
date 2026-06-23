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

	"google.golang.org/api/securityposture/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var securityPostureAllowEmptyValues = []string{""}

var securityPostureAdditionalFields = map[string]interface{}{}

type SecurityPostureGenerator struct {
	GCPService
}

// Generate TerraformResources from GCP API,
// Postures are organization-scoped; requires GOOGLE_ORGANIZATION.
func (g *SecurityPostureGenerator) InitResources() error {
	org, _ := g.GetArgs()["organization"].(string)
	if org == "" {
		log.Println("securityPosture: GOOGLE_ORGANIZATION not set; skipping org-scoped postures")
		return nil
	}
	ctx := context.Background()
	svc, err := securityposture.NewService(ctx)
	if err != nil {
		return err
	}

	if err := svc.Organizations.Locations.Postures.List("organizations/"+org+"/locations/global").Pages(ctx, func(page *securityposture.ListPosturesResponse) error {
		for _, obj := range page.Postures {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name, name, "google_securityposture_posture", g.ProviderName,
				map[string]string{"posture_id": name, "parent": "organizations/" + org, "location": "global"},
				securityPostureAllowEmptyValues, securityPostureAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}
