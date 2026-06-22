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

	ctx := context.TODO()
	if insts, err := svc.ListAccessGrantsInstances(ctx, &s3control.ListAccessGrantsInstancesInput{AccountId: aws.String(accountID)}); err == nil {
		if len(insts.AccessGrantsInstancesList) > 0 {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				accountID, accountID, "aws_s3control_access_grants_instance", "aws", defaultAllowEmptyValues))
		}
	}
	for p := s3control.NewListAccessGrantsLocationsPaginator(svc, &s3control.ListAccessGrantsLocationsInput{AccountId: aws.String(accountID)}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, loc := range page.AccessGrantsLocationsList {
			id := StringValue(loc.AccessGrantsLocationId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				accountID+":"+id, id, "aws_s3control_access_grants_location", "aws", defaultAllowEmptyValues))
		}
	}
	for p := s3control.NewListAccessGrantsPaginator(svc, &s3control.ListAccessGrantsInput{AccountId: aws.String(accountID)}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, grant := range page.AccessGrantsList {
			id := StringValue(grant.AccessGrantId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				accountID+":"+id, id, "aws_s3control_access_grant", "aws", defaultAllowEmptyValues))
		}
	}

	for p := s3control.NewListMultiRegionAccessPointsPaginator(svc, &s3control.ListMultiRegionAccessPointsInput{AccountId: aws.String(accountID)}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, ap := range page.AccessPoints {
			name := StringValue(ap.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				accountID+":"+name, name, "aws_s3control_multi_region_access_point", "aws", defaultAllowEmptyValues))
		}
	}

	for p := s3control.NewListAccessPointsForObjectLambdaPaginator(svc, &s3control.ListAccessPointsForObjectLambdaInput{AccountId: aws.String(accountID)}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, ap := range page.ObjectLambdaAccessPointList {
			name := StringValue(ap.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				accountID+":"+name, name, "aws_s3control_object_lambda_access_point", "aws", defaultAllowEmptyValues))
		}
	}

	for p := s3control.NewListAccessPointsPaginator(svc, &s3control.ListAccessPointsInput{AccountId: aws.String(accountID)}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, ap := range page.AccessPointList {
			name := StringValue(ap.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				accountID+":"+name, name, "aws_s3_access_point", "aws", defaultAllowEmptyValues))
		}
	}

	if _, err := svc.GetPublicAccessBlock(ctx, &s3control.GetPublicAccessBlockInput{AccountId: aws.String(accountID)}); err == nil {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			accountID, accountID, "aws_s3_account_public_access_block", "aws", defaultAllowEmptyValues))
	}
	return nil
}
