// Copyright 2019 The Terraformer Authors.
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

package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cdn/armcdn"
)

type CDNGenerator struct {
	AzureService
}

func (g *CDNGenerator) InitResources() error {
	subscriptionID, cred, opts := g.getClientOptions()
	if cred == nil {
		return nil
	}
	profilesClient, err := armcdn.NewProfilesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	endpointsClient, err := armcdn.NewEndpointsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}

	profiles, err := g.listProfiles(profilesClient)
	if err != nil {
		return err
	}
	for _, profile := range profiles {
		profileID := valueOrEmpty(profile.ID)
		if profileID == "" {
			continue
		}
		g.AppendSimpleResource(profileID, valueOrEmpty(profile.Name), "azurerm_cdn_profile")

		parsed, err := ParseAzureResourceID(profileID)
		if err != nil {
			return err
		}
		if err := appendFromPager(&g.AzureService,
			endpointsClient.NewListByProfilePager(parsed.ResourceGroup, valueOrEmpty(profile.Name), nil),
			func(p armcdn.EndpointsClientListByProfileResponse) []*armcdn.Endpoint { return p.Value },
			func(i *armcdn.Endpoint) string { return valueOrEmpty(i.ID) },
			func(i *armcdn.Endpoint) string { return valueOrEmpty(i.Name) },
			"azurerm_cdn_endpoint"); err != nil {
			return err
		}
	}
	return nil
}

func (g *CDNGenerator) listProfiles(client *armcdn.ProfilesClient) ([]*armcdn.Profile, error) {
	var profiles []*armcdn.Profile
	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}
		for _, p := range page.Value {
			if p != nil {
				profiles = append(profiles, p)
			}
		}
	}
	return profiles, nil
}
