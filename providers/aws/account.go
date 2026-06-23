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

	"github.com/aws/aws-sdk-go-v2/service/account"
	accounttypes "github.com/aws/aws-sdk-go-v2/service/account/types"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type AccountGenerator struct {
	AWSService
}

// InitResources enumerates AWS Account-level settings: opt-in Regions that have
// been explicitly enabled, the per-type alternate contacts, and the primary
// contact (an account singleton).
func (g *AccountGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := account.NewFromConfig(config)
	ctx := context.TODO()

	// Opt-in Regions explicitly enabled (default-on Regions aren't managed).
	for p := account.NewListRegionsPaginator(svc, &account.ListRegionsInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, r := range page.Regions {
			name := StringValue(r.RegionName)
			if name == "" || r.RegionOptStatus != accounttypes.RegionOptStatusEnabled {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_account_region", "aws", defaultAllowEmptyValues))
		}
	}

	// Alternate contacts, one per type (import ID is the contact type).
	for _, ct := range accounttypes.AlternateContactType("").Values() {
		if _, err := svc.GetAlternateContact(ctx, &account.GetAlternateContactInput{AlternateContactType: ct}); err == nil {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				string(ct), string(ct), "aws_account_alternate_contact", "aws", defaultAllowEmptyValues))
		}
	}

	// Primary contact (singleton; import ID is the account id).
	if _, err := svc.GetContactInformation(ctx, &account.GetContactInformationInput{}); err == nil {
		if acct, err := g.getAccountNumber(config); err == nil {
			id := StringValue(acct)
			if id != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, id, "aws_account_primary_contact", "aws", defaultAllowEmptyValues))
			}
		}
	}
	return nil
}
