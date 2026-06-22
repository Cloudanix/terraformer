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

	"github.com/aws/aws-sdk-go-v2/service/shield"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type ShieldGenerator struct {
	AWSService
}

// InitResources enumerates Shield Advanced protections and protection groups.
// Shield is a partition-global service (endpoint signed for us-east-1), imported
// in the aws-global pass. Import IDs are the protection / protection-group id.
func (g *ShieldGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := shield.NewFromConfig(config)
	ctx := context.TODO()

	protections := shield.NewListProtectionsPaginator(svc, &shield.ListProtectionsInput{})
	for protections.HasMorePages() {
		page, err := protections.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, p := range page.Protections {
			id := StringValue(p.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_shield_protection", "aws", defaultAllowEmptyValues))
		}
	}

	groups := shield.NewListProtectionGroupsPaginator(svc, &shield.ListProtectionGroupsInput{})
	for groups.HasMorePages() {
		page, err := groups.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, pg := range page.ProtectionGroups {
			id := StringValue(pg.ProtectionGroupId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_shield_protection_group", "aws", defaultAllowEmptyValues))
		}
	}

	return nil
}
