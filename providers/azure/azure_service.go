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
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/hashicorp/go-azure-helpers/authentication"
)

// defaultAllowEmptyValues lists attribute prefixes that may be empty without
// being dropped. Azure tags map the same way across every service, so this
// package var replaces the per-file empty slices used by Track 1 generators.
var defaultAllowEmptyValues = []string{"tags."}

type AzureService struct { //nolint
	terraformutils.Service
}

// getClientOptions returns the subscription ID, the azcore credential, and
// *arm.ClientOptions (sovereign cloud + retry preconfigured) for constructing
// armXxx clients. All Azure generators are now Track 2.
func (az *AzureService) getClientOptions() (subscriptionID string, cred azcore.TokenCredential, opts *arm.ClientOptions) {
	cfg := az.Args["config"].(authentication.Config)
	subscriptionID = cfg.SubscriptionID
	cred, _ = az.Args["token_credential"].(azcore.TokenCredential)
	opts = &arm.ClientOptions{
		ClientOptions: policy.ClientOptions{
			Cloud: cloudConfig(cfg.Environment),
			Retry: policy.RetryOptions{MaxRetries: 5},
		},
	}
	return subscriptionID, cred, opts
}

// cloudConfig maps the autorest environment name to an azcore cloud config so
// Gov/China users keep working after the Track 2 swap (replaces the Track 1
// CustomResourceManagerEndpoint behavior).
func cloudConfig(environment string) cloud.Configuration {
	switch strings.ToLower(strings.ReplaceAll(environment, " ", "")) {
	case "usgovernment", "usgovernmentcloud", "azureusgovernment", "azureusgovernmentcloud":
		return cloud.AzureGovernment
	case "china", "chinacloud", "azurechina", "azurechinacloud":
		return cloud.AzureChina
	default:
		return cloud.AzurePublic
	}
}

// resourceGroups returns the requested resource groups. The -R flag accepts a
// colon-separated list (resource_group=name1:name2:name3); an empty value means
// subscription scope (caller should use the subscription-wide List instead of
// ListByResourceGroup).
func (az *AzureService) resourceGroups() []string {
	raw, _ := az.Args["resource_group"].(string)
	if raw == "" {
		return nil
	}
	var out []string
	for _, rg := range strings.Split(raw, ":") {
		if rg = strings.TrimSpace(rg); rg != "" {
			out = append(out, rg)
		}
	}
	return out
}

func (az *AzureService) AppendSimpleResource(id string, resourceName string, resourceType string) {
	newResource := terraformutils.NewSimpleResource(id, resourceName, resourceType, az.ProviderName, []string{})
	az.Resources = append(az.Resources, newResource)
}

func (az *AzureService) AppendSimpleResourceWithDuplicateCheck(id string, resourceName string, resourceType string) {
	tferexist, _ := az.DuplicateCheck(id, resourceName, resourceType)
	if !tferexist {
		resourceName = resourceName + "_" + id
	}
	newResource := terraformutils.NewSimpleResource(id, resourceName, resourceType, az.ProviderName, []string{})
	az.Resources = append(az.Resources, newResource)
}

// This method checks if same resource name(tfer) exists with
// same id
func (az *AzureService) DuplicateCheck(id string, resourceName string, resourceType string) (bool, bool) {
	var tferexist, idexist bool
	tferName := terraformutils.TfSanitize(resourceName)
	for _, resource := range az.Resources {
		if tferName == resource.ResourceName {
			if id == resource.InstanceState.ID {
				tferexist = true
				idexist = true
			} else {
				tferexist = true
				idexist = false
			}
		}
	}
	return tferexist, idexist
}

func (az *AzureService) appendSimpleAssociation(id string, linkedResourceName string, resourceName *string, resourceType string, attributes map[string]string) {
	var resourceName2 string
	if resourceName != nil {
		resourceName2 = *resourceName
	} else {
		resourceName0 := strings.ReplaceAll(resourceType, "azurerm_", "")
		resourceName1 := resourceName0[strings.IndexByte(resourceName0, '_'):]
		resourceName2 = linkedResourceName + resourceName1
	}
	newResource := terraformutils.NewResource(
		id, resourceName2, resourceType, az.ProviderName, attributes,
		[]string{"name"},
		map[string]interface{}{},
	)
	az.Resources = append(az.Resources, newResource)
}
