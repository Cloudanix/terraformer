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
	"github.com/aws/aws-sdk-go-v2/service/securitylake"
	securitylaketypes "github.com/aws/aws-sdk-go-v2/service/securitylake/types"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type SecurityLakeGenerator struct {
	AWSService
}

// InitResources enumerates Security Lake data lakes and subscribers. Import IDs
// are the resource ARN.
func (g *SecurityLakeGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := securitylake.NewFromConfig(config)
	ctx := awsContext()

	lakes, err := svc.ListDataLakes(ctx, &securitylake.ListDataLakesInput{})
	if err != nil {
		return err
	}
	for _, lake := range lakes.DataLakes {
		arn := StringValue(lake.DataLakeArn)
		if arn == "" {
			continue
		}
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			arn, arn, "aws_securitylake_data_lake", "aws", defaultAllowEmptyValues))
	}

	subscribers := securitylake.NewListSubscribersPaginator(svc, &securitylake.ListSubscribersInput{})
	for subscribers.HasMorePages() {
		page, err := subscribers.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, sub := range page.Subscribers {
			arn := StringValue(sub.SubscriberArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_securitylake_subscriber", "aws", defaultAllowEmptyValues))
		}
	}

	// Log sources, deduped by source name (same source repeats per account/region).
	seenLog := map[string]bool{}
	for lp := securitylake.NewListLogSourcesPaginator(svc, &securitylake.ListLogSourcesInput{}); lp.HasMorePages(); {
		page, err := lp.NextPage(ctx)
		if err != nil {
			break
		}
		for _, src := range page.Sources {
			for _, r := range src.Sources {
				tfType, name, ok := securityLakeLogSource(r, seenLog)
				if !ok {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					name, name, tfType, "aws", defaultAllowEmptyValues))
			}
		}
	}
	return nil
}

// securityLakeLogSource classifies one log-source union member into its TF
// resource type + source name, deduping against seen (keyed by type+name so the
// same source repeated across accounts/regions is emitted once). Returns
// ok=false for empty names, already-seen sources, or unknown union members.
func securityLakeLogSource(r securitylaketypes.LogSourceResource, seen map[string]bool) (tfType, name string, ok bool) {
	switch v := r.(type) {
	case *securitylaketypes.LogSourceResourceMemberAwsLogSource:
		name, tfType = string(v.Value.SourceName), "aws_securitylake_aws_log_source"
	case *securitylaketypes.LogSourceResourceMemberCustomLogSource:
		name, tfType = StringValue(v.Value.SourceName), "aws_securitylake_custom_log_source"
	default:
		return "", "", false
	}
	if name == "" {
		return "", "", false
	}
	key := tfType + ":" + name
	if seen[key] {
		return "", "", false
	}
	seen[key] = true
	return tfType, name, true
}
