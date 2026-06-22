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
	"github.com/aws/aws-sdk-go-v2/service/kafka"
)

var mskAllowEmptyValues = []string{"tags."}

type MskGenerator struct {
	AWSService
}

func (g *MskGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := kafka.NewFromConfig(config)
	p := kafka.NewListClustersPaginator(svc, &kafka.ListClustersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, clusterInfo := range page.ClusterInfoList {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(clusterInfo.ClusterArn),
				StringValue(clusterInfo.ClusterName),
				"aws_msk_cluster",
				"aws",
				mskAllowEmptyValues,
			))
		}
	}

	ctx := context.TODO()
	for cp := kafka.NewListConfigurationsPaginator(svc, &kafka.ListConfigurationsInput{}); cp.HasMorePages(); {
		page, err := cp.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, c := range page.Configurations {
			arn := StringValue(c.Arn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(c.Name), "aws_msk_configuration", "aws", mskAllowEmptyValues))
		}
	}
	for rp := kafka.NewListReplicatorsPaginator(svc, &kafka.ListReplicatorsInput{}); rp.HasMorePages(); {
		page, err := rp.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, r := range page.Replicators {
			arn := StringValue(r.ReplicatorArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(r.ReplicatorName), "aws_msk_replicator", "aws", mskAllowEmptyValues))
		}
	}
	for vp := kafka.NewListVpcConnectionsPaginator(svc, &kafka.ListVpcConnectionsInput{}); vp.HasMorePages(); {
		page, err := vp.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, v := range page.VpcConnections {
			arn := StringValue(v.VpcConnectionArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_msk_vpc_connection", "aws", mskAllowEmptyValues))
		}
	}

	return nil
}

func (g *MskGenerator) PostConvertHook() error {
	for _, r := range g.Resources {
		if r.InstanceInfo.Type != "aws_msk_cluster" {
			continue
		}
		if r.InstanceState.Attributes["configuration_info.0.revision"] == "0" {
			delete(r.Item, "configuration_info")
		}
	}
	return nil
}
