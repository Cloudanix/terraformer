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

	"google.golang.org/api/analyticshub/v1"
	"google.golang.org/api/compute/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var analyticsHubAllowEmptyValues = []string{""}

var analyticsHubAdditionalFields = map[string]interface{}{}

type AnalyticsHubGenerator struct {
	GCPService
}

// Generate TerraformResources from GCP API,
func (g *AnalyticsHubGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := analyticshub.NewService(ctx)
	if err != nil {
		return err
	}
	location := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + g.GetArgs()["project"].(string) + "/locations/" + location

	exchangeNames := []string{}
	list := svc.Projects.Locations.DataExchanges.List(parent)
	if err := list.Pages(ctx, func(page *analyticshub.ListDataExchangesResponse) error {
		for _, obj := range page.DataExchanges {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			exchangeNames = append(exchangeNames, obj.Name)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name, name, "google_bigquery_analytics_hub_data_exchange", g.ProviderName,
				map[string]string{"data_exchange_id": name, "project": g.GetArgs()["project"].(string), "location": location},
				analyticsHubAllowEmptyValues, analyticsHubAdditionalFields,
			))
			if policy, perr := svc.Projects.Locations.DataExchanges.GetIamPolicy(obj.Name, &analyticshub.GetIamPolicyRequest{}).Do(); perr == nil {
				for _, b := range policy.Bindings {
					for _, m := range b.Members {
						g.Resources = append(g.Resources, terraformutils.NewResource(
							obj.Name+" "+b.Role+" "+m, name+"_"+b.Role+"_"+m,
							"google_bigquery_analytics_hub_data_exchange_iam_member", g.ProviderName,
							map[string]string{"data_exchange_id": name, "role": b.Role, "member": m, "project": g.GetArgs()["project"].(string), "location": location},
							analyticsHubAllowEmptyValues, analyticsHubAdditionalFields))
					}
				}
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	for _, ex := range exchangeNames {
		exID := strings.Split(ex, "/")[len(strings.Split(ex, "/"))-1]
		if err := svc.Projects.Locations.DataExchanges.Listings.List(ex).Pages(ctx, func(page *analyticshub.ListListingsResponse) error {
			for _, obj := range page.Listings {
				t := strings.Split(obj.Name, "/")
				name := t[len(t)-1]
				g.Resources = append(g.Resources, terraformutils.NewResource(
					obj.Name, name, "google_bigquery_analytics_hub_listing", g.ProviderName,
					map[string]string{"listing_id": name, "data_exchange_id": exID, "project": g.GetArgs()["project"].(string), "location": location},
					analyticsHubAllowEmptyValues, analyticsHubAdditionalFields))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
	}
	return nil
}
