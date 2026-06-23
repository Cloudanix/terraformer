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

	project := g.GetArgs()["project"].(string)
	databasesList, err := firestoreService.Projects.Databases.List("projects/" + project).Do()
	if err != nil {
		return err
	}
	g.Resources = g.createResources(databasesList)

	// Walk each database for its backup schedules.
	for _, db := range databasesList.Databases {
		bs, berr := firestoreService.Projects.Databases.BackupSchedules.List(db.Name).Do()
		if berr != nil {
			continue
		}
		for _, obj := range bs.BackupSchedules {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			dt := strings.Split(db.Name, "/")
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name, name, "google_firestore_backup_schedule", g.ProviderName,
				map[string]string{"name": name, "database": dt[len(dt)-1], "project": project},
				firestoreAllowEmptyValues, firestoreAdditionalFields))
		}

		dt := strings.Split(db.Name, "/")
		dbID := dt[len(dt)-1]
		if ierr := firestoreService.Projects.Databases.CollectionGroups.Indexes.List(db.Name+"/collectionGroups/-").Pages(ctx, func(p *firestore.GoogleFirestoreAdminV1ListIndexesResponse) error {
			for _, idx := range p.Indexes {
				it := strings.Split(idx.Name, "/")
				g.Resources = append(g.Resources, terraformutils.NewResource(
					idx.Name, it[len(it)-1], "google_firestore_index", g.ProviderName,
					map[string]string{"database": dbID, "collection": it[len(it)-3], "project": project},
					firestoreAllowEmptyValues, firestoreAdditionalFields))
			}
			return nil
		}); ierr != nil {
			log.Println(ierr)
		}
		if ferr := firestoreService.Projects.Databases.CollectionGroups.Fields.List(db.Name+"/collectionGroups/-").Pages(ctx, func(p *firestore.GoogleFirestoreAdminV1ListFieldsResponse) error {
			for _, fld := range p.Fields {
				ft := strings.Split(fld.Name, "/")
				g.Resources = append(g.Resources, terraformutils.NewResource(
					fld.Name, ft[len(ft)-1], "google_firestore_field", g.ProviderName,
					map[string]string{"database": dbID, "collection": ft[len(ft)-3], "field": ft[len(ft)-1], "project": project},
					firestoreAllowEmptyValues, firestoreAdditionalFields))
			}
			return nil
		}); ferr != nil {
			log.Println(ferr)
		}
		if uc, uerr := firestoreService.Projects.Databases.UserCreds.List(db.Name).Do(); uerr == nil {
			for _, u := range uc.UserCreds {
				ut := strings.Split(u.Name, "/")
				g.Resources = append(g.Resources, terraformutils.NewResource(
					u.Name, ut[len(ut)-1], "google_firestore_user_creds", g.ProviderName,
					map[string]string{"user_creds_id": ut[len(ut)-1], "database": dbID, "project": project},
					firestoreAllowEmptyValues, firestoreAdditionalFields))
			}
		}
	}
	return nil
}
