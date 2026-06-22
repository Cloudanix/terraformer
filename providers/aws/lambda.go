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
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/smithy-go"
)

var lambdaAllowEmptyValues = []string{"tags."}

type LambdaGenerator struct {
	AWSService
}

type Statement struct {
	Sid string `json:"Sid"`
}

type Policy struct {
	Version   string       `json:"Version"`
	ID        string       `json:"Id"`
	Statement []*Statement `json:"Statement"`
}

func (g *LambdaGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := lambda.NewFromConfig(config)

	err := g.addFunctions(svc)
	if err != nil {
		return err
	}
	err = g.addEventSourceMappings(svc)
	if err != nil {
		return err
	}
	err = g.addLayerVersions(svc)
	if err != nil {
		return err
	}
	return g.addCodeSigningConfigs(svc)
}

func (g *LambdaGenerator) PostConvertHook() error {
	for i, r := range g.Resources {
		if _, exist := r.Item["environment"]; !exist {
			continue
		}
		variables := g.Resources[i].Item["environment"].([]interface{})[0].(map[string]interface{})["variables"]
		g.Resources[i].Item["environment"] = []interface{}{
			map[string]interface{}{
				"variables": []map[string]interface{}{variables.(map[string]interface{})},
			},
		}
	}
	for _, r := range g.Resources {
		if r.InstanceInfo.Type != "aws_lambda_function_event_invoke_config" {
			continue
		}
		if r.InstanceState.Attributes["maximum_event_age_in_seconds"] == "0" {
			delete(r.Item, "maximum_event_age_in_seconds")
		}
	}
	return nil
}

func (g *LambdaGenerator) addFunctions(svc *lambda.Client) error {
	p := lambda.NewListFunctionsPaginator(svc, &lambda.ListFunctionsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, function := range page.Functions {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				*function.FunctionArn,
				*function.FunctionName,
				"aws_lambda_function",
				"aws",
				map[string]string{
					"function_name": *function.FunctionName,
				},
				lambdaAllowEmptyValues,
				map[string]interface{}{},
			))

			gp, err := svc.GetPolicy(context.TODO(), &lambda.GetPolicyInput{
				FunctionName: aws.String(*function.FunctionArn),
			})

			if err != nil {
				// skip ResourceNotFoundException, because there may be only inline policy defined
				var apiErr smithy.APIError
				if !errors.As(err, &apiErr) || apiErr.ErrorCode() != "ResourceNotFoundException" {
					return err
				}
			}

			if gp != nil {
				outputPolicy := *gp.Policy
				var policy Policy
				err = json.Unmarshal([]byte(outputPolicy), &policy)

				if err != nil {
					return err
				}

				for _, statement := range policy.Statement {
					g.Resources = append(g.Resources, terraformutils.NewResource(
						statement.Sid,
						statement.Sid,
						"aws_lambda_permission",
						"aws",
						map[string]string{
							"statement_id":  statement.Sid,
							"function_name": *function.FunctionArn,
						},
						lambdaAllowEmptyValues,
						map[string]interface{}{},
					))
				}
			}

			pi := lambda.NewListFunctionEventInvokeConfigsPaginator(svc,
				&lambda.ListFunctionEventInvokeConfigsInput{
					FunctionName: function.FunctionName,
				})
			for pi.HasMorePages() {
				piage, err := pi.NextPage(context.TODO())
				if err != nil {
					return err
				}
				for _, functionEventInvokeConfig := range piage.FunctionEventInvokeConfigs {
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						*function.FunctionArn,
						"feic_"+*functionEventInvokeConfig.FunctionArn,
						"aws_lambda_function_event_invoke_config",
						"aws",
						lambdaAllowEmptyValues,
					))
				}
			}

			functionName := StringValue(function.FunctionName)

			pa := lambda.NewListAliasesPaginator(svc, &lambda.ListAliasesInput{FunctionName: function.FunctionName})
			for pa.HasMorePages() {
				apage, err := pa.NextPage(context.TODO())
				if err != nil {
					return err
				}
				for _, alias := range apage.Aliases {
					aliasName := StringValue(alias.Name)
					if aliasName == "" {
						continue
					}
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						functionName+"/"+aliasName,
						functionName+"_"+aliasName,
						"aws_lambda_alias",
						"aws",
						lambdaAllowEmptyValues,
					))
				}
			}

			pu := lambda.NewListFunctionUrlConfigsPaginator(svc, &lambda.ListFunctionUrlConfigsInput{FunctionName: function.FunctionName})
			for pu.HasMorePages() {
				upage, err := pu.NextPage(context.TODO())
				if err != nil {
					return err
				}
				for range upage.FunctionUrlConfigs {
					// One URL config per function qualifier; import ID is the function name.
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						functionName,
						functionName,
						"aws_lambda_function_url",
						"aws",
						lambdaAllowEmptyValues,
					))
				}
			}

			ppc := lambda.NewListProvisionedConcurrencyConfigsPaginator(svc, &lambda.ListProvisionedConcurrencyConfigsInput{FunctionName: function.FunctionName})
			for ppc.HasMorePages() {
				ppage, err := ppc.NextPage(context.TODO())
				if err != nil {
					break
				}
				for _, pc := range ppage.ProvisionedConcurrencyConfigs {
					arn := StringValue(pc.FunctionArn)
					parts := strings.Split(arn, ":")
					qualifier := parts[len(parts)-1]
					if qualifier == "" {
						continue
					}
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						functionName+","+qualifier,
						functionName+"_"+qualifier,
						"aws_lambda_provisioned_concurrency_config",
						"aws",
						lambdaAllowEmptyValues,
					))
				}
			}

			if rmc, err := svc.GetRuntimeManagementConfig(context.TODO(), &lambda.GetRuntimeManagementConfigInput{FunctionName: function.FunctionName}); err == nil &&
				rmc.UpdateRuntimeOn != "" && rmc.UpdateRuntimeOn != "Auto" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					StringValue(function.FunctionName),
					StringValue(function.FunctionName),
					"aws_lambda_runtime_management_config",
					"aws",
					lambdaAllowEmptyValues,
				))
			}
		}
	}
	return nil
}

func (g *LambdaGenerator) addCodeSigningConfigs(svc *lambda.Client) error {
	p := lambda.NewListCodeSigningConfigsPaginator(svc, &lambda.ListCodeSigningConfigsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, csc := range page.CodeSigningConfigs {
			arn := StringValue(csc.CodeSigningConfigArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_lambda_code_signing_config", "aws", lambdaAllowEmptyValues))
		}
	}
	return nil
}

func (g *LambdaGenerator) addEventSourceMappings(svc *lambda.Client) error {
	p := lambda.NewListEventSourceMappingsPaginator(svc, &lambda.ListEventSourceMappingsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, mapping := range page.EventSourceMappings {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				*mapping.UUID,
				*mapping.UUID,
				"aws_lambda_event_source_mapping",
				"aws",
				map[string]string{
					"event_source_arn": *mapping.EventSourceArn,
					"function_name":    *mapping.FunctionArn,
				},
				lambdaAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

func (g *LambdaGenerator) addLayerVersions(svc *lambda.Client) error {
	pl := lambda.NewListLayersPaginator(svc, &lambda.ListLayersInput{})
	for pl.HasMorePages() {
		plage, err := pl.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, layer := range plage.Layers {
			pv := lambda.NewListLayerVersionsPaginator(svc, &lambda.ListLayerVersionsInput{
				LayerName: layer.LayerName,
			})
			for pv.HasMorePages() {
				pvage, err := pv.NextPage(context.TODO())
				if err != nil {
					return err
				}
				for _, layerVersion := range pvage.LayerVersions {
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						*layerVersion.LayerVersionArn,
						*layerVersion.LayerVersionArn,
						"aws_lambda_layer_version",
						"aws",
						lambdaAllowEmptyValues,
					))
					if _, err := svc.GetLayerVersionPolicy(context.TODO(), &lambda.GetLayerVersionPolicyInput{
						LayerName: layer.LayerName, VersionNumber: &layerVersion.Version}); err == nil {
						layerName := StringValue(layer.LayerName)
						ver := strconv.FormatInt(layerVersion.Version, 10)
						g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
							layerName+","+ver, layerName+"_"+ver, "aws_lambda_layer_version_permission", "aws", lambdaAllowEmptyValues))
					}
				}
			}
		}
	}
	return nil
}
