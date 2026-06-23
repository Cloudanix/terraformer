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

	"google.golang.org/api/containeranalysis/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var containerAnalysisAllowEmptyValues = []string{""}

var containerAnalysisAdditionalFields = map[string]interface{}{}

type ContainerAnalysisGenerator struct {
	GCPService
}

// Generate TerraformResources from GCP API,
func (g *ContainerAnalysisGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := containeranalysis.NewService(ctx)
	if err != nil {
		return err
	}
	project := g.GetArgs()["project"].(string)
	parent := "projects/" + project

	if err := svc.Projects.Notes.List(parent).Pages(ctx, func(p *containeranalysis.ListNotesResponse) error {
		for _, o := range p.Notes {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_container_analysis_note", g.ProviderName,
				map[string]string{"name": name, "project": project},
				containerAnalysisAllowEmptyValues, containerAnalysisAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	if err := svc.Projects.Occurrences.List(parent).Pages(ctx, func(p *containeranalysis.ListOccurrencesResponse) error {
		for _, o := range p.Occurrences {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_container_analysis_occurrence", g.ProviderName,
				map[string]string{"project": project},
				containerAnalysisAllowEmptyValues, containerAnalysisAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}
