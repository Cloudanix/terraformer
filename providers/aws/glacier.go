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
	"github.com/aws/aws-sdk-go-v2/service/glacier"
	"github.com/aws/aws-sdk-go-v2/service/glacier/types"
)

type GlacierGenerator struct {
	AWSService
}

// InitResources enumerates S3 Glacier vaults. The Terraform import ID for
// aws_glacier_vault is the vault name. ListVaults requires an AccountId; "-"
// resolves to the account that owns the credentials.
func (g *GlacierGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := glacier.NewFromConfig(config)

	p := glacier.NewListVaultsPaginator(svc, &glacier.ListVaultsInput{AccountId: aws.String("-")})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.VaultList, "aws_glacier_vault",
			defaultAllowEmptyValues,
			func(v types.DescribeVaultOutput) string { return StringValue(v.VaultName) },
			func(v types.DescribeVaultOutput) string { return StringValue(v.VaultName) })
	}
	return nil
}
