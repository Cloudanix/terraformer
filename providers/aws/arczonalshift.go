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

	"github.com/aws/aws-sdk-go-v2/service/arczonalshift"
	arczonalshifttypes "github.com/aws/aws-sdk-go-v2/service/arczonalshift/types"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type ARCZonalShiftGenerator struct {
	AWSService
}

// InitResources enumerates resources with zonal autoshift enabled (import by
// resource identifier) plus the account observer-notification status singleton.
func (g *ARCZonalShiftGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := arczonalshift.NewFromConfig(config)
	ctx := context.TODO()
	for p := arczonalshift.NewListManagedResourcesPaginator(svc, &arczonalshift.ListManagedResourcesInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, r := range page.Items {
			if r.ZonalAutoshiftStatus != arczonalshifttypes.ZonalAutoshiftStatusEnabled {
				continue
			}
			id := StringValue(r.Arn)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(r.Name), "aws_arczonalshift_zonal_autoshift_configuration", "aws", defaultAllowEmptyValues))
		}
	}
	if _, err := svc.GetAutoshiftObserverNotificationStatus(ctx, &arczonalshift.GetAutoshiftObserverNotificationStatusInput{}); err == nil {
		if acct, err := g.getAccountNumber(config); err == nil {
			id := StringValue(acct)
			if id != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, id, "aws_arczonalshift_autoshift_observer_notification_status", "aws", defaultAllowEmptyValues))
			}
		}
	}
	return nil
}
