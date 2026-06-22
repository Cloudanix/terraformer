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
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/guardduty"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type GuardDutyGenerator struct {
	AWSService
}

// InitResources enumerates GuardDuty detectors and their children. Import IDs:
//   - aws_guardduty_detector              → detector ID
//   - everything else                     → "<detector-id>:<child-id>"
func (g *GuardDutyGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := guardduty.NewFromConfig(config)
	ctx := context.TODO()

	var detectorIDs []string
	detectors := guardduty.NewListDetectorsPaginator(svc, &guardduty.ListDetectorsInput{})
	for detectors.HasMorePages() {
		page, err := detectors.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, id := range page.DetectorIds {
			if id == "" {
				continue
			}
			detectorIDs = append(detectorIDs, id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_guardduty_detector", "aws", defaultAllowEmptyValues))
		}
	}

	for _, detectorID := range detectorIDs {
		if err := g.addDetectorChildren(svc, detectorID); err != nil {
			return err
		}
	}
	return nil
}

func (g *GuardDutyGenerator) addDetectorChildren(svc *guardduty.Client, detectorID string) error {
	ctx := context.TODO()
	child := func(childID, tfType string) {
		if childID == "" {
			return
		}
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			fmt.Sprintf("%s:%s", detectorID, childID),
			fmt.Sprintf("%s_%s", detectorID, childID),
			tfType, "aws", defaultAllowEmptyValues))
	}

	filters := guardduty.NewListFiltersPaginator(svc, &guardduty.ListFiltersInput{DetectorId: aws.String(detectorID)})
	for filters.HasMorePages() {
		page, err := filters.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, name := range page.FilterNames {
			child(name, "aws_guardduty_filter")
		}
	}

	ipSets := guardduty.NewListIPSetsPaginator(svc, &guardduty.ListIPSetsInput{DetectorId: aws.String(detectorID)})
	for ipSets.HasMorePages() {
		page, err := ipSets.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, id := range page.IpSetIds {
			child(id, "aws_guardduty_ipset")
		}
	}

	threatIntelSets := guardduty.NewListThreatIntelSetsPaginator(svc, &guardduty.ListThreatIntelSetsInput{DetectorId: aws.String(detectorID)})
	for threatIntelSets.HasMorePages() {
		page, err := threatIntelSets.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, id := range page.ThreatIntelSetIds {
			child(id, "aws_guardduty_threatintelset")
		}
	}

	members := guardduty.NewListMembersPaginator(svc, &guardduty.ListMembersInput{DetectorId: aws.String(detectorID)})
	for members.HasMorePages() {
		page, err := members.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, member := range page.Members {
			child(StringValue(member.AccountId), "aws_guardduty_member")
		}
	}

	destinations := guardduty.NewListPublishingDestinationsPaginator(svc, &guardduty.ListPublishingDestinationsInput{DetectorId: aws.String(detectorID)})
	for destinations.HasMorePages() {
		page, err := destinations.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, dest := range page.Destinations {
			child(StringValue(dest.DestinationId), "aws_guardduty_publishing_destination")
		}
	}

	return nil
}
