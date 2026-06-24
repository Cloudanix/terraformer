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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/attestation/armattestation"
)

type AttestationGenerator struct {
	AzureService
}

// InitResources imports azurerm_attestation_provider. The attestation Providers
// API returns the full list in one response (no pager).
func (g *AttestationGenerator) InitResources() error {
	subscriptionID, cred, opts := g.getClientOptions()
	if cred == nil {
		return nil
	}
	client, err := armattestation.NewProvidersClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}

	var providers []*armattestation.Provider
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		resp, err := client.List(context.TODO(), nil)
		if err != nil {
			return err
		}
		providers = resp.Value
	} else {
		for _, rg := range rgs {
			resp, err := client.ListByResourceGroup(context.TODO(), rg, nil)
			if err != nil {
				return err
			}
			providers = append(providers, resp.Value...)
		}
	}

	for _, p := range providers {
		if p == nil {
			continue
		}
		if id := valueOrEmpty(p.ID); id != "" {
			g.AppendSimpleResource(id, valueOrEmpty(p.Name), "azurerm_attestation_provider")
		}
	}
	return nil
}
