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
	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type VerifiedPermissionsGenerator struct {
	AWSService
}

// InitResources enumerates Verified Permissions policy stores. Import ID is the
// policy store id.
func (g *VerifiedPermissionsGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := verifiedpermissions.NewFromConfig(config)

	ctx := awsContext()
	var storeIDs []string
	p := verifiedpermissions.NewListPolicyStoresPaginator(svc, &verifiedpermissions.ListPolicyStoresInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, ps := range page.PolicyStores {
			id := StringValue(ps.PolicyStoreId)
			if id == "" {
				continue
			}
			storeIDs = append(storeIDs, id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_verifiedpermissions_policy_store", "aws", defaultAllowEmptyValues))
		}
	}

	for _, storeID := range storeIDs {
		store := storeID
		for pp := verifiedpermissions.NewListPoliciesPaginator(svc, &verifiedpermissions.ListPoliciesInput{PolicyStoreId: &store}); pp.HasMorePages(); {
			page, err := pp.NextPage(ctx)
			if err != nil {
				break
			}
			for _, pol := range page.Policies {
				id := StringValue(pol.PolicyId)
				if id == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id+":"+store, store+"_"+id, "aws_verifiedpermissions_policy", "aws", defaultAllowEmptyValues))
			}
		}
		for tp := verifiedpermissions.NewListPolicyTemplatesPaginator(svc, &verifiedpermissions.ListPolicyTemplatesInput{PolicyStoreId: &store}); tp.HasMorePages(); {
			page, err := tp.NextPage(ctx)
			if err != nil {
				break
			}
			for _, t := range page.PolicyTemplates {
				id := StringValue(t.PolicyTemplateId)
				if id == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					store+":"+id, store+"_"+id, "aws_verifiedpermissions_policy_template", "aws", defaultAllowEmptyValues))
			}
		}
		for ip := verifiedpermissions.NewListIdentitySourcesPaginator(svc, &verifiedpermissions.ListIdentitySourcesInput{PolicyStoreId: &store}); ip.HasMorePages(); {
			page, err := ip.NextPage(ctx)
			if err != nil {
				break
			}
			for _, is := range page.IdentitySources {
				id := StringValue(is.IdentitySourceId)
				if id == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					store+":"+id, store+"_"+id, "aws_verifiedpermissions_identity_source", "aws", defaultAllowEmptyValues))
			}
		}
		if _, err := svc.GetSchema(ctx, &verifiedpermissions.GetSchemaInput{PolicyStoreId: &store}); err == nil {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				store, store, "aws_verifiedpermissions_schema", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
