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

	"github.com/aws/aws-sdk-go-v2/service/ec2"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type VerifiedAccessGenerator struct {
	AWSService
}

// InitResources enumerates Verified Access instances and trust providers (both
// served by the EC2 API). Import IDs are the resource ids.
func (g *VerifiedAccessGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := ec2.NewFromConfig(config)
	ctx := context.TODO()

	instances := ec2.NewDescribeVerifiedAccessInstancesPaginator(svc, &ec2.DescribeVerifiedAccessInstancesInput{})
	for instances.HasMorePages() {
		page, err := instances.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, vai := range page.VerifiedAccessInstances {
			id := StringValue(vai.VerifiedAccessInstanceId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_verifiedaccess_instance", "aws", defaultAllowEmptyValues))
		}
	}

	trustProviders := ec2.NewDescribeVerifiedAccessTrustProvidersPaginator(svc, &ec2.DescribeVerifiedAccessTrustProvidersInput{})
	for trustProviders.HasMorePages() {
		page, err := trustProviders.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, tp := range page.VerifiedAccessTrustProviders {
			id := StringValue(tp.VerifiedAccessTrustProviderId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_verifiedaccess_trust_provider", "aws", defaultAllowEmptyValues))
		}
	}

	groups := ec2.NewDescribeVerifiedAccessGroupsPaginator(svc, &ec2.DescribeVerifiedAccessGroupsInput{})
	for groups.HasMorePages() {
		page, err := groups.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, grp := range page.VerifiedAccessGroups {
			id := StringValue(grp.VerifiedAccessGroupId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_verifiedaccess_group", "aws", defaultAllowEmptyValues))
		}
	}

	endpoints := ec2.NewDescribeVerifiedAccessEndpointsPaginator(svc, &ec2.DescribeVerifiedAccessEndpointsInput{})
	for endpoints.HasMorePages() {
		page, err := endpoints.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, ep := range page.VerifiedAccessEndpoints {
			id := StringValue(ep.VerifiedAccessEndpointId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_verifiedaccess_endpoint", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
