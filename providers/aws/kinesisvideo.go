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

	"github.com/aws/aws-sdk-go-v2/service/kinesisvideo"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type KinesisVideoGenerator struct {
	AWSService
}

// InitResources enumerates Kinesis Video streams. Import ID is the stream name.
func (g *KinesisVideoGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := kinesisvideo.NewFromConfig(config)

	p := kinesisvideo.NewListStreamsPaginator(svc, &kinesisvideo.ListStreamsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, stream := range page.StreamInfoList {
			name := StringValue(stream.StreamName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_kinesis_video_stream", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
