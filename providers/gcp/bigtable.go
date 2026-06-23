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
	"strings"

	"google.golang.org/api/bigtableadmin/v2"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var bigtableAllowEmptyValues = []string{""}

var bigtableAdditionalFields = map[string]interface{}{}

type BigtableGenerator struct {
	GCPService
}

// Run on instancesList and create for each TerraformResource
func (g BigtableGenerator) createResources(instancesList *bigtableadmin.ListInstancesResponse) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	project := g.GetArgs()["project"].(string)
	for _, obj := range instancesList.Instances {
		t := strings.Split(obj.Name, "/")
		name := t[len(t)-1]
		resources = append(resources, terraformutils.NewResource(
			project+"/"+name,
			name,
			"google_bigtable_instance",
			g.ProviderName,
			map[string]string{
				"name":    name,
				"project": project,
			},
			bigtableAllowEmptyValues,
			bigtableAdditionalFields,
		))
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *BigtableGenerator) InitResources() error {
	ctx := context.Background()
	bigtableService, err := bigtableadmin.NewService(ctx)
	if err != nil {
		return err
	}

	instancesList, err := bigtableService.Projects.Instances.List("projects/" + g.GetArgs()["project"].(string)).Do()
	if err != nil {
		return err
	}
	g.Resources = g.createResources(instancesList)
	return nil
}
