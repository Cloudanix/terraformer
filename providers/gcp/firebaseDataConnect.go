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
	"google.golang.org/api/firebasedataconnect/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var firebaseDataConnectAllowEmptyValues = []string{""}

var firebaseDataConnectAdditionalFields = map[string]interface{}{}

type FirebaseDataConnectGenerator struct {
	GCPService
}

// Generate TerraformResources from GCP API,
func (g *FirebaseDataConnectGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := firebasedataconnect.NewService(ctx)
	if err != nil {
		return err
	}
	project := g.GetArgs()["project"].(string)
	location := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + location

	if err := svc.Projects.Locations.Services.List(parent).Pages(ctx, func(p *firebasedataconnect.ListServicesResponse) error {
		for _, o := range p.Services {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_firebase_data_connect_service", g.ProviderName,
				map[string]string{"service_id": name, "location": location, "project": project},
				firebaseDataConnectAllowEmptyValues, firebaseDataConnectAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}
