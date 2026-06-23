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

	"github.com/aws/aws-sdk-go-v2/service/route53profiles"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type Route53ProfilesGenerator struct {
	AWSService
}

// InitResources enumerates Route 53 Profiles. Import ID is the profile id.
func (g *Route53ProfilesGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := route53profiles.NewFromConfig(config)

	p := route53profiles.NewListProfilesPaginator(svc, &route53profiles.ListProfilesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, profile := range page.ProfileSummaries {
			id := StringValue(profile.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(profile.Name), "aws_route53profiles_profile", "aws", defaultAllowEmptyValues))
			profileID := id
			for rp := route53profiles.NewListProfileResourceAssociationsPaginator(svc, &route53profiles.ListProfileResourceAssociationsInput{ProfileId: &profileID}); rp.HasMorePages(); {
				rpage, err := rp.NextPage(context.TODO())
				if err != nil {
					break
				}
				for _, ra := range rpage.ProfileResourceAssociations {
					raID := StringValue(ra.Id)
					if raID == "" {
						continue
					}
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						raID, raID, "aws_route53profiles_resource_association", "aws", defaultAllowEmptyValues))
				}
			}
		}
	}

	for ap := route53profiles.NewListProfileAssociationsPaginator(svc, &route53profiles.ListProfileAssociationsInput{}); ap.HasMorePages(); {
		page, err := ap.NextPage(context.TODO())
		if err != nil {
			break
		}
		for _, a := range page.ProfileAssociations {
			id := StringValue(a.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_route53profiles_association", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
