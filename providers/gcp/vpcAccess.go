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
	"google.golang.org/api/vpcaccess/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var vpcAccessAllowEmptyValues = []string{""}

var vpcAccessAdditionalFields = map[string]interface{}{}

type VpcAccessGenerator struct {
	GCPService
}

// Run on connectorsList and create for each TerraformResource
func (g VpcAccessGenerator) createResources(ctx context.Context, connectorsList *vpcaccess.ProjectsLocationsConnectorsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	region := g.GetArgs()["region"].(compute.Region).Name
	if err := connectorsList.Pages(ctx, func(page *vpcaccess.ListConnectorsResponse) error {
		for _, obj := range page.Connectors {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_vpc_access_connector",
				g.ProviderName,
				map[string]string{
					"name":    name,
					"project": g.GetArgs()["project"].(string),
					"region":  region,
				},
				vpcAccessAllowEmptyValues,
				vpcAccessAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *VpcAccessGenerator) InitResources() error {
	ctx := context.Background()
	vpcAccessService, err := vpcaccess.NewService(ctx)
	if err != nil {
		return err
	}

	connectorsList := vpcAccessService.Projects.Locations.Connectors.List(
		"projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name)
	g.Resources = g.createResources(ctx, connectorsList)
	return nil
}
