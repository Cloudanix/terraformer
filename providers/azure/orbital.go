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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/orbital/armorbital"
)

type OrbitalGenerator struct {
	AzureService
}

func (g *OrbitalGenerator) InitResources() error {
	if _, cred, _ := g.getClientOptions(); cred == nil {
		return nil
	}
	if err := g.initSpacecrafts(); err != nil {
		return err
	}
	return g.initContactProfiles()
}

func (g *OrbitalGenerator) initSpacecrafts() error {
	subscriptionID, cred, opts := g.getClientOptions()
	client, err := armorbital.NewSpacecraftsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	id := func(i *armorbital.Spacecraft) string { return valueOrEmpty(i.ID) }
	name := func(i *armorbital.Spacecraft) string { return valueOrEmpty(i.Name) }
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		return appendFromPager(&g.AzureService, client.NewListBySubscriptionPager(nil),
			func(p armorbital.SpacecraftsClientListBySubscriptionResponse) []*armorbital.Spacecraft {
				return p.Value
			},
			id, name, "azurerm_orbital_spacecraft")
	}
	for _, rg := range rgs {
		if err := appendFromPager(&g.AzureService, client.NewListPager(rg, nil),
			func(p armorbital.SpacecraftsClientListResponse) []*armorbital.Spacecraft { return p.Value },
			id, name, "azurerm_orbital_spacecraft"); err != nil {
			return err
		}
	}
	return nil
}

func (g *OrbitalGenerator) initContactProfiles() error {
	subscriptionID, cred, opts := g.getClientOptions()
	client, err := armorbital.NewContactProfilesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	id := func(i *armorbital.ContactProfile) string { return valueOrEmpty(i.ID) }
	name := func(i *armorbital.ContactProfile) string { return valueOrEmpty(i.Name) }
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		return appendFromPager(&g.AzureService, client.NewListBySubscriptionPager(nil),
			func(p armorbital.ContactProfilesClientListBySubscriptionResponse) []*armorbital.ContactProfile {
				return p.Value
			},
			id, name, "azurerm_orbital_contact_profile")
	}
	for _, rg := range rgs {
		if err := appendFromPager(&g.AzureService, client.NewListPager(rg, nil),
			func(p armorbital.ContactProfilesClientListResponse) []*armorbital.ContactProfile { return p.Value },
			id, name, "azurerm_orbital_contact_profile"); err != nil {
			return err
		}
	}
	return nil
}
