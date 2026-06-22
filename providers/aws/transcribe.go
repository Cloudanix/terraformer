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
		page, err := p.NextPage(context.TODO())
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
	return nil
}
