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
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type OpenSearchServerlessGenerator struct {
	AWSService
}

// InitResources enumerates OpenSearch Serverless collections. Import ID is the
// collection id.
func (g *OpenSearchServerlessGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := opensearchserverless.NewFromConfig(config)

	p := opensearchserverless.NewListCollectionsPaginator(svc, &opensearchserverless.ListCollectionsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, c := range page.CollectionSummaries {
			id := StringValue(c.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(c.Name), "aws_opensearchserverless_collection", "aws", defaultAllowEmptyValues))
		}
	}

	ctx := awsContext()
	for _, scType := range types.SecurityConfigType("").Values() {
		t := scType
		for sp := opensearchserverless.NewListSecurityConfigsPaginator(svc, &opensearchserverless.ListSecurityConfigsInput{Type: t}); sp.HasMorePages(); {
			page, err := sp.NextPage(ctx)
			if err != nil {
				break
			}
			for _, s := range page.SecurityConfigSummaries {
				id := StringValue(s.Id)
				if id == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, id, "aws_opensearchserverless_security_config", "aws", defaultAllowEmptyValues))
			}
		}
	}
	for _, apType := range types.AccessPolicyType("").Values() {
		t := apType
		for pp := opensearchserverless.NewListAccessPoliciesPaginator(svc, &opensearchserverless.ListAccessPoliciesInput{Type: t}); pp.HasMorePages(); {
			page, err := pp.NextPage(ctx)
			if err != nil {
				break
			}
			for _, s := range page.AccessPolicySummaries {
				name := StringValue(s.Name)
				if name == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					name+"/"+string(s.Type), name, "aws_opensearchserverless_access_policy", "aws", defaultAllowEmptyValues))
			}
		}
	}
	for _, spType := range types.SecurityPolicyType("").Values() {
		t := spType
		for pp := opensearchserverless.NewListSecurityPoliciesPaginator(svc, &opensearchserverless.ListSecurityPoliciesInput{Type: t}); pp.HasMorePages(); {
			page, err := pp.NextPage(ctx)
			if err != nil {
				break
			}
			for _, s := range page.SecurityPolicySummaries {
				name := StringValue(s.Name)
				if name == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					name+"/"+string(s.Type), name, "aws_opensearchserverless_security_policy", "aws", defaultAllowEmptyValues))
			}
		}
	}
	for _, lpType := range types.LifecyclePolicyType("").Values() {
		t := lpType
		for pp := opensearchserverless.NewListLifecyclePoliciesPaginator(svc, &opensearchserverless.ListLifecyclePoliciesInput{Type: t}); pp.HasMorePages(); {
			page, err := pp.NextPage(ctx)
			if err != nil {
				break
			}
			for _, s := range page.LifecyclePolicySummaries {
				name := StringValue(s.Name)
				if name == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					name+"/"+string(s.Type), name, "aws_opensearchserverless_lifecycle_policy", "aws", defaultAllowEmptyValues))
			}
		}
	}
	for vp := opensearchserverless.NewListVpcEndpointsPaginator(svc, &opensearchserverless.ListVpcEndpointsInput{}); vp.HasMorePages(); {
		page, err := vp.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, v := range page.VpcEndpointSummaries {
			id := StringValue(v.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(v.Name), "aws_opensearchserverless_vpc_endpoint", "aws", defaultAllowEmptyValues))
		}
	}

	for cg := opensearchserverless.NewListCollectionGroupsPaginator(svc, &opensearchserverless.ListCollectionGroupsInput{}); cg.HasMorePages(); {
		page, err := cg.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, c := range page.CollectionGroupSummaries {
			id := StringValue(c.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(c.Name), "aws_opensearchserverless_collection_group", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
