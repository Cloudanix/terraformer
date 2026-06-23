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

	"google.golang.org/api/firebaserules/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var firebaseRulesAllowEmptyValues = []string{""}

var firebaseRulesAdditionalFields = map[string]interface{}{}

type FirebaseRulesGenerator struct {
	GCPService
}

// Generate TerraformResources from GCP API,
func (g *FirebaseRulesGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := firebaserules.NewService(ctx)
	if err != nil {
		return err
	}
	project := g.GetArgs()["project"].(string)
	parent := "projects/" + project
	tail := func(s string) string { p := strings.Split(s, "/"); return p[len(p)-1] }

	if err := svc.Projects.Rulesets.List(parent).Pages(ctx, func(p *firebaserules.ListRulesetsResponse) error {
		for _, o := range p.Rulesets {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, tail(o.Name), "google_firebaserules_ruleset", g.ProviderName,
				map[string]string{"project": project},
				firebaseRulesAllowEmptyValues, firebaseRulesAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := svc.Projects.Releases.List(parent).Pages(ctx, func(p *firebaserules.ListReleasesResponse) error {
		for _, o := range p.Releases {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, tail(o.Name), "google_firebaserules_release", g.ProviderName,
				map[string]string{"name": tail(o.Name), "project": project},
				firebaseRulesAllowEmptyValues, firebaseRulesAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}
