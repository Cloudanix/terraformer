// Copyright 2021 The Terraformer Authors.
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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v7"
)

type RouteTableGenerator struct {
	AzureService
}

// InitResources imports azurerm_route_table (+ nested azurerm_route) and
// azurerm_route_filter. Migrated to the Track 2 armnetwork SDK; preserves the
// duplicate-name handling of the Track 1 implementation.
func (g *RouteTableGenerator) InitResources() error {
	subscriptionID, cred, opts := g.getClientOptions()
	if cred == nil {
		return nil
	}
	tablesClient, err := armnetwork.NewRouteTablesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	routesClient, err := armnetwork.NewRoutesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	filtersClient, err := armnetwork.NewRouteFiltersClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}

	tables, err := g.listRouteTables(tablesClient)
	if err != nil {
		return err
	}
	for _, table := range tables {
		tableID := valueOrEmpty(table.ID)
		if tableID == "" {
			continue
		}
		g.AppendSimpleResourceWithDuplicateCheck(tableID, valueOrEmpty(table.Name), "azurerm_route_table")
		parsed, err := ParseAzureResourceID(tableID)
		if err != nil {
			return err
		}
		if err := g.appendRoutes(routesClient, parsed.ResourceGroup, valueOrEmpty(table.Name)); err != nil {
			return err
		}
	}

	return g.appendRouteFilters(filtersClient)
}

func (g *RouteTableGenerator) listRouteTables(client *armnetwork.RouteTablesClient) ([]*armnetwork.RouteTable, error) {
	var tables []*armnetwork.RouteTable
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		pager := client.NewListAllPager(nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return nil, err
			}
			tables = append(tables, page.Value...)
		}
		return tables, nil
	}
	for _, rg := range rgs {
		pager := client.NewListPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return nil, err
			}
			tables = append(tables, page.Value...)
		}
	}
	return tables, nil
}

func (g *RouteTableGenerator) appendRoutes(client *armnetwork.RoutesClient, resourceGroup, routeTableName string) error {
	pager := client.NewListPager(resourceGroup, routeTableName, nil)
	for pager.More() {
		page, err := pager.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, route := range page.Value {
			if route == nil {
				continue
			}
			if id := valueOrEmpty(route.ID); id != "" {
				g.AppendSimpleResourceWithDuplicateCheck(id, valueOrEmpty(route.Name), "azurerm_route")
			}
		}
	}
	return nil
}

func (g *RouteTableGenerator) appendRouteFilters(client *armnetwork.RouteFiltersClient) error {
	id := func(i *armnetwork.RouteFilter) string { return valueOrEmpty(i.ID) }
	name := func(i *armnetwork.RouteFilter) string { return valueOrEmpty(i.Name) }
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		return appendFromPager(&g.AzureService, client.NewListPager(nil),
			func(p armnetwork.RouteFiltersClientListResponse) []*armnetwork.RouteFilter { return p.Value },
			id, name, "azurerm_route_filter")
	}
	for _, rg := range rgs {
		if err := appendFromPager(&g.AzureService, client.NewListByResourceGroupPager(rg, nil),
			func(p armnetwork.RouteFiltersClientListByResourceGroupResponse) []*armnetwork.RouteFilter {
				return p.Value
			},
			id, name, "azurerm_route_filter"); err != nil {
			return err
		}
	}
	return nil
}
