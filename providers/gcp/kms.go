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

	"google.golang.org/api/cloudkms/v1"
	"google.golang.org/api/compute/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var kmsAllowEmptyValues = []string{""}

var kmsAdditionalFields = map[string]interface{}{}

type KmsGenerator struct {
	GCPService
}

func (g KmsGenerator) createKmsRingResources(ctx context.Context, keyRingList *cloudkms.ProjectsLocationsKeyRingsListCall, kmsService *cloudkms.Service) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	if err := keyRingList.Pages(ctx, func(page *cloudkms.ListKeyRingsResponse) error {
		for _, obj := range page.KeyRings {
			tm := strings.Split(obj.Name, "/")
			ID := tm[1] + "/" + tm[3] + "/" + tm[5]
			resources = append(resources, terraformutils.NewResource(
				ID,
				tm[len(tm)-3]+"_"+tm[len(tm)-1],
				"google_kms_key_ring",
				g.ProviderName,
				map[string]string{
					"project":  g.GetArgs()["project"].(string),
					"location": tm[3],
					"name":     tm[5],
				},
				kmsAllowEmptyValues,
				kmsAdditionalFields,
			))
			resources = append(resources, g.createKmsKeyResources(ctx, obj.Name, kmsService)...)
			if policy, perr := kmsService.Projects.Locations.KeyRings.GetIamPolicy(obj.Name).Do(); perr == nil {
				for _, b := range policy.Bindings {
					for _, m := range b.Members {
						resources = append(resources, terraformutils.NewResource(
							obj.Name+" "+b.Role+" "+m, tm[5]+"_"+b.Role+"_"+m,
							"google_kms_key_ring_iam_member", g.ProviderName,
							map[string]string{"key_ring_id": obj.Name, "role": b.Role, "member": m, "project": g.GetArgs()["project"].(string)},
							kmsAllowEmptyValues, kmsAdditionalFields))
					}
				}
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

func (g *KmsGenerator) createKmsKeyResources(ctx context.Context, keyRingName string, kmsService *cloudkms.Service) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	keyList := kmsService.Projects.Locations.KeyRings.CryptoKeys.List(keyRingName)
	if err := keyList.Pages(ctx, func(page *cloudkms.ListCryptoKeysResponse) error {
		for _, key := range page.CryptoKeys {
			tm := strings.Split(key.Name, "/")
			resources = append(resources, terraformutils.NewResource(
				key.Name,
				tm[1]+"_"+tm[3]+"_"+tm[5]+"_"+tm[7],
				"google_kms_crypto_key",
				g.ProviderName,
				map[string]string{
					"project": g.GetArgs()["project"].(string),
					"name":    key.Name,
				},
				kmsAllowEmptyValues,
				kmsAdditionalFields,
			))
			resources = append(resources, g.createKmsKeyVersionResources(ctx, key.Name, kmsService)...)
			if policy, perr := kmsService.Projects.Locations.KeyRings.CryptoKeys.GetIamPolicy(key.Name).Do(); perr == nil {
				for _, b := range policy.Bindings {
					for _, m := range b.Members {
						resources = append(resources, terraformutils.NewResource(
							key.Name+" "+b.Role+" "+m, tm[len(tm)-1]+"_"+b.Role+"_"+m,
							"google_kms_crypto_key_iam_member", g.ProviderName,
							map[string]string{"crypto_key_id": key.Name, "role": b.Role, "member": m, "project": g.GetArgs()["project"].(string)},
							kmsAllowEmptyValues, kmsAdditionalFields))
					}
				}
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

func (g *KmsGenerator) createKmsKeyVersionResources(ctx context.Context, keyName string, kmsService *cloudkms.Service) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	versionList := kmsService.Projects.Locations.KeyRings.CryptoKeys.CryptoKeyVersions.List(keyName)
	if err := versionList.Pages(ctx, func(page *cloudkms.ListCryptoKeyVersionsResponse) error {
		for _, version := range page.CryptoKeyVersions {
			tm := strings.Split(version.Name, "/")
			resources = append(resources, terraformutils.NewResource(
				version.Name,
				tm[5]+"_"+tm[7]+"_"+tm[len(tm)-1],
				"google_kms_crypto_key_version",
				g.ProviderName,
				map[string]string{
					"project":    g.GetArgs()["project"].(string),
					"crypto_key": keyName,
				},
				kmsAllowEmptyValues,
				kmsAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *KmsGenerator) InitResources() error {
	ctx := context.Background()
	kmsService, err := cloudkms.NewService(ctx)
	if err != nil {
		return err
	}

	keyRingList := kmsService.Projects.Locations.KeyRings.List("projects/" + g.GetArgs()["project"].(string) + "/locations/global")

	g.Resources = g.createKmsRingResources(ctx, keyRingList, kmsService)

	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	regionParent := "projects/" + project + "/locations/" + region
	if err := kmsService.Projects.Locations.EkmConnections.List(regionParent).Pages(ctx, func(p *cloudkms.ListEkmConnectionsResponse) error {
		for _, o := range p.EkmConnections {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_kms_ekm_connection", g.ProviderName,
				map[string]string{"name": name, "project": project, "location": region},
				kmsAllowEmptyValues, kmsAdditionalFields))
			if policy, perr := kmsService.Projects.Locations.EkmConnections.GetIamPolicy(o.Name).Do(); perr == nil {
				for _, b := range policy.Bindings {
					for _, m := range b.Members {
						g.Resources = append(g.Resources, terraformutils.NewResource(
							o.Name+" "+b.Role+" "+m, name+"_"+b.Role+"_"+m,
							"google_kms_ekm_connection_iam_member", g.ProviderName,
							map[string]string{"name": name, "role": b.Role, "member": m, "project": project, "location": region},
							kmsAllowEmptyValues, kmsAdditionalFields))
					}
				}
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := kmsService.Projects.Locations.KeyHandles.List(regionParent).Pages(ctx, func(p *cloudkms.ListKeyHandlesResponse) error {
		for _, o := range p.KeyHandles {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_kms_key_handle", g.ProviderName,
				map[string]string{"name": name, "project": project, "location": region},
				kmsAllowEmptyValues, kmsAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}

func (g *KmsGenerator) PostConvertHook() error {
	for i, key := range g.Resources {
		if key.InstanceInfo.Type != "google_kms_crypto_key" {
			continue
		}
		for _, keyRing := range g.Resources {
			if keyRing.InstanceInfo.Type != "google_kms_key_ring" {
				continue
			}
			if key.Item["key_ring"] == keyRing.InstanceState.ID {
				g.Resources[i].Item["key_ring"] = "${google_kms_key_ring." + keyRing.ResourceName + ".self_link}"
			}
		}
	}
	return nil
}
