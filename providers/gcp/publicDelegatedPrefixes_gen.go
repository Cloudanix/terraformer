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

// AUTO-GENERATED CODE. DO NOT EDIT.
package gcp

import (
	"context"
	"log"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"

	"google.golang.org/api/compute/v1"
)

var publicDelegatedPrefixesAllowEmptyValues = []string{""}

var publicDelegatedPrefixesAdditionalFields = map[string]interface{}{}

type PublicDelegatedPrefixesGenerator struct {
	GCPService
}

// Run on publicDelegatedPrefixesList and create for each TerraformResource
func (g PublicDelegatedPrefixesGenerator) createResources(ctx context.Context, publicDelegatedPrefixesList *compute.PublicDelegatedPrefixesListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	if err := publicDelegatedPrefixesList.Pages(ctx, func(page *compute.PublicDelegatedPrefixList) error {
		for _, obj := range page.Items {
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				obj.Name,
				"google_compute_public_delegated_prefix",
				g.ProviderName,
				map[string]string{
					"name":    obj.Name,
					"project": g.GetArgs()["project"].(string),
					"region":  g.GetArgs()["region"].(compute.Region).Name,
				},
				publicDelegatedPrefixesAllowEmptyValues,
				publicDelegatedPrefixesAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
// from each publicDelegatedPrefixes create 1 TerraformResource
// Need publicDelegatedPrefixes name as ID for terraform resource
func (g *PublicDelegatedPrefixesGenerator) InitResources() error {
	ctx := context.Background()
	computeService, err := compute.NewService(ctx)
	if err != nil {
		return err
	}

	publicDelegatedPrefixesList := computeService.PublicDelegatedPrefixes.List(g.GetArgs()["project"].(string), g.GetArgs()["region"].(compute.Region).Name)
	g.Resources = g.createResources(ctx, publicDelegatedPrefixesList)

	return nil

}
