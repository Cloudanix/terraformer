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

	"google.golang.org/api/binaryauthorization/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var binaryAuthorizationAllowEmptyValues = []string{""}

var binaryAuthorizationAdditionalFields = map[string]interface{}{}

type BinaryAuthorizationGenerator struct {
	GCPService
}

// Run on attestorsList and create for each TerraformResource
func (g BinaryAuthorizationGenerator) createResources(ctx context.Context, attestorsList *binaryauthorization.ProjectsAttestorsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	if err := attestorsList.Pages(ctx, func(page *binaryauthorization.ListAttestorsResponse) error {
		for _, obj := range page.Attestors {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_binary_authorization_attestor",
				g.ProviderName,
				map[string]string{
					"name":    name,
					"project": g.GetArgs()["project"].(string),
				},
				binaryAuthorizationAllowEmptyValues,
				binaryAuthorizationAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *BinaryAuthorizationGenerator) InitResources() error {
	ctx := context.Background()
	binaryAuthorizationService, err := binaryauthorization.NewService(ctx)
	if err != nil {
		return err
	}

	attestorsList := binaryAuthorizationService.Projects.Attestors.List("projects/" + g.GetArgs()["project"].(string))
	g.Resources = g.createResources(ctx, attestorsList)
	return nil
}
