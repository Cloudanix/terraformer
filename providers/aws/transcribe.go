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
	"github.com/aws/aws-sdk-go-v2/service/transcribe"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type TranscribeGenerator struct {
	AWSService
}

// InitResources enumerates Transcribe custom vocabularies. Import ID is the name.
func (g *TranscribeGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := transcribe.NewFromConfig(config)
	p := transcribe.NewListVocabulariesPaginator(svc, &transcribe.ListVocabulariesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, v := range page.Vocabularies {
			name := StringValue(v.VocabularyName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_transcribe_vocabulary", "aws", defaultAllowEmptyValues))
		}
	}

	ctx := awsContext()
	for lm := transcribe.NewListLanguageModelsPaginator(svc, &transcribe.ListLanguageModelsInput{}); lm.HasMorePages(); {
		page, err := lm.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, m := range page.Models {
			name := StringValue(m.ModelName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_transcribe_language_model", "aws", defaultAllowEmptyValues))
		}
	}
	for mv := transcribe.NewListMedicalVocabulariesPaginator(svc, &transcribe.ListMedicalVocabulariesInput{}); mv.HasMorePages(); {
		page, err := mv.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, v := range page.Vocabularies {
			name := StringValue(v.VocabularyName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_transcribe_medical_vocabulary", "aws", defaultAllowEmptyValues))
		}
	}
	for vf := transcribe.NewListVocabularyFiltersPaginator(svc, &transcribe.ListVocabularyFiltersInput{}); vf.HasMorePages(); {
		page, err := vf.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, f := range page.VocabularyFilters {
			name := StringValue(f.VocabularyFilterName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_transcribe_vocabulary_filter", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
