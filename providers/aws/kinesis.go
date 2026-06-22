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

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
)

var kinesisAllowEmptyValues = []string{"tags."}

type KinesisGenerator struct {
	AWSService
}

func (g *KinesisGenerator) createResources(streamNames []string) []terraformutils.Resource {
	var resources []terraformutils.Resource
	for _, resourceName := range streamNames {
		resources = append(resources, terraformutils.NewResource(
			resourceName,
			resourceName,
			"aws_kinesis_stream",
			"aws",
			map[string]string{"name": resourceName},
			kinesisAllowEmptyValues,
			map[string]interface{}{}))
	}
	return resources
}

func (g *KinesisGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := kinesis.NewFromConfig(config)

	var results *kinesis.ListStreamsOutput
	var request = kinesis.ListStreamsInput{}
	var err error

	for results == nil || *results.HasMoreStreams {
		results, err = svc.ListStreams(context.TODO(), &request)
		if err != nil {
			return err
		}

		g.Resources = append(g.Resources, g.createResources(results.StreamNames)...)

		// StreamSummaries carry the ARN needed to enumerate registered consumers.
		for _, summary := range results.StreamSummaries {
			streamARN := StringValue(summary.StreamARN)
			if err := g.addStreamConsumers(svc, streamARN); err != nil {
				return err
			}
			if streamARN == "" {
				continue
			}
			if rp, err := svc.GetResourcePolicy(context.TODO(), &kinesis.GetResourcePolicyInput{ResourceARN: &streamARN}); err == nil && StringValue(rp.Policy) != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					streamARN, streamARN, "aws_kinesis_resource_policy", "aws", []string{}))
			}
		}

		if len(results.StreamNames) > 0 {
			request = kinesis.ListStreamsInput{
				ExclusiveStartStreamName: &results.StreamNames[len(results.StreamNames)-1],
			}
		}
	}
	return nil
}

func (g *KinesisGenerator) addStreamConsumers(svc *kinesis.Client, streamARN string) error {
	if streamARN == "" {
		return nil
	}
	p := kinesis.NewListStreamConsumersPaginator(svc, &kinesis.ListStreamConsumersInput{StreamARN: &streamARN})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, consumer := range page.Consumers {
			arn := StringValue(consumer.ConsumerARN)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(consumer.ConsumerName), "aws_kinesis_stream_consumer", "aws", kinesisAllowEmptyValues))
		}
	}
	return nil
}
