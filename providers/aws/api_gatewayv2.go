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

package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var apiGatewayV2AllowEmptyValues = []string{"tags."}

type APIGatewayV2Generator struct {
	AWSService
}

// InitResources enumerates API Gateway v2 (HTTP/WebSocket) resources via the
// aws-sdk-go-v2 client. Per-API children use composite import IDs the aws
// provider expects, e.g. "<api-id>/<route-id>".
func (g *APIGatewayV2Generator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := apigatewayv2.NewFromConfig(config)
	ctx := context.TODO()

	apiIDs, err := g.loadAPIs(ctx, svc)
	if err != nil {
		return err
	}
	for _, apiID := range apiIDs {
		if err := g.loadAPIChildren(ctx, svc, apiID); err != nil {
			return err
		}
	}
	if err := g.loadVPCLinks(ctx, svc); err != nil {
		return err
	}
	return g.loadDomainNames(ctx, svc)
}

func (g *APIGatewayV2Generator) add(id, name, tfType string) {
	if id == "" {
		return
	}
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
		id, name, tfType, "aws", apiGatewayV2AllowEmptyValues))
}

func (g *APIGatewayV2Generator) loadAPIs(ctx context.Context, svc *apigatewayv2.Client) ([]string, error) {
	var ids []string
	var token *string
	for {
		out, err := svc.GetApis(ctx, &apigatewayv2.GetApisInput{NextToken: token})
		if err != nil {
			return nil, err
		}
		for _, api := range out.Items {
			id := StringValue(api.ApiId)
			if id == "" {
				continue
			}
			ids = append(ids, id)
			g.add(id, StringValue(api.Name), "aws_apigatewayv2_api")
		}
		if out.NextToken == nil {
			return ids, nil
		}
		token = out.NextToken
	}
}

func (g *APIGatewayV2Generator) loadAPIChildren(ctx context.Context, svc *apigatewayv2.Client, apiID string) error {
	authorizers, err := svc.GetAuthorizers(ctx, &apigatewayv2.GetAuthorizersInput{ApiId: aws.String(apiID)})
	if err != nil {
		return err
	}
	for _, a := range authorizers.Items {
		g.add(fmt.Sprintf("%s/%s", apiID, StringValue(a.AuthorizerId)), StringValue(a.Name), "aws_apigatewayv2_authorizer")
	}

	models, err := svc.GetModels(ctx, &apigatewayv2.GetModelsInput{ApiId: aws.String(apiID)})
	if err != nil {
		return err
	}
	for _, m := range models.Items {
		g.add(fmt.Sprintf("%s/%s", apiID, StringValue(m.ModelId)), StringValue(m.Name), "aws_apigatewayv2_model")
	}

	integrations, err := svc.GetIntegrations(ctx, &apigatewayv2.GetIntegrationsInput{ApiId: aws.String(apiID)})
	if err != nil {
		return err
	}
	for _, i := range integrations.Items {
		id := StringValue(i.IntegrationId)
		g.add(fmt.Sprintf("%s/%s", apiID, id), fmt.Sprintf("%s_%s", apiID, id), "aws_apigatewayv2_integration")
	}

	routes, err := svc.GetRoutes(ctx, &apigatewayv2.GetRoutesInput{ApiId: aws.String(apiID)})
	if err != nil {
		return err
	}
	for _, r := range routes.Items {
		routeID := StringValue(r.RouteId)
		g.add(fmt.Sprintf("%s/%s", apiID, routeID), fmt.Sprintf("%s_%s", apiID, routeID), "aws_apigatewayv2_route")

		responses, err := svc.GetRouteResponses(ctx, &apigatewayv2.GetRouteResponsesInput{
			ApiId: aws.String(apiID), RouteId: aws.String(routeID),
		})
		if err != nil {
			return err
		}
		for _, rr := range responses.Items {
			rrID := StringValue(rr.RouteResponseId)
			g.add(fmt.Sprintf("%s/%s/%s", apiID, routeID, rrID), fmt.Sprintf("%s_%s", routeID, rrID), "aws_apigatewayv2_route_response")
		}
	}

	stages, err := svc.GetStages(ctx, &apigatewayv2.GetStagesInput{ApiId: aws.String(apiID)})
	if err != nil {
		return err
	}
	for _, s := range stages.Items {
		stageName := StringValue(s.StageName)
		g.add(fmt.Sprintf("%s/%s", apiID, stageName), fmt.Sprintf("%s_%s", apiID, stageName), "aws_apigatewayv2_stage")
	}

	deployments, err := svc.GetDeployments(ctx, &apigatewayv2.GetDeploymentsInput{ApiId: aws.String(apiID)})
	if err != nil {
		return err
	}
	for _, d := range deployments.Items {
		depID := StringValue(d.DeploymentId)
		g.add(fmt.Sprintf("%s/%s", apiID, depID), fmt.Sprintf("%s_%s", apiID, depID), "aws_apigatewayv2_deployment")
	}
	return nil
}

func (g *APIGatewayV2Generator) loadVPCLinks(ctx context.Context, svc *apigatewayv2.Client) error {
	var token *string
	for {
		out, err := svc.GetVpcLinks(ctx, &apigatewayv2.GetVpcLinksInput{NextToken: token})
		if err != nil {
			return err
		}
		for _, vl := range out.Items {
			g.add(StringValue(vl.VpcLinkId), StringValue(vl.Name), "aws_apigatewayv2_vpc_link")
		}
		if out.NextToken == nil {
			return nil
		}
		token = out.NextToken
	}
}

func (g *APIGatewayV2Generator) loadDomainNames(ctx context.Context, svc *apigatewayv2.Client) error {
	var token *string
	for {
		out, err := svc.GetDomainNames(ctx, &apigatewayv2.GetDomainNamesInput{NextToken: token})
		if err != nil {
			return err
		}
		for _, d := range out.Items {
			domain := StringValue(d.DomainName)
			g.add(domain, domain, "aws_apigatewayv2_domain_name")

			mappings, err := svc.GetApiMappings(ctx, &apigatewayv2.GetApiMappingsInput{DomainName: aws.String(domain)})
			if err != nil {
				return err
			}
			for _, m := range mappings.Items {
				mapID := StringValue(m.ApiMappingId)
				g.add(fmt.Sprintf("%s/%s", mapID, domain), fmt.Sprintf("%s_%s", domain, mapID), "aws_apigatewayv2_api_mapping")
			}
		}
		if out.NextToken == nil {
			return nil
		}
		token = out.NextToken
	}
}
