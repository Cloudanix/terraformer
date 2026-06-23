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

	"google.golang.org/api/managedidentities/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var activeDirectoryAllowEmptyValues = []string{""}

var activeDirectoryAdditionalFields = map[string]interface{}{}

type ActiveDirectoryGenerator struct {
	GCPService
}

// Generate TerraformResources from GCP API,
func (g *ActiveDirectoryGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := managedidentities.NewService(ctx)
	if err != nil {
		return err
	}
	project := g.GetArgs()["project"].(string)

	if err := svc.Projects.Locations.Global.Domains.List("projects/" + project + "/locations/global").Pages(ctx, func(page *managedidentities.ListDomainsResponse) error {
		for _, obj := range page.Domains {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name, name, "google_active_directory_domain", g.ProviderName,
				map[string]string{"domain_name": name, "project": project},
				activeDirectoryAllowEmptyValues, activeDirectoryAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}
