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

	"github.com/aws/aws-sdk-go-v2/service/apprunner"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type AppRunnerGenerator struct {
	AWSService
}

// InitResources enumerates App Runner services. Import ID is the service ARN.
func (g *AppRunnerGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := apprunner.NewFromConfig(config)

	ctx := context.TODO()
	add := func(arn, tfType string) {
		if arn != "" {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, tfType, "aws", defaultAllowEmptyValues))
		}
	}

	p := apprunner.NewListServicesPaginator(svc, &apprunner.ListServicesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, service := range page.ServiceSummaryList {
			arn := StringValue(service.ServiceArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(service.ServiceName), "aws_apprunner_service", "aws", defaultAllowEmptyValues))
		}
	}

	for c := apprunner.NewListConnectionsPaginator(svc, &apprunner.ListConnectionsInput{}); c.HasMorePages(); {
		page, err := c.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range page.ConnectionSummaryList {
			add(StringValue(x.ConnectionArn), "aws_apprunner_connection")
		}
	}
	for a := apprunner.NewListAutoScalingConfigurationsPaginator(svc, &apprunner.ListAutoScalingConfigurationsInput{}); a.HasMorePages(); {
		page, err := a.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range page.AutoScalingConfigurationSummaryList {
			add(StringValue(x.AutoScalingConfigurationArn), "aws_apprunner_auto_scaling_configuration_version")
		}
	}
	for v := apprunner.NewListVpcConnectorsPaginator(svc, &apprunner.ListVpcConnectorsInput{}); v.HasMorePages(); {
		page, err := v.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range page.VpcConnectors {
			add(StringValue(x.VpcConnectorArn), "aws_apprunner_vpc_connector")
		}
	}
	for o := apprunner.NewListObservabilityConfigurationsPaginator(svc, &apprunner.ListObservabilityConfigurationsInput{}); o.HasMorePages(); {
		page, err := o.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range page.ObservabilityConfigurationSummaryList {
			add(StringValue(x.ObservabilityConfigurationArn), "aws_apprunner_observability_configuration")
		}
	}
	for i := apprunner.NewListVpcIngressConnectionsPaginator(svc, &apprunner.ListVpcIngressConnectionsInput{}); i.HasMorePages(); {
		page, err := i.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range page.VpcIngressConnectionSummaryList {
			add(StringValue(x.VpcIngressConnectionArn), "aws_apprunner_vpc_ingress_connection")
		}
	}
	return nil
}
