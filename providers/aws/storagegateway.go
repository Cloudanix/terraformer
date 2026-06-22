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

	"github.com/aws/aws-sdk-go-v2/service/storagegateway"
	"github.com/aws/aws-sdk-go-v2/service/storagegateway/types"
)

type StorageGatewayGenerator struct {
	AWSService
}

// InitResources enumerates Storage Gateway gateways. The Terraform import ID for
// aws_storagegateway_gateway is the gateway ARN.
func (g *StorageGatewayGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := storagegateway.NewFromConfig(config)

	p := storagegateway.NewListGatewaysPaginator(svc, &storagegateway.ListGatewaysInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.Gateways, "aws_storagegateway_gateway",
			defaultAllowEmptyValues,
			func(gw types.GatewayInfo) string { return StringValue(gw.GatewayARN) },
			func(gw types.GatewayInfo) string { return StringValue(gw.GatewayName) })
	}

	pools := storagegateway.NewListTapePoolsPaginator(svc, &storagegateway.ListTapePoolsInput{})
	for pools.HasMorePages() {
		page, err := pools.NextPage(context.TODO())
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.PoolInfos, "aws_storagegateway_tape_pool",
			defaultAllowEmptyValues,
			func(p types.PoolInfo) string { return StringValue(p.PoolARN) },
			func(p types.PoolInfo) string { return StringValue(p.PoolName) })
	}
	return nil
}
