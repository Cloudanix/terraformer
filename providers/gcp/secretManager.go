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
	"google.golang.org/api/secretmanager/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var secretManagerAllowEmptyValues = []string{""}

var secretManagerAdditionalFields = map[string]interface{}{}

type SecretManagerGenerator struct {
	GCPService
}

// Run on secretsList and create for each TerraformResource
func (g SecretManagerGenerator) createSecretsResources(ctx context.Context, secretsList *secretmanager.ProjectsSecretsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	if err := secretsList.Pages(ctx, func(page *secretmanager.ListSecretsResponse) error {
		for _, obj := range page.Secrets {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_secret_manager_secret",
				g.ProviderName,
				map[string]string{
					"name":      name,
					"secret_id": name,
					"project":   g.GetArgs()["project"].(string),
				},
				secretManagerAllowEmptyValues,
				secretManagerAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *SecretManagerGenerator) InitResources() error {
	ctx := context.Background()
	secretManagerService, err := secretmanager.NewService(ctx)
	if err != nil {
		return err
	}

	secretsList := secretManagerService.Projects.Secrets.List("projects/" + g.GetArgs()["project"].(string))
	g.Resources = append(g.Resources, g.createSecretsResources(ctx, secretsList)...)

	g.Resources = append(g.Resources, g.createRegionalSecretsResources(ctx, secretManagerService)...)
	return nil
}

// Run on regional secrets and create for each TerraformResource
func (g SecretManagerGenerator) createRegionalSecretsResources(ctx context.Context, svc *secretmanager.Service) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	project := g.GetArgs()["project"].(string)
	location := g.GetArgs()["region"].(compute.Region).Name
	list := svc.Projects.Locations.Secrets.List("projects/" + project + "/locations/" + location)
	if err := list.Pages(ctx, func(page *secretmanager.ListSecretsResponse) error {
		for _, obj := range page.Secrets {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_secret_manager_regional_secret",
				g.ProviderName,
				map[string]string{
					"secret_id": name,
					"location":  location,
					"project":   project,
				},
				secretManagerAllowEmptyValues,
				secretManagerAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}
