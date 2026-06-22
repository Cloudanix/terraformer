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

	"github.com/aws/aws-sdk-go-v2/service/securitylake"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type SecurityLakeGenerator struct {
	AWSService
}

// InitResources enumerates Security Lake data lakes and subscribers. Import IDs
// are the resource ARN.
func (g *SecurityLakeGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := securitylake.NewFromConfig(config)
	ctx := context.TODO()

	lakes, err := svc.ListDataLakes(ctx, &securitylake.ListDataLakesInput{})
	if err != nil {
		return err
	}
	for _, lake := range lakes.DataLakes {
		arn := StringValue(lake.DataLakeArn)
		if arn == "" {
			continue
		}
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			arn, arn, "aws_securitylake_data_lake", "aws", defaultAllowEmptyValues))
	}

	subscribers := securitylake.NewListSubscribersPaginator(svc, &securitylake.ListSubscribersInput{})
	for subscribers.HasMorePages() {
		page, err := subscribers.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, sub := range page.Subscribers {
			arn := StringValue(sub.SubscriberArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_securitylake_subscriber", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
