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

	"google.golang.org/api/firestore/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var firestoreAllowEmptyValues = []string{""}

var firestoreAdditionalFields = map[string]interface{}{}

type FirestoreGenerator struct {
	GCPService
}

// Run on databasesList and create for each TerraformResource
func (g FirestoreGenerator) createResources(databasesList *firestore.GoogleFirestoreAdminV1ListDatabasesResponse) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	project := g.GetArgs()["project"].(string)
	for _, obj := range databasesList.Databases {
		t := strings.Split(obj.Name, "/")
		name := t[len(t)-1]
		resources = append(resources, terraformutils.NewResource(
			obj.Name,
			name,
			"google_firestore_database",
			g.ProviderName,
			map[string]string{
				"name":    name,
				"project": project,
			},
			firestoreAllowEmptyValues,
			firestoreAdditionalFields,
		))
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *FirestoreGenerator) InitResources() error {
	ctx := context.Background()
	firestoreService, err := firestore.NewService(ctx)
	if err != nil {
		return err
	}

	databasesList, err := firestoreService.Projects.Databases.List("projects/" + g.GetArgs()["project"].(string)).Do()
	if err != nil {
		return err
	}
	g.Resources = g.createResources(databasesList)
	return nil
}
