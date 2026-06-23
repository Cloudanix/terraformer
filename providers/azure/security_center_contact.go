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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"
)

type SecurityCenterContactGenerator struct {
	AzureService
}

// InitResources imports azurerm_security_center_contact. Subscription-scoped
// (skipped when -R is set). Migrated to the Track 2 armsecurity SDK.
func (g *SecurityCenterContactGenerator) InitResources() error {
	if len(g.resourceGroups()) > 0 {
		return nil
	}
	subscriptionID, cred, opts := g.getClientOptions()
	if cred == nil {
		return nil
	}
	client, err := armsecurity.NewContactsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	return appendFromPager(&g.AzureService, client.NewListPager(nil),
		func(p armsecurity.ContactsClientListResponse) []*armsecurity.Contact { return p.Value },
		func(i *armsecurity.Contact) string { return valueOrEmpty(i.ID) },
		func(i *armsecurity.Contact) string { return valueOrEmpty(i.Name) },
		"azurerm_security_center_contact")
}
