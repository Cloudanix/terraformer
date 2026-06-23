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

	"google.golang.org/api/iap/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var iapAllowEmptyValues = []string{""}

var iapAdditionalFields = map[string]interface{}{}

type IapGenerator struct {
	GCPService
}

// Generate TerraformResources from GCP API,
func (g *IapGenerator) InitResources() error {
	ctx := context.Background()
	iapService, err := iap.NewService(ctx)
	if err != nil {
		return err
	}
	project := g.GetArgs()["project"].(string)

	brandsList, err := iapService.Projects.Brands.List("projects/" + project).Do()
	if err != nil {
		log.Println(err)
		return nil
	}
	for _, obj := range brandsList.Brands {
		t := strings.Split(obj.Name, "/")
		name := t[len(t)-1]
		g.Resources = append(g.Resources, terraformutils.NewResource(
			obj.Name,
			name,
			"google_iap_brand",
			g.ProviderName,
			map[string]string{
				"project": project,
			},
			iapAllowEmptyValues,
			iapAdditionalFields,
		))
	}
	return nil
}
