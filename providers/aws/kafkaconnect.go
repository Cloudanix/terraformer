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
	"github.com/aws/aws-sdk-go-v2/service/kafkaconnect"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type KafkaConnectGenerator struct {
	AWSService
}

// InitResources enumerates MSK Connect connectors. Import ID is the connector ARN.
func (g *KafkaConnectGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := kafkaconnect.NewFromConfig(config)

	p := kafkaconnect.NewListConnectorsPaginator(svc, &kafkaconnect.ListConnectorsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, conn := range page.Connectors {
			arn := StringValue(conn.ConnectorArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(conn.ConnectorName), "aws_mskconnect_connector", "aws", defaultAllowEmptyValues))
		}
	}

	for cp := kafkaconnect.NewListCustomPluginsPaginator(svc, &kafkaconnect.ListCustomPluginsInput{}); cp.HasMorePages(); {
		page, err := cp.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, p := range page.CustomPlugins {
			arn := StringValue(p.CustomPluginArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(p.Name), "aws_mskconnect_custom_plugin", "aws", defaultAllowEmptyValues))
		}
	}
	for wp := kafkaconnect.NewListWorkerConfigurationsPaginator(svc, &kafkaconnect.ListWorkerConfigurationsInput{}); wp.HasMorePages(); {
		page, err := wp.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, w := range page.WorkerConfigurations {
			arn := StringValue(w.WorkerConfigurationArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(w.Name), "aws_mskconnect_worker_configuration", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
