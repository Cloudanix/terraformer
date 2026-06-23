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
	"github.com/aws/aws-sdk-go-v2/service/ram"
	"github.com/aws/aws-sdk-go-v2/service/ram/types"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type RamGenerator struct {
	AWSService
}

// InitResources enumerates self-owned Resource Access Manager resource shares.
// Import ID is the resource share ARN.
func (g *RamGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := ram.NewFromConfig(config)

	p := ram.NewGetResourceSharesPaginator(svc, &ram.GetResourceSharesInput{
		ResourceOwner: types.ResourceOwnerSelf,
	})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, share := range page.ResourceShares {
			arn := StringValue(share.ResourceShareArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_ram_resource_share", "aws", defaultAllowEmptyValues))
		}
	}

	// Principal and resource associations on self-owned shares.
	assocTypes := map[types.ResourceShareAssociationType]string{
		types.ResourceShareAssociationTypePrincipal: "aws_ram_principal_association",
		types.ResourceShareAssociationTypeResource:  "aws_ram_resource_association",
	}
	for at, tfType := range assocTypes {
		ap := ram.NewGetResourceShareAssociationsPaginator(svc, &ram.GetResourceShareAssociationsInput{AssociationType: at})
		for ap.HasMorePages() {
			page, err := ap.NextPage(awsContext())
			if err != nil {
				break
			}
			for _, a := range page.ResourceShareAssociations {
				shareArn, entity := StringValue(a.ResourceShareArn), StringValue(a.AssociatedEntity)
				if shareArn == "" || entity == "" || a.Status != types.ResourceShareAssociationStatusAssociated {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					shareArn+","+entity, shareArn+"_"+entity, tfType, "aws", defaultAllowEmptyValues))
			}
		}
	}
	return nil
}
