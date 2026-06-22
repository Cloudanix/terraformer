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

	"github.com/aws/aws-sdk-go-v2/service/macie2"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type Macie2Generator struct {
	AWSService
}

// InitResources enumerates Macie classification jobs, custom data identifiers,
// findings filters, and members. Import IDs are the resource's own id (account
// id for members).
func (g *Macie2Generator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := macie2.NewFromConfig(config)
	ctx := context.TODO()

	jobs := macie2.NewListClassificationJobsPaginator(svc, &macie2.ListClassificationJobsInput{})
	for jobs.HasMorePages() {
		page, err := jobs.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, job := range page.Items {
			id := StringValue(job.JobId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_macie2_classification_job", "aws", defaultAllowEmptyValues))
		}
	}

	identifiers := macie2.NewListCustomDataIdentifiersPaginator(svc, &macie2.ListCustomDataIdentifiersInput{})
	for identifiers.HasMorePages() {
		page, err := identifiers.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, cdi := range page.Items {
			id := StringValue(cdi.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_macie2_custom_data_identifier", "aws", defaultAllowEmptyValues))
		}
	}

	filters := macie2.NewListFindingsFiltersPaginator(svc, &macie2.ListFindingsFiltersInput{})
	for filters.HasMorePages() {
		page, err := filters.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, ff := range page.FindingsFilterListItems {
			id := StringValue(ff.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_macie2_findings_filter", "aws", defaultAllowEmptyValues))
		}
	}

	members := macie2.NewListMembersPaginator(svc, &macie2.ListMembersInput{})
	for members.HasMorePages() {
		page, err := members.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, member := range page.Members {
			id := StringValue(member.AccountId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_macie2_member", "aws", defaultAllowEmptyValues))
		}
	}

	return nil
}
