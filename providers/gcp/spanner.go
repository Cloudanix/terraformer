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

	"google.golang.org/api/spanner/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var spannerAllowEmptyValues = []string{""}

var spannerAdditionalFields = map[string]interface{}{}

type SpannerGenerator struct {
	GCPService
}

// Run on instancesList and create for each TerraformResource
func (g SpannerGenerator) createResources(ctx context.Context, instancesList *spanner.ProjectsInstancesListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	project := g.GetArgs()["project"].(string)
	if err := instancesList.Pages(ctx, func(page *spanner.ListInstancesResponse) error {
		for _, obj := range page.Instances {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				project+"/"+name,
				name,
				"google_spanner_instance",
				g.ProviderName,
				map[string]string{
					"name":    name,
					"project": project,
				},
				spannerAllowEmptyValues,
				spannerAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *SpannerGenerator) InitResources() error {
	ctx := context.Background()
	spannerService, err := spanner.NewService(ctx)
	if err != nil {
		return err
	}

	instancesList := spannerService.Projects.Instances.List("projects/" + g.GetArgs()["project"].(string))
	g.Resources = g.createResources(ctx, instancesList)
	return nil
}
