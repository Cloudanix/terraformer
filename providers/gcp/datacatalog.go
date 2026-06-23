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
	"google.golang.org/api/datacatalog/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var datacatalogAllowEmptyValues = []string{""}

var datacatalogAdditionalFields = map[string]interface{}{}

type DatacatalogGenerator struct {
	GCPService
}

// Run on entryGroupsList and create for each TerraformResource
func (g DatacatalogGenerator) createResources(ctx context.Context, entryGroupsList *datacatalog.ProjectsLocationsEntryGroupsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	region := g.GetArgs()["region"].(compute.Region).Name
	if err := entryGroupsList.Pages(ctx, func(page *datacatalog.GoogleCloudDatacatalogV1ListEntryGroupsResponse) error {
		for _, obj := range page.EntryGroups {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_data_catalog_entry_group",
				g.ProviderName,
				map[string]string{
					"name":    name,
					"project": g.GetArgs()["project"].(string),
					"region":  region,
				},
				datacatalogAllowEmptyValues,
				datacatalogAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *DatacatalogGenerator) InitResources() error {
	ctx := context.Background()
	datacatalogService, err := datacatalog.NewService(ctx)
	if err != nil {
		return err
	}

	parent := "projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name
	entryGroupsList := datacatalogService.Projects.Locations.EntryGroups.List(parent)
	g.Resources = append(g.Resources, g.createResources(ctx, entryGroupsList)...)

	taxNames := []string{}
	if err := datacatalogService.Projects.Locations.Taxonomies.List(parent).Pages(ctx, func(page *datacatalog.GoogleCloudDatacatalogV1ListTaxonomiesResponse) error {
		for _, obj := range page.Taxonomies {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			taxNames = append(taxNames, obj.Name)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name, name, "google_data_catalog_taxonomy", g.ProviderName,
				map[string]string{"name": name, "project": g.GetArgs()["project"].(string), "region": g.GetArgs()["region"].(compute.Region).Name},
				datacatalogAllowEmptyValues, datacatalogAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	for _, tax := range taxNames {
		if err := datacatalogService.Projects.Locations.Taxonomies.PolicyTags.List(tax).Pages(ctx, func(page *datacatalog.GoogleCloudDatacatalogV1ListPolicyTagsResponse) error {
			for _, obj := range page.PolicyTags {
				t := strings.Split(obj.Name, "/")
				name := t[len(t)-1]
				g.Resources = append(g.Resources, terraformutils.NewResource(
					obj.Name, name, "google_data_catalog_policy_tag", g.ProviderName,
					map[string]string{"taxonomy": tax, "project": g.GetArgs()["project"].(string)},
					datacatalogAllowEmptyValues, datacatalogAdditionalFields))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
	}
	return nil
}
