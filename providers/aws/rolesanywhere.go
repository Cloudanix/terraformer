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

	"github.com/aws/aws-sdk-go-v2/service/rolesanywhere"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type RolesAnywhereGenerator struct {
	AWSService
}

// InitResources enumerates IAM Roles Anywhere trust anchors and profiles.
// Import IDs are the resource id.
func (g *RolesAnywhereGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := rolesanywhere.NewFromConfig(config)
	ctx := context.TODO()

	anchors := rolesanywhere.NewListTrustAnchorsPaginator(svc, &rolesanywhere.ListTrustAnchorsInput{})
	for anchors.HasMorePages() {
		page, err := anchors.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, anchor := range page.TrustAnchors {
			id := StringValue(anchor.TrustAnchorId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_rolesanywhere_trust_anchor", "aws", defaultAllowEmptyValues))
		}
	}

	profiles := rolesanywhere.NewListProfilesPaginator(svc, &rolesanywhere.ListProfilesInput{})
	for profiles.HasMorePages() {
		page, err := profiles.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, profile := range page.Profiles {
			id := StringValue(profile.ProfileId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_rolesanywhere_profile", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
