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
		}
		nextToken = apis.NextToken
		if nextToken == nil {
			break
		}
	}

	return nil
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
