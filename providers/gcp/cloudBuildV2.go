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

	"google.golang.org/api/cloudbuild/v2"
	"google.golang.org/api/compute/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var cloudBuildV2AllowEmptyValues = []string{""}

var cloudBuildV2AdditionalFields = map[string]interface{}{}

type CloudBuildV2Generator struct {
	GCPService
}

// Generate TerraformResources from GCP API,
func (g *CloudBuildV2Generator) InitResources() error {
	ctx := context.Background()
	svc, err := cloudbuild.NewService(ctx)
	if err != nil {
		return err
	}
	location := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + g.GetArgs()["project"].(string) + "/locations/" + location

	if err := svc.Projects.Locations.Connections.List(parent).Pages(ctx, func(page *cloudbuild.ListConnectionsResponse) error {
		for _, obj := range page.Connections {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name, name, "google_cloudbuildv2_connection", g.ProviderName,
				map[string]string{"name": name, "project": g.GetArgs()["project"].(string), "location": location},
				cloudBuildV2AllowEmptyValues, cloudBuildV2AdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}
