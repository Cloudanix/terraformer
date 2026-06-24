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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cosmos/armcosmos/v3"
)

type CosmosDBGenerator struct {
	AzureService
}

// sqlOldFormatID rewrites a Cosmos DB SQL resource ID from the current
// "sqlDatabases" segment to the legacy "databases" segment the azurerm provider
// still expects (see terraform-provider-azurerm#7472).
func sqlOldFormatID(id string) string {
	return strings.Replace(id, "sqlDatabases", "databases", 1)
}

// InitResources imports azurerm_cosmosdb_account and its SQL, Table, MongoDB,
// Cassandra and Gremlin child resources. Migrated to the Track 2 armcosmos SDK.
func (g *CosmosDBGenerator) InitResources() error {
	subscriptionID, cred, opts := g.getClientOptions()
	if cred == nil {
		return nil
	}
	accountsClient, err := armcosmos.NewDatabaseAccountsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	sqlClient, err := armcosmos.NewSQLResourcesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	tableClient, err := armcosmos.NewTableResourcesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	mongoClient, err := armcosmos.NewMongoDBResourcesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	cassandraClient, err := armcosmos.NewCassandraResourcesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	gremlinClient, err := armcosmos.NewGremlinResourcesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}

	accounts, err := g.listAccounts(accountsClient)
	if err != nil {
		return err
	}
	for _, account := range accounts {
		accountID := valueOrEmpty(account.ID)
		if accountID == "" {
			continue
		}
		g.AppendSimpleResource(accountID, valueOrEmpty(account.Name), "azurerm_cosmosdb_account")
		parsed, err := ParseAzureResourceID(accountID)
		if err != nil {
			return err
		}
		rg, acc := parsed.ResourceGroup, valueOrEmpty(account.Name)

		if err := g.appendSQL(sqlClient, rg, acc); err != nil {
			return err
		}
		if err := g.appendTables(tableClient, rg, acc); err != nil {
			return err
		}
		if err := g.appendMongo(mongoClient, rg, acc); err != nil {
			return err
		}
		if err := g.appendCassandra(cassandraClient, rg, acc); err != nil {
			return err
		}
		if err := g.appendGremlin(gremlinClient, rg, acc); err != nil {
			return err
		}
	}
	return nil
}

func (g *CosmosDBGenerator) listAccounts(client *armcosmos.DatabaseAccountsClient) ([]*armcosmos.DatabaseAccountGetResults, error) {
	var accounts []*armcosmos.DatabaseAccountGetResults
	rgs := g.resourceGroups()
	if len(rgs) == 0 {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return nil, err
			}
			accounts = append(accounts, page.Value...)
		}
		return accounts, nil
	}
	for _, rg := range rgs {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(context.TODO())
			if err != nil {
				return nil, err
			}
			accounts = append(accounts, page.Value...)
		}
	}
	return accounts, nil
}

func (g *CosmosDBGenerator) appendSQL(client *armcosmos.SQLResourcesClient, rg, acc string) error {
	dbPager := client.NewListSQLDatabasesPager(rg, acc, nil)
	for dbPager.More() {
		page, err := dbPager.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, db := range page.Value {
			if db == nil || valueOrEmpty(db.ID) == "" {
				continue
			}
			dbName := valueOrEmpty(db.Name)
			g.AppendSimpleResource(sqlOldFormatID(valueOrEmpty(db.ID)), dbName, "azurerm_cosmosdb_sql_database")

			cPager := client.NewListSQLContainersPager(rg, acc, dbName, nil)
			for cPager.More() {
				cPage, err := cPager.NextPage(context.TODO())
				if err != nil {
					return err
				}
				for _, c := range cPage.Value {
					if c == nil || valueOrEmpty(c.ID) == "" {
						continue
					}
					g.AppendSimpleResource(sqlOldFormatID(valueOrEmpty(c.ID)), valueOrEmpty(c.Name), "azurerm_cosmosdb_sql_container")
				}
			}
		}
	}
	return nil
}

func (g *CosmosDBGenerator) appendTables(client *armcosmos.TableResourcesClient, rg, acc string) error {
	return appendFromPager(&g.AzureService, client.NewListTablesPager(rg, acc, nil),
		func(p armcosmos.TableResourcesClientListTablesResponse) []*armcosmos.TableGetResults { return p.Value },
		func(i *armcosmos.TableGetResults) string { return valueOrEmpty(i.ID) },
		func(i *armcosmos.TableGetResults) string { return valueOrEmpty(i.Name) },
		"azurerm_cosmosdb_table")
}

func (g *CosmosDBGenerator) appendMongo(client *armcosmos.MongoDBResourcesClient, rg, acc string) error {
	dbPager := client.NewListMongoDBDatabasesPager(rg, acc, nil)
	for dbPager.More() {
		page, err := dbPager.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, db := range page.Value {
			if db == nil || valueOrEmpty(db.ID) == "" {
				continue
			}
			dbName := valueOrEmpty(db.Name)
			g.AppendSimpleResource(valueOrEmpty(db.ID), dbName, "azurerm_cosmosdb_mongo_database")
			if err := appendFromPager(&g.AzureService, client.NewListMongoDBCollectionsPager(rg, acc, dbName, nil),
				func(p armcosmos.MongoDBResourcesClientListMongoDBCollectionsResponse) []*armcosmos.MongoDBCollectionGetResults {
					return p.Value
				},
				func(i *armcosmos.MongoDBCollectionGetResults) string { return valueOrEmpty(i.ID) },
				func(i *armcosmos.MongoDBCollectionGetResults) string { return valueOrEmpty(i.Name) },
				"azurerm_cosmosdb_mongo_collection"); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *CosmosDBGenerator) appendCassandra(client *armcosmos.CassandraResourcesClient, rg, acc string) error {
	ksPager := client.NewListCassandraKeyspacesPager(rg, acc, nil)
	for ksPager.More() {
		page, err := ksPager.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, ks := range page.Value {
			if ks == nil || valueOrEmpty(ks.ID) == "" {
				continue
			}
			ksName := valueOrEmpty(ks.Name)
			g.AppendSimpleResource(valueOrEmpty(ks.ID), ksName, "azurerm_cosmosdb_cassandra_keyspace")
			if err := appendFromPager(&g.AzureService, client.NewListCassandraTablesPager(rg, acc, ksName, nil),
				func(p armcosmos.CassandraResourcesClientListCassandraTablesResponse) []*armcosmos.CassandraTableGetResults {
					return p.Value
				},
				func(i *armcosmos.CassandraTableGetResults) string { return valueOrEmpty(i.ID) },
				func(i *armcosmos.CassandraTableGetResults) string { return valueOrEmpty(i.Name) },
				"azurerm_cosmosdb_cassandra_table"); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *CosmosDBGenerator) appendGremlin(client *armcosmos.GremlinResourcesClient, rg, acc string) error {
	dbPager := client.NewListGremlinDatabasesPager(rg, acc, nil)
	for dbPager.More() {
		page, err := dbPager.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, db := range page.Value {
			if db == nil || valueOrEmpty(db.ID) == "" {
				continue
			}
			dbName := valueOrEmpty(db.Name)
			g.AppendSimpleResource(valueOrEmpty(db.ID), dbName, "azurerm_cosmosdb_gremlin_database")
			if err := appendFromPager(&g.AzureService, client.NewListGremlinGraphsPager(rg, acc, dbName, nil),
				func(p armcosmos.GremlinResourcesClientListGremlinGraphsResponse) []*armcosmos.GremlinGraphGetResults {
					return p.Value
				},
				func(i *armcosmos.GremlinGraphGetResults) string { return valueOrEmpty(i.ID) },
				func(i *armcosmos.GremlinGraphGetResults) string { return valueOrEmpty(i.Name) },
				"azurerm_cosmosdb_gremlin_graph"); err != nil {
				return err
			}
		}
	}
	return nil
}
