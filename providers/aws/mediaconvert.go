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
	"github.com/aws/aws-sdk-go-v2/service/mediaconvert"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type MediaConvertGenerator struct {
	AWSService
}

// InitResources enumerates MediaConvert queues. The AWS-managed "Default" queue
// is skipped. Import ID is the queue name. (Account-specific endpoints are no
// longer required — the regional endpoint is used directly.)
func (g *MediaConvertGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := mediaconvert.NewFromConfig(config)
	ctx := awsContext()

	p := mediaconvert.NewListQueuesPaginator(svc, &mediaconvert.ListQueuesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, queue := range page.Queues {
			name := StringValue(queue.Name)
			if name == "" || name == "Default" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_media_convert_queue", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
