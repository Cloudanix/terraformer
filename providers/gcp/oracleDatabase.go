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
	"google.golang.org/api/oracledatabase/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var oracleDatabaseAllowEmptyValues = []string{""}

var oracleDatabaseAdditionalFields = map[string]interface{}{}

type OracleDatabaseGenerator struct {
	GCPService
}

// Run on infraList and create for each TerraformResource
func (g OracleDatabaseGenerator) createResources(ctx context.Context, infraList *oracledatabase.ProjectsLocationsCloudExadataInfrastructuresListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := infraList.Pages(ctx, func(page *oracledatabase.ListCloudExadataInfrastructuresResponse) error {
		for _, obj := range page.CloudExadataInfrastructures {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_oracle_database_cloud_exadata_infrastructure",
				g.ProviderName,
				map[string]string{
					"name":     name,
					"project":  g.GetArgs()["project"].(string),
					"location": location,
				},
				oracleDatabaseAllowEmptyValues,
				oracleDatabaseAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *OracleDatabaseGenerator) InitResources() error {
	ctx := context.Background()
	oracleDatabaseService, err := oracledatabase.NewService(ctx)
	if err != nil {
		return err
	}

	infraList := oracleDatabaseService.Projects.Locations.CloudExadataInfrastructures.List(
		"projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name)
	g.Resources = g.createResources(ctx, infraList)
	return nil
}
