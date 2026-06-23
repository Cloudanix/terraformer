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

	"google.golang.org/api/essentialcontacts/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var essentialContactsAllowEmptyValues = []string{""}

var essentialContactsAdditionalFields = map[string]interface{}{}

type EssentialContactsGenerator struct {
	GCPService
}

// Run on contactsList and create for each TerraformResource
func (g EssentialContactsGenerator) createResources(ctx context.Context, contactsList *essentialcontacts.ProjectsContactsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	project := g.GetArgs()["project"].(string)
	if err := contactsList.Pages(ctx, func(page *essentialcontacts.GoogleCloudEssentialcontactsV1ListContactsResponse) error {
		for _, obj := range page.Contacts {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_essential_contacts_contact",
				g.ProviderName,
				map[string]string{
					"name":    name,
					"parent":  "projects/" + project,
					"project": project,
				},
				essentialContactsAllowEmptyValues,
				essentialContactsAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *EssentialContactsGenerator) InitResources() error {
	ctx := context.Background()
	essentialContactsService, err := essentialcontacts.NewService(ctx)
	if err != nil {
		return err
	}

	contactsList := essentialContactsService.Projects.Contacts.List("projects/" + g.GetArgs()["project"].(string))
	g.Resources = g.createResources(ctx, contactsList)
	return nil
}
