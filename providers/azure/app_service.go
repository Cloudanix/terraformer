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
	"strings"

	armappservice "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v3"
)

type AppServiceGenerator struct {
	AzureService
}

// siteResourceType maps an App Service site's `kind` to the modern azurerm
// resource type. The legacy azurerm_app_service was removed in azurerm v4 (the
// supported floor), so sites are emitted as the Linux/Windows web/function-app
// resources instead.
func siteResourceType(kind string) string {
	k := strings.ToLower(kind)
	isLinux := strings.Contains(k, "linux")
	switch {
	case strings.Contains(k, "functionapp") && isLinux:
		return "azurerm_linux_function_app"
	case strings.Contains(k, "functionapp"):
		return "azurerm_windows_function_app"
	case isLinux:
		return "azurerm_linux_web_app"
	default:
		return "azurerm_windows_web_app"
	}
}

func (g *AppServiceGenerator) appendSites(values []*armappservice.Site) {
	for _, site := range values {
		if site == nil {
			continue
		}
		id := valueOrEmpty(site.ID)
		if id == "" {
			continue
		}
		g.AppendSimpleResource(id, valueOrEmpty(site.Name), siteResourceType(valueOrEmpty(site.Kind)))
	}
}

func (g *AppServiceGenerator) initSites() error {
	subscriptionID, cred, opts := g.getClientOptions()
	if cred == nil {
		return nil
	}
	client, err := armappservice.NewWebAppsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return err
			}
			g.appendSites(page.Value)
		}
		return nil
	}
	for _, rg := range rgs {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return err
			}
			g.appendSites(page.Value)
		}
	}
	return nil
}

// initServicePlans enumerates azurerm_service_plan via the Track 2 armappservice
// SDK (the modern replacement for the legacy azurerm_app_service_plan).
func (g *AppServiceGenerator) initServicePlans() error {
	subscriptionID, cred, opts := g.getClientOptions()
	if cred == nil {
		return nil
	}
	client, err := armappservice.NewPlansClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	return appendFromPager(&g.AzureService, client.NewListPager(nil),
		func(p armappservice.PlansClientListResponse) []*armappservice.Plan { return p.Value },
		func(i *armappservice.Plan) string { return valueOrEmpty(i.ID) },
		func(i *armappservice.Plan) string { return valueOrEmpty(i.Name) },
		"azurerm_service_plan")
}

func (g *AppServiceGenerator) InitResources() error {
	if err := g.initServicePlans(); err != nil {
		return err
	}
	return g.initSites()
}
