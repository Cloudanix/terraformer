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
	"github.com/aws/aws-sdk-go-v2/service/kendra"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type KendraGenerator struct {
	AWSService
}

// InitResources enumerates Kendra indices. Import ID is the index id.
func (g *KendraGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := kendra.NewFromConfig(config)

	ctx := awsContext()
	var indexIDs []string
	p := kendra.NewListIndicesPaginator(svc, &kendra.ListIndicesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, idx := range page.IndexConfigurationSummaryItems {
			id := StringValue(idx.Id)
			if id == "" {
				continue
			}
			indexIDs = append(indexIDs, id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(idx.Name), "aws_kendra_index", "aws", defaultAllowEmptyValues))
		}
	}

	for _, indexID := range indexIDs {
		idx := indexID
		add := func(childID, tfType string) {
			if childID != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					childID+"/"+idx, childID+"_"+idx, tfType, "aws", defaultAllowEmptyValues))
			}
		}
		for c := kendra.NewListDataSourcesPaginator(svc, &kendra.ListDataSourcesInput{IndexId: &idx}); c.HasMorePages(); {
			page, err := c.NextPage(ctx)
			if err != nil {
				break
			}
			for _, x := range page.SummaryItems {
				add(StringValue(x.Id), "aws_kendra_data_source")
			}
		}
		for c := kendra.NewListExperiencesPaginator(svc, &kendra.ListExperiencesInput{IndexId: &idx}); c.HasMorePages(); {
			page, err := c.NextPage(ctx)
			if err != nil {
				break
			}
			for _, x := range page.SummaryItems {
				add(StringValue(x.Id), "aws_kendra_experience")
			}
		}
		for c := kendra.NewListFaqsPaginator(svc, &kendra.ListFaqsInput{IndexId: &idx}); c.HasMorePages(); {
			page, err := c.NextPage(ctx)
			if err != nil {
				break
			}
			for _, x := range page.FaqSummaryItems {
				add(StringValue(x.Id), "aws_kendra_faq")
			}
		}
		for c := kendra.NewListQuerySuggestionsBlockListsPaginator(svc, &kendra.ListQuerySuggestionsBlockListsInput{IndexId: &idx}); c.HasMorePages(); {
			page, err := c.NextPage(ctx)
			if err != nil {
				break
			}
			for _, x := range page.BlockListSummaryItems {
				add(StringValue(x.Id), "aws_kendra_query_suggestions_block_list")
			}
		}
		for c := kendra.NewListThesauriPaginator(svc, &kendra.ListThesauriInput{IndexId: &idx}); c.HasMorePages(); {
			page, err := c.NextPage(ctx)
			if err != nil {
				break
			}
			for _, x := range page.ThesaurusSummaryItems {
				add(StringValue(x.Id), "aws_kendra_thesaurus")
			}
		}
	}
	return nil
}
