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

	"github.com/aws/aws-sdk-go-v2/service/paymentcryptography"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type PaymentCryptographyGenerator struct {
	AWSService
}

// InitResources enumerates Payment Cryptography keys. Import ID is the key ARN.
func (g *PaymentCryptographyGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := paymentcryptography.NewFromConfig(config)

	p := paymentcryptography.NewListKeysPaginator(svc, &paymentcryptography.ListKeysInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, key := range page.Keys {
			arn := StringValue(key.KeyArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_paymentcryptography_key", "aws", defaultAllowEmptyValues))
		}
	}

	for ap := paymentcryptography.NewListAliasesPaginator(svc, &paymentcryptography.ListAliasesInput{}); ap.HasMorePages(); {
		page, err := ap.NextPage(context.TODO())
		if err != nil {
			break
		}
		for _, a := range page.Aliases {
			name := StringValue(a.AliasName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_paymentcryptography_key_alias", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
