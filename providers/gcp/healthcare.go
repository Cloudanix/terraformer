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
	"google.golang.org/api/healthcare/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var healthcareAllowEmptyValues = []string{""}

var healthcareAdditionalFields = map[string]interface{}{}

type HealthcareGenerator struct {
	GCPService
}

// Run on fhirStoresList and create for each TerraformResource (per-dataset walk)
func (g HealthcareGenerator) createFhirStoresResources(ctx context.Context, list *healthcare.ProjectsLocationsDatasetsFhirStoresService, datasetName string) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	if err := list.List(datasetName).Pages(ctx, func(page *healthcare.ListFhirStoresResponse) error {
		for _, obj := range page.FhirStores {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_healthcare_fhir_store",
				g.ProviderName,
				map[string]string{
					"name":    name,
					"dataset": datasetName,
					"project": g.GetArgs()["project"].(string),
				},
				healthcareAllowEmptyValues,
				healthcareAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Run on dicomStoresList and create for each TerraformResource (per-dataset walk)
func (g HealthcareGenerator) createDicomStoresResources(ctx context.Context, list *healthcare.ProjectsLocationsDatasetsDicomStoresService, datasetName string) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	if err := list.List(datasetName).Pages(ctx, func(page *healthcare.ListDicomStoresResponse) error {
		for _, obj := range page.DicomStores {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_healthcare_dicom_store",
				g.ProviderName,
				map[string]string{
					"name":    name,
					"dataset": datasetName,
					"project": g.GetArgs()["project"].(string),
				},
				healthcareAllowEmptyValues,
				healthcareAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Run on hl7V2StoresList and create for each TerraformResource (per-dataset walk)
func (g HealthcareGenerator) createHl7V2StoresResources(ctx context.Context, list *healthcare.ProjectsLocationsDatasetsHl7V2StoresService, datasetName string) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	if err := list.List(datasetName).Pages(ctx, func(page *healthcare.ListHl7V2StoresResponse) error {
		for _, obj := range page.Hl7V2Stores {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_healthcare_hl7_v2_store",
				g.ProviderName,
				map[string]string{
					"name":    name,
					"dataset": datasetName,
					"project": g.GetArgs()["project"].(string),
				},
				healthcareAllowEmptyValues,
				healthcareAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *HealthcareGenerator) InitResources() error {
	ctx := context.Background()
	healthcareService, err := healthcare.NewService(ctx)
	if err != nil {
		return err
	}
	project := g.GetArgs()["project"].(string)
	location := g.GetArgs()["region"].(compute.Region).Name

	datasetNames := []string{}
	datasetsList := healthcareService.Projects.Locations.Datasets.List("projects/" + project + "/locations/" + location)
	if err := datasetsList.Pages(ctx, func(page *healthcare.ListDatasetsResponse) error {
		for _, obj := range page.Datasets {
			datasetNames = append(datasetNames, obj.Name)
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_healthcare_dataset",
				g.ProviderName,
				map[string]string{"name": name, "project": project, "location": location},
				healthcareAllowEmptyValues,
				healthcareAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	for _, dataset := range datasetNames {
		g.Resources = append(g.Resources, g.createFhirStoresResources(ctx, healthcareService.Projects.Locations.Datasets.FhirStores, dataset)...)
		g.Resources = append(g.Resources, g.createDicomStoresResources(ctx, healthcareService.Projects.Locations.Datasets.DicomStores, dataset)...)
		g.Resources = append(g.Resources, g.createHl7V2StoresResources(ctx, healthcareService.Projects.Locations.Datasets.Hl7V2Stores, dataset)...)
	}
	return nil
}
