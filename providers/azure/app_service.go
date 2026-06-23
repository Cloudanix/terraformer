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
	"log"

	"github.com/Azure/go-autorest/autorest"
	"github.com/hashicorp/go-azure-helpers/authentication"

	armappservice "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v3"
	"github.com/Azure/azure-sdk-for-go/services/web/mgmt/2019-08-01/web"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type AppServiceGenerator struct {
	AzureService
}

func (g AppServiceGenerator) listApps() ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()

	subscriptionID := g.Args["config"].(authentication.Config).SubscriptionID
	resourceManagerEndpoint := g.Args["config"].(authentication.Config).CustomResourceManagerEndpoint
	appServiceClient := web.NewAppsClientWithBaseURI(resourceManagerEndpoint, subscriptionID)
	appServiceClient.Authorizer = g.Args["authorizer"].(autorest.Authorizer)
	var (
		appsIterator web.AppCollectionIterator
		err          error
	)
	if rg := g.Args["resource_group"].(string); rg != "" {
		appsIterator, err = appServiceClient.ListByResourceGroupComplete(ctx, rg, nil)
	} else {
		appsIterator, err = appServiceClient.ListComplete(ctx)
	}
	if err != nil {
		return nil, err
	}
	for appsIterator.NotDone() {
		site := appsIterator.Value()
		resources = append(resources, terraformutils.NewSimpleResource(
			*site.ID,
			*site.Name,
			"azurerm_app_service",
			g.ProviderName,
			[]string{}))

		if err := appsIterator.NextWithContext(ctx); err != nil {
			log.Println(err)
			return resources, err
		}
	}

	return resources, nil
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
	resources, err := g.listApps()
	if err != nil {
		return err
	}

	g.Resources = append(g.Resources, resources...)

	if err := g.initServicePlans(); err != nil {
		return err
	}

	return nil
}
