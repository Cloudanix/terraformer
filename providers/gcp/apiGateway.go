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

	"google.golang.org/api/apigateway/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var apiGatewayAllowEmptyValues = []string{""}

var apiGatewayAdditionalFields = map[string]interface{}{}

type APIGatewayGenerator struct {
	GCPService
}

// Run on apisList and create for each TerraformResource.
// google_api_gateway_* are beta-only; refresh with --provider-type beta.
func (g APIGatewayGenerator) createResources(ctx context.Context, list *apigateway.ProjectsLocationsApisListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	if err := list.Pages(ctx, func(page *apigateway.ApigatewayListApisResponse) error {
		for _, obj := range page.Apis {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_api_gateway_api",
				g.ProviderName,
				map[string]string{
					"api_id":  name,
					"project": g.GetArgs()["project"].(string),
				},
				apiGatewayAllowEmptyValues,
				apiGatewayAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *APIGatewayGenerator) InitResources() error {
	ctx := context.Background()
	apiGatewayService, err := apigateway.NewService(ctx)
	if err != nil {
		return err
	}

	// APIs are global (locations/global).
	apisList := apiGatewayService.Projects.Locations.Apis.List(
		"projects/" + g.GetArgs()["project"].(string) + "/locations/global")
	g.Resources = g.createResources(ctx, apisList)
	return nil
}
