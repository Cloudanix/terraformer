package aws

import (
	"context"
	"fmt"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appsync"
)

type AppSyncGenerator struct {
	AWSService
}

func (g *AppSyncGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}

	svc := appsync.NewFromConfig(config)

	var nextToken *string
	for {
		apis, err := svc.ListGraphqlApis(context.TODO(), &appsync.ListGraphqlApisInput{
			NextToken: nextToken,
		})
		if err != nil {
			return err
		}

		for _, api := range apis.GraphqlApis {
			var id = *api.ApiId
			var name = *api.Name
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				name,
				"aws_appsync_graphql_api",
				"aws",
				[]string{}))
			g.loadAppSyncChildren(svc, id)
			g.loadAppSyncTypesAndResolvers(svc, id)
			g.loadAppSyncApiCache(svc, id)
		}
		nextToken = apis.NextToken
		if nextToken == nil {
			break
		}
	}

	g.loadAppSyncDomainNames(svc)

	return nil
}

// loadAppSyncTypesAndResolvers enumerates an API's SDL types and their resolvers.
// Import IDs: "<api-id>:SDL:<type>" for types, "<api-id>-<type>-<field>" for resolvers.
func (g *AppSyncGenerator) loadAppSyncTypesAndResolvers(svc *appsync.Client, apiID string) {
	ctx := context.TODO()
	out, err := svc.ListTypes(ctx, &appsync.ListTypesInput{ApiId: aws.String(apiID), Format: "SDL"})
	if err != nil {
		return
	}
	for _, t := range out.Types {
		typeName := StringValue(t.Name)
		if typeName == "" {
			continue
		}
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			fmt.Sprintf("%s:SDL:%s", apiID, typeName), fmt.Sprintf("%s_%s", apiID, typeName),
			"aws_appsync_type", "aws", defaultAllowEmptyValues))
		res, err := svc.ListResolvers(ctx, &appsync.ListResolversInput{ApiId: aws.String(apiID), TypeName: aws.String(typeName)})
		if err != nil {
			continue
		}
		for _, r := range res.Resolvers {
			field := StringValue(r.FieldName)
			if field == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				fmt.Sprintf("%s-%s-%s", apiID, typeName, field), fmt.Sprintf("%s_%s_%s", apiID, typeName, field),
				"aws_appsync_resolver", "aws", defaultAllowEmptyValues))
		}
	}
}

// loadAppSyncApiCache emits the per-API cache (import ID is the API ID).
func (g *AppSyncGenerator) loadAppSyncApiCache(svc *appsync.Client, apiID string) {
	if out, err := svc.GetApiCache(context.TODO(), &appsync.GetApiCacheInput{ApiId: aws.String(apiID)}); err == nil && out.ApiCache != nil {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			apiID, apiID, "aws_appsync_api_cache", "aws", defaultAllowEmptyValues))
	}
}

// loadAppSyncDomainNames enumerates custom domain names (import ID is the domain).
func (g *AppSyncGenerator) loadAppSyncDomainNames(svc *appsync.Client) {
	out, err := svc.ListDomainNames(context.TODO(), &appsync.ListDomainNamesInput{})
	if err != nil {
		return
	}
	for _, d := range out.DomainNameConfigs {
		name := StringValue(d.DomainName)
		if name == "" {
			continue
		}
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			name, name, "aws_appsync_domain_name", "aws", defaultAllowEmptyValues))
	}
}

// loadAppSyncChildren enumerates an API's data sources, functions, and API keys.
// Import IDs: "<api-id>-<datasource-name>", "<api-id>-<function-id>",
// "<api-id>:<key-id>".
func (g *AppSyncGenerator) loadAppSyncChildren(svc *appsync.Client, apiID string) {
	ctx := context.TODO()
	if out, err := svc.ListDataSources(ctx, &appsync.ListDataSourcesInput{ApiId: aws.String(apiID)}); err == nil {
		for _, ds := range out.DataSources {
			name := StringValue(ds.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				fmt.Sprintf("%s-%s", apiID, name), fmt.Sprintf("%s_%s", apiID, name),
				"aws_appsync_datasource", "aws", defaultAllowEmptyValues))
		}
	}
	if out, err := svc.ListFunctions(ctx, &appsync.ListFunctionsInput{ApiId: aws.String(apiID)}); err == nil {
		for _, fn := range out.Functions {
			fid := StringValue(fn.FunctionId)
			if fid == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				fmt.Sprintf("%s-%s", apiID, fid), fmt.Sprintf("%s_%s", apiID, fid),
				"aws_appsync_function", "aws", defaultAllowEmptyValues))
		}
	}
	if out, err := svc.ListApiKeys(ctx, &appsync.ListApiKeysInput{ApiId: aws.String(apiID)}); err == nil {
		for _, k := range out.ApiKeys {
			kid := StringValue(k.Id)
			if kid == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				fmt.Sprintf("%s:%s", apiID, kid), fmt.Sprintf("%s_%s", apiID, kid),
				"aws_appsync_api_key", "aws", defaultAllowEmptyValues))
		}
	}
}
