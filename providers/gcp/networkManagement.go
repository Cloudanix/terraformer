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

	"google.golang.org/api/networkmanagement/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var networkManagementAllowEmptyValues = []string{""}

var networkManagementAdditionalFields = map[string]interface{}{}

type NetworkManagementGenerator struct {
	GCPService
}

// Generate TerraformResources from GCP API,
// Connectivity tests are global (response field is Resources).
func (g *NetworkManagementGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := networkmanagement.NewService(ctx)
	if err != nil {
		return err
	}
	project := g.GetArgs()["project"].(string)

	if err := svc.Projects.Locations.Global.ConnectivityTests.List("projects/"+project+"/locations/global").Pages(ctx, func(page *networkmanagement.ListConnectivityTestsResponse) error {
		for _, obj := range page.Resources {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name, name, "google_network_management_connectivity_test", g.ProviderName,
				map[string]string{"name": name, "project": project},
				networkManagementAllowEmptyValues, networkManagementAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}
