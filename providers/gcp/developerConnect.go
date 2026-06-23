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
	"google.golang.org/api/developerconnect/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var developerConnectAllowEmptyValues = []string{""}

var developerConnectAdditionalFields = map[string]interface{}{}

type DeveloperConnectGenerator struct {
	GCPService
}

// Run on connectionsList and create for each TerraformResource
func (g DeveloperConnectGenerator) createResources(ctx context.Context, list *developerconnect.ProjectsLocationsConnectionsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := list.Pages(ctx, func(page *developerconnect.ListConnectionsResponse) error {
		for _, obj := range page.Connections {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_developer_connect_connection",
				g.ProviderName,
				map[string]string{
					"connection_id": name,
					"project":       g.GetArgs()["project"].(string),
					"location":      location,
				},
				developerConnectAllowEmptyValues,
				developerConnectAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *DeveloperConnectGenerator) InitResources() error {
	ctx := context.Background()
	dcService, err := developerconnect.NewService(ctx)
	if err != nil {
		return err
	}

	connectionsList := dcService.Projects.Locations.Connections.List(
		"projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name)
	g.Resources = g.createResources(ctx, connectionsList)
	return nil
}
