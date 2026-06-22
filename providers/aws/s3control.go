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
	"github.com/aws/aws-sdk-go-v2/service/s3control"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type S3ControlGenerator struct {
	AWSService
}

// InitResources enumerates S3 Control storage lens configurations for the
// account. Import ID is "<account-id>:<config-id>".
func (g *S3ControlGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	account, err := g.getAccountNumber(config)
	if err != nil {
		return err
	}
	accountID := StringValue(account)
	svc := s3control.NewFromConfig(config)

	p := s3control.NewListStorageLensConfigurationsPaginator(svc, &s3control.ListStorageLensConfigurationsInput{AccountId: aws.String(accountID)})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, cfg := range page.StorageLensConfigurationList {
			id := StringValue(cfg.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				accountID+":"+id, id, "aws_s3control_storage_lens_configuration", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
