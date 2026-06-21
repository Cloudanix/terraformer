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

	"github.com/aws/aws-sdk-go-v2/service/transfer"
	"github.com/aws/aws-sdk-go-v2/service/transfer/types"
)

type TransferGenerator struct {
	AWSService
}

// InitResources enumerates AWS Transfer Family servers. The Terraform import ID
// for aws_transfer_server is the server ID.
func (g *TransferGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := transfer.NewFromConfig(config)

	p := transfer.NewListServersPaginator(svc, &transfer.ListServersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.Servers, "aws_transfer_server",
			defaultAllowEmptyValues,
			func(s types.ListedServer) string { return StringValue(s.ServerId) },
			func(s types.ListedServer) string { return StringValue(s.ServerId) })
	}
	return nil
}
