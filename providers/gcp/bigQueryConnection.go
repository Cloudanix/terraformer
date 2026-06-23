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

	"google.golang.org/api/bigqueryconnection/v1"
	"google.golang.org/api/compute/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var bigQueryConnectionAllowEmptyValues = []string{""}

var bigQueryConnectionAdditionalFields = map[string]interface{}{}

type BigQueryConnectionGenerator struct {
	GCPService
}

// Run on connectionsList and create for each TerraformResource
func (g BigQueryConnectionGenerator) createResources(ctx context.Context, list *bigqueryconnection.ProjectsLocationsConnectionsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := list.Pages(ctx, func(page *bigqueryconnection.ListConnectionsResponse) error {
		for _, obj := range page.Connections {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_bigquery_connection",
				g.ProviderName,
				map[string]string{
					"connection_id": name,
					"project":       g.GetArgs()["project"].(string),
					"location":      location,
				},
				bigQueryConnectionAllowEmptyValues,
				bigQueryConnectionAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *BigQueryConnectionGenerator) InitResources() error {
	ctx := context.Background()
	bqConnService, err := bigqueryconnection.NewService(ctx)
	if err != nil {
		return err
	}

	project := g.GetArgs()["project"].(string)
	location := g.GetArgs()["region"].(compute.Region).Name
	connectionsList := bqConnService.Projects.Locations.Connections.List(
		"projects/" + project + "/locations/" + location)
	connRes := g.createResources(ctx, connectionsList)
	g.Resources = connRes
	for _, r := range connRes {
		res := r.InstanceState.ID
		short := strings.Split(res, "/")[len(strings.Split(res, "/"))-1]
		if policy, perr := bqConnService.Projects.Locations.Connections.GetIamPolicy(res, &bigqueryconnection.GetIamPolicyRequest{}).Do(); perr == nil {
			for _, b := range policy.Bindings {
				for _, m := range b.Members {
					g.Resources = append(g.Resources, terraformutils.NewResource(
						res+" "+b.Role+" "+m, short+"_"+b.Role+"_"+m,
						"google_bigquery_connection_iam_member", g.ProviderName,
						map[string]string{"connection_id": short, "role": b.Role, "member": m, "project": project, "location": location},
						bigQueryConnectionAllowEmptyValues, bigQueryConnectionAdditionalFields))
				}
			}
		}
	}
	return nil
}
