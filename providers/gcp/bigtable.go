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

// Generate TerraformResources from GCP API,
func (g *BigtableGenerator) InitResources() error {
	ctx := context.Background()
	bigtableService, err := bigtableadmin.NewService(ctx)
	if err != nil {
		return err
	}
	project := g.GetArgs()["project"].(string)

	instancesList, err := bigtableService.Projects.Instances.List("projects/" + project).Do()
	if err != nil {
		return err
	}

	for _, obj := range instancesList.Instances {
		t := strings.Split(obj.Name, "/")
		instanceName := t[len(t)-1]
		g.Resources = append(g.Resources, terraformutils.NewResource(
			project+"/"+instanceName,
			instanceName,
			"google_bigtable_instance",
			g.ProviderName,
			map[string]string{"name": instanceName, "project": project},
			bigtableAllowEmptyValues,
			bigtableAdditionalFields,
		))

		// Walk the instance for its tables.
		tablesList, terr := bigtableService.Projects.Instances.Tables.List(obj.Name).Do()
		if terr == nil {
			for _, tbl := range tablesList.Tables {
				tt := strings.Split(tbl.Name, "/")
				tableName := tt[len(tt)-1]
				g.Resources = append(g.Resources, terraformutils.NewResource(
					project+"/"+instanceName+"/"+tableName,
					tableName,
					"google_bigtable_table",
					g.ProviderName,
					map[string]string{
						"name":          tableName,
						"instance_name": instanceName,
						"project":       project,
					},
					bigtableAllowEmptyValues,
					bigtableAdditionalFields,
				))
			}
		}

		// Walk the instance for its app profiles.
		profilesList, perr := bigtableService.Projects.Instances.AppProfiles.List(obj.Name).Do()
		if perr == nil {
			for _, prof := range profilesList.AppProfiles {
				pt := strings.Split(prof.Name, "/")
				profName := pt[len(pt)-1]
				g.Resources = append(g.Resources, terraformutils.NewResource(
					project+"/"+instanceName+"/"+profName,
					profName,
					"google_bigtable_app_profile",
					g.ProviderName,
					map[string]string{
						"app_profile_id": profName,
						"instance":       instanceName,
						"project":        project,
					},
					bigtableAllowEmptyValues,
					bigtableAdditionalFields,
				))
			}
		}
	}
	return nil
}
