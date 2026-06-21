// Copyright 2018 The Terraformer Authors.
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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/globalaccelerator"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type GlobalAcceleratorGenerator struct {
	AWSService
}

// InitResources enumerates Global Accelerator standard accelerators and their
// listener/endpoint-group hierarchy, plus custom-routing accelerators. Global
// Accelerator is a partition-global service (endpoint signed for us-west-2), so
// it is imported in the aws-global pass. Every import ID is the resource ARN.
func (g *GlobalAcceleratorGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := globalaccelerator.NewFromConfig(config)
	ctx := context.TODO()

	var acceleratorArns []string
	accelerators := globalaccelerator.NewListAcceleratorsPaginator(svc, &globalaccelerator.ListAcceleratorsInput{})
	for accelerators.HasMorePages() {
		page, err := accelerators.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, a := range page.Accelerators {
			arn := StringValue(a.AcceleratorArn)
			if arn == "" {
				continue
			}
			acceleratorArns = append(acceleratorArns, arn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(a.Name), "aws_globalaccelerator_accelerator", "aws", defaultAllowEmptyValues))
		}
	}

	for _, acceleratorArn := range acceleratorArns {
		var listenerArns []string
		listeners := globalaccelerator.NewListListenersPaginator(svc, &globalaccelerator.ListListenersInput{AcceleratorArn: aws.String(acceleratorArn)})
		for listeners.HasMorePages() {
			page, err := listeners.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, l := range page.Listeners {
				arn := StringValue(l.ListenerArn)
				if arn == "" {
					continue
				}
				listenerArns = append(listenerArns, arn)
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					arn, arn, "aws_globalaccelerator_listener", "aws", defaultAllowEmptyValues))
			}
		}

		for _, listenerArn := range listenerArns {
			groups := globalaccelerator.NewListEndpointGroupsPaginator(svc, &globalaccelerator.ListEndpointGroupsInput{ListenerArn: aws.String(listenerArn)})
			for groups.HasMorePages() {
				page, err := groups.NextPage(ctx)
				if err != nil {
					return err
				}
				for _, eg := range page.EndpointGroups {
					arn := StringValue(eg.EndpointGroupArn)
					if arn == "" {
						continue
					}
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						arn, arn, "aws_globalaccelerator_endpoint_group", "aws", defaultAllowEmptyValues))
				}
			}
		}
	}

	customAccelerators := globalaccelerator.NewListCustomRoutingAcceleratorsPaginator(svc, &globalaccelerator.ListCustomRoutingAcceleratorsInput{})
	for customAccelerators.HasMorePages() {
		page, err := customAccelerators.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, a := range page.Accelerators {
			arn := StringValue(a.AcceleratorArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(a.Name), "aws_globalaccelerator_custom_routing_accelerator", "aws", defaultAllowEmptyValues))
		}
	}

	return nil
}
