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

	"google.golang.org/api/apikeys/v2"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var apiKeysAllowEmptyValues = []string{""}

var apiKeysAdditionalFields = map[string]interface{}{}

type APIKeysGenerator struct {
	GCPService
}

// Generate TerraformResources from GCP API,
// API keys live under the global location.
func (g *APIKeysGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := apikeys.NewService(ctx)
	if err != nil {
		return err
	}
	project := g.GetArgs()["project"].(string)

	if err := svc.Projects.Locations.Keys.List("projects/"+project+"/locations/global").Pages(ctx, func(page *apikeys.V2ListKeysResponse) error {
		for _, obj := range page.Keys {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name, name, "google_apikeys_key", g.ProviderName,
				map[string]string{"name": name, "project": project},
				apiKeysAllowEmptyValues, apiKeysAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}
