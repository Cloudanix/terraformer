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
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"

	"google.golang.org/api/compute/v1"
)

var storagePoolsAllowEmptyValues = []string{""}

var storagePoolsAdditionalFields = map[string]interface{}{}

type StoragePoolsGenerator struct {
	GCPService
}

// Run on storagePoolsList and create for each TerraformResource
func (g StoragePoolsGenerator) createResources(ctx context.Context, storagePoolsList *compute.StoragePoolsListCall, zone string) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	if err := storagePoolsList.Pages(ctx, func(page *compute.StoragePoolList) error {
		for _, obj := range page.Items {
			resources = append(resources, terraformutils.NewResource(
				zone+"/"+obj.Name,
				zone+"/"+obj.Name,
				"google_compute_storage_pool",
				g.ProviderName,
				map[string]string{
					"name":    obj.Name,
					"project": g.GetArgs()["project"].(string),
					"region":  g.GetArgs()["region"].(compute.Region).Name,
					"zone":    zone,
				},
				storagePoolsAllowEmptyValues,
				storagePoolsAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
// from each storagePools create 1 TerraformResource
// Need storagePools name as ID for terraform resource
func (g *StoragePoolsGenerator) InitResources() error {
	ctx := context.Background()
	computeService, err := compute.NewService(ctx)
	if err != nil {
		return err
	}

	for _, zoneLink := range g.GetArgs()["region"].(compute.Region).Zones {
		t := strings.Split(zoneLink, "/")
		zone := t[len(t)-1]
		storagePoolsList := computeService.StoragePools.List(g.GetArgs()["project"].(string), zone)
		g.Resources = append(g.Resources, g.createResources(ctx, storagePoolsList, zone)...)
	}

	return nil

}
