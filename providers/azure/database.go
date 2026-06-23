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
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mariadb/armmariadb"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysql"
	armmysqlflex "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysqlflexibleservers"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresql"
	armpgflex "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresqlflexibleservers"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
)

type DatabasesGenerator struct {
	AzureService
}

// InitResources imports the MariaDB / MySQL / PostgreSQL / SQL server families
// (servers + databases/configurations/firewall rules/vnet rules) plus the
// MySQL and PostgreSQL flexible-server tiers. Migrated to the Track 2 SDKs.
func (g *DatabasesGenerator) InitResources() error {
	if _, cred, _ := g.getClientOptions(); cred == nil {
		return nil
	}
	for _, fn := range []func() error{
		g.initMariaDB,
		g.initMySQL,
		g.initPostgreSQL,
		g.initSQL,
		g.initMySQLFlexibleServers,
		g.initPostgreSQLFlexibleServers,
	} {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}

func (g *DatabasesGenerator) initMariaDB() error {
	subscriptionID, cred, opts := g.getClientOptions()
	servers, err := armmariadb.NewServersClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	dbs, err := armmariadb.NewDatabasesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	cfgs, err := armmariadb.NewConfigurationsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	fws, err := armmariadb.NewFirewallRulesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	vnets, err := armmariadb.NewVirtualNetworkRulesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}

	var list []*armmariadb.Server
	for _, rg := range g.resourceGroups() {
		pager := servers.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return err
			}
			list = append(list, page.Value...)
		}
	}
	if len(g.resourceGroups()) == 0 {
		pager := servers.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return err
			}
			list = append(list, page.Value...)
		}
	}

	for _, s := range list {
		sid := valueOrEmpty(s.ID)
		if sid == "" {
			continue
		}
		g.AppendSimpleResource(sid, valueOrEmpty(s.Name), "azurerm_mariadb_server")
		parsed, err := ParseAzureResourceID(sid)
		if err != nil {
			return err
		}
		rg, name := parsed.ResourceGroup, valueOrEmpty(s.Name)
		if err := appendFromPager(&g.AzureService, dbs.NewListByServerPager(rg, name, nil),
			func(p armmariadb.DatabasesClientListByServerResponse) []*armmariadb.Database { return p.Value },
			func(i *armmariadb.Database) string { return valueOrEmpty(i.ID) },
			func(i *armmariadb.Database) string { return valueOrEmpty(i.Name) },
			"azurerm_mariadb_database"); err != nil {
			return err
		}
		if err := appendFromPager(&g.AzureService, cfgs.NewListByServerPager(rg, name, nil),
			func(p armmariadb.ConfigurationsClientListByServerResponse) []*armmariadb.Configuration {
				return p.Value
			},
			func(i *armmariadb.Configuration) string { return valueOrEmpty(i.ID) },
			func(i *armmariadb.Configuration) string { return valueOrEmpty(i.Name) },
			"azurerm_mariadb_configuration"); err != nil {
			return err
		}
		if err := appendFromPager(&g.AzureService, fws.NewListByServerPager(rg, name, nil),
			func(p armmariadb.FirewallRulesClientListByServerResponse) []*armmariadb.FirewallRule { return p.Value },
			func(i *armmariadb.FirewallRule) string { return valueOrEmpty(i.ID) },
			func(i *armmariadb.FirewallRule) string { return valueOrEmpty(i.Name) },
			"azurerm_mariadb_firewall_rule"); err != nil {
			return err
		}
		if err := appendFromPager(&g.AzureService, vnets.NewListByServerPager(rg, name, nil),
			func(p armmariadb.VirtualNetworkRulesClientListByServerResponse) []*armmariadb.VirtualNetworkRule {
				return p.Value
			},
			func(i *armmariadb.VirtualNetworkRule) string { return valueOrEmpty(i.ID) },
			func(i *armmariadb.VirtualNetworkRule) string { return valueOrEmpty(i.Name) },
			"azurerm_mariadb_virtual_network_rule"); err != nil {
			return err
		}
	}
	return nil
}

func (g *DatabasesGenerator) initMySQL() error {
	subscriptionID, cred, opts := g.getClientOptions()
	servers, err := armmysql.NewServersClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	dbs, err := armmysql.NewDatabasesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	cfgs, err := armmysql.NewConfigurationsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	fws, err := armmysql.NewFirewallRulesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	vnets, err := armmysql.NewVirtualNetworkRulesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}

	var list []*armmysql.Server
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		pager := servers.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return err
			}
			list = append(list, page.Value...)
		}
	}
	for _, rg := range rgs {
		pager := servers.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return err
			}
			list = append(list, page.Value...)
		}
	}

	for _, s := range list {
		sid := valueOrEmpty(s.ID)
		if sid == "" {
			continue
		}
		g.AppendSimpleResource(sid, valueOrEmpty(s.Name), "azurerm_mysql_server")
		parsed, err := ParseAzureResourceID(sid)
		if err != nil {
			return err
		}
		rg, name := parsed.ResourceGroup, valueOrEmpty(s.Name)
		if err := appendFromPager(&g.AzureService, dbs.NewListByServerPager(rg, name, nil),
			func(p armmysql.DatabasesClientListByServerResponse) []*armmysql.Database { return p.Value },
			func(i *armmysql.Database) string { return valueOrEmpty(i.ID) },
			func(i *armmysql.Database) string { return valueOrEmpty(i.Name) },
			"azurerm_mysql_database"); err != nil {
			return err
		}
		if err := appendFromPager(&g.AzureService, cfgs.NewListByServerPager(rg, name, nil),
			func(p armmysql.ConfigurationsClientListByServerResponse) []*armmysql.Configuration { return p.Value },
			func(i *armmysql.Configuration) string { return valueOrEmpty(i.ID) },
			func(i *armmysql.Configuration) string { return valueOrEmpty(i.Name) },
			"azurerm_mysql_configuration"); err != nil {
			return err
		}
		if err := appendFromPager(&g.AzureService, fws.NewListByServerPager(rg, name, nil),
			func(p armmysql.FirewallRulesClientListByServerResponse) []*armmysql.FirewallRule { return p.Value },
			func(i *armmysql.FirewallRule) string { return valueOrEmpty(i.ID) },
			func(i *armmysql.FirewallRule) string { return valueOrEmpty(i.Name) },
			"azurerm_mysql_firewall_rule"); err != nil {
			return err
		}
		if err := appendFromPager(&g.AzureService, vnets.NewListByServerPager(rg, name, nil),
			func(p armmysql.VirtualNetworkRulesClientListByServerResponse) []*armmysql.VirtualNetworkRule {
				return p.Value
			},
			func(i *armmysql.VirtualNetworkRule) string { return valueOrEmpty(i.ID) },
			func(i *armmysql.VirtualNetworkRule) string { return valueOrEmpty(i.Name) },
			"azurerm_mysql_virtual_network_rule"); err != nil {
			return err
		}
	}
	return nil
}

func (g *DatabasesGenerator) initPostgreSQL() error {
	subscriptionID, cred, opts := g.getClientOptions()
	servers, err := armpostgresql.NewServersClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	dbs, err := armpostgresql.NewDatabasesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	cfgs, err := armpostgresql.NewConfigurationsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	fws, err := armpostgresql.NewFirewallRulesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	vnets, err := armpostgresql.NewVirtualNetworkRulesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}

	var list []*armpostgresql.Server
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		pager := servers.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return err
			}
			list = append(list, page.Value...)
		}
	}
	for _, rg := range rgs {
		pager := servers.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return err
			}
			list = append(list, page.Value...)
		}
	}

	for _, s := range list {
		sid := valueOrEmpty(s.ID)
		if sid == "" {
			continue
		}
		g.AppendSimpleResource(sid, valueOrEmpty(s.Name), "azurerm_postgresql_server")
		parsed, err := ParseAzureResourceID(sid)
		if err != nil {
			return err
		}
		rg, name := parsed.ResourceGroup, valueOrEmpty(s.Name)
		if err := appendFromPager(&g.AzureService, dbs.NewListByServerPager(rg, name, nil),
			func(p armpostgresql.DatabasesClientListByServerResponse) []*armpostgresql.Database { return p.Value },
			func(i *armpostgresql.Database) string { return valueOrEmpty(i.ID) },
			func(i *armpostgresql.Database) string { return valueOrEmpty(i.Name) },
			"azurerm_postgresql_database"); err != nil {
			return err
		}
		if err := appendFromPager(&g.AzureService, cfgs.NewListByServerPager(rg, name, nil),
			func(p armpostgresql.ConfigurationsClientListByServerResponse) []*armpostgresql.Configuration {
				return p.Value
			},
			func(i *armpostgresql.Configuration) string { return valueOrEmpty(i.ID) },
			func(i *armpostgresql.Configuration) string { return valueOrEmpty(i.Name) },
			"azurerm_postgresql_configuration"); err != nil {
			return err
		}
		if err := appendFromPager(&g.AzureService, fws.NewListByServerPager(rg, name, nil),
			func(p armpostgresql.FirewallRulesClientListByServerResponse) []*armpostgresql.FirewallRule {
				return p.Value
			},
			func(i *armpostgresql.FirewallRule) string { return valueOrEmpty(i.ID) },
			func(i *armpostgresql.FirewallRule) string { return valueOrEmpty(i.Name) },
			"azurerm_postgresql_firewall_rule"); err != nil {
			return err
		}
		if err := appendFromPager(&g.AzureService, vnets.NewListByServerPager(rg, name, nil),
			func(p armpostgresql.VirtualNetworkRulesClientListByServerResponse) []*armpostgresql.VirtualNetworkRule {
				return p.Value
			},
			func(i *armpostgresql.VirtualNetworkRule) string { return valueOrEmpty(i.ID) },
			func(i *armpostgresql.VirtualNetworkRule) string { return valueOrEmpty(i.Name) },
			"azurerm_postgresql_virtual_network_rule"); err != nil {
			return err
		}
	}
	return nil
}

func (g *DatabasesGenerator) initSQL() error {
	subscriptionID, cred, opts := g.getClientOptions()
	servers, err := armsql.NewServersClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	dbs, err := armsql.NewDatabasesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	fws, err := armsql.NewFirewallRulesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	admins, err := armsql.NewServerAzureADAdministratorsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	pools, err := armsql.NewElasticPoolsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	failovers, err := armsql.NewFailoverGroupsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	vnets, err := armsql.NewVirtualNetworkRulesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}

	var list []*armsql.Server
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		pager := servers.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return err
			}
			list = append(list, page.Value...)
		}
	}
	for _, rg := range rgs {
		pager := servers.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return err
			}
			list = append(list, page.Value...)
		}
	}

	for _, s := range list {
		sid := valueOrEmpty(s.ID)
		if sid == "" {
			continue
		}
		g.AppendSimpleResource(sid, valueOrEmpty(s.Name), "azurerm_mssql_server")
		parsed, err := ParseAzureResourceID(sid)
		if err != nil {
			return err
		}
		rg, name := parsed.ResourceGroup, valueOrEmpty(s.Name)
		if err := appendFromPager(&g.AzureService, dbs.NewListByServerPager(rg, name, nil),
			func(p armsql.DatabasesClientListByServerResponse) []*armsql.Database { return p.Value },
			func(i *armsql.Database) string { return valueOrEmpty(i.ID) },
			func(i *armsql.Database) string { return valueOrEmpty(i.Name) },
			"azurerm_mssql_database"); err != nil {
			return err
		}
		if err := appendFromPager(&g.AzureService, fws.NewListByServerPager(rg, name, nil),
			func(p armsql.FirewallRulesClientListByServerResponse) []*armsql.FirewallRule { return p.Value },
			func(i *armsql.FirewallRule) string { return valueOrEmpty(i.ID) },
			func(i *armsql.FirewallRule) string { return valueOrEmpty(i.Name) },
			"azurerm_mssql_firewall_rule"); err != nil {
			return err
		}
		if err := appendFromPager(&g.AzureService, admins.NewListByServerPager(rg, name, nil),
			func(p armsql.ServerAzureADAdministratorsClientListByServerResponse) []*armsql.ServerAzureADAdministrator {
				return p.Value
			},
			func(i *armsql.ServerAzureADAdministrator) string { return valueOrEmpty(i.ID) },
			func(i *armsql.ServerAzureADAdministrator) string { return valueOrEmpty(i.Name) },
			"azurerm_sql_active_directory_administrator"); err != nil {
			return err
		}
		if err := appendFromPager(&g.AzureService, pools.NewListByServerPager(rg, name, nil),
			func(p armsql.ElasticPoolsClientListByServerResponse) []*armsql.ElasticPool { return p.Value },
			func(i *armsql.ElasticPool) string { return valueOrEmpty(i.ID) },
			func(i *armsql.ElasticPool) string { return valueOrEmpty(i.Name) },
			"azurerm_sql_elasticpool"); err != nil {
			return err
		}
		if err := appendFromPager(&g.AzureService, failovers.NewListByServerPager(rg, name, nil),
			func(p armsql.FailoverGroupsClientListByServerResponse) []*armsql.FailoverGroup { return p.Value },
			func(i *armsql.FailoverGroup) string { return valueOrEmpty(i.ID) },
			func(i *armsql.FailoverGroup) string { return valueOrEmpty(i.Name) },
			"azurerm_sql_failover_group"); err != nil {
			return err
		}
		if err := appendFromPager(&g.AzureService, vnets.NewListByServerPager(rg, name, nil),
			func(p armsql.VirtualNetworkRulesClientListByServerResponse) []*armsql.VirtualNetworkRule {
				return p.Value
			},
			func(i *armsql.VirtualNetworkRule) string { return valueOrEmpty(i.ID) },
			func(i *armsql.VirtualNetworkRule) string { return valueOrEmpty(i.Name) },
			"azurerm_sql_virtual_network_rule"); err != nil {
			return err
		}
	}
	return nil
}

// initMySQLFlexibleServers enumerates azurerm_mysql_flexible_server (newer
// flexible tier). Subscription-wide unless -R is set.
func (g *DatabasesGenerator) initMySQLFlexibleServers() error {
	subscriptionID, cred, opts := g.getClientOptions()
	client, err := armmysqlflex.NewServersClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	id := func(i *armmysqlflex.Server) string { return valueOrEmpty(i.ID) }
	name := func(i *armmysqlflex.Server) string { return valueOrEmpty(i.Name) }
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		return appendFromPager(&g.AzureService, client.NewListPager(nil),
			func(p armmysqlflex.ServersClientListResponse) []*armmysqlflex.Server { return p.Value },
			id, name, "azurerm_mysql_flexible_server")
	}
	for _, rg := range rgs {
		if err := appendFromPager(&g.AzureService, client.NewListByResourceGroupPager(rg, nil),
			func(p armmysqlflex.ServersClientListByResourceGroupResponse) []*armmysqlflex.Server { return p.Value },
			id, name, "azurerm_mysql_flexible_server"); err != nil {
			return err
		}
	}
	return nil
}

// initPostgreSQLFlexibleServers enumerates azurerm_postgresql_flexible_server.
func (g *DatabasesGenerator) initPostgreSQLFlexibleServers() error {
	subscriptionID, cred, opts := g.getClientOptions()
	client, err := armpgflex.NewServersClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	id := func(i *armpgflex.Server) string { return valueOrEmpty(i.ID) }
	name := func(i *armpgflex.Server) string { return valueOrEmpty(i.Name) }
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		return appendFromPager(&g.AzureService, client.NewListPager(nil),
			func(p armpgflex.ServersClientListResponse) []*armpgflex.Server { return p.Value },
			id, name, "azurerm_postgresql_flexible_server")
	}
	for _, rg := range rgs {
		if err := appendFromPager(&g.AzureService, client.NewListByResourceGroupPager(rg, nil),
			func(p armpgflex.ServersClientListByResourceGroupResponse) []*armpgflex.Server { return p.Value },
			id, name, "azurerm_postgresql_flexible_server"); err != nil {
			return err
		}
	}
	return nil
}

func (g *DatabasesGenerator) PostConvertHook() error {
	dbEngines := []string{
		"mariadb",
		"mysql",
		"postgresql",
		"sql",
	}

	for _, engineName := range dbEngines {
		for _, resource := range g.Resources {
			dbServerResourceType := fmt.Sprintf("azurerm_%s_server", engineName)
			if resource.InstanceInfo.Type == dbServerResourceType {
				dbName := resource.Item["name"]
				for rIdx, r := range g.Resources {
					if r.InstanceInfo.Type != dbServerResourceType &&
						strings.Contains(r.InstanceInfo.Type, engineName) &&
						r.Item["server_name"] == dbName {
						g.Resources[rIdx].Item["server_name"] = fmt.Sprintf("${%s.%s}", resource.InstanceInfo.Id, "name")
					}
				}
			}
		}
	}

	return nil
}
