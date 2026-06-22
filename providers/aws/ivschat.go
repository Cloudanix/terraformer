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

	"github.com/aws/aws-sdk-go-v2/service/ivschat"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type IVSChatGenerator struct {
	AWSService
}

func (g *IVSChatGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := ivschat.NewFromConfig(config)
	p := ivschat.NewListRoomsPaginator(svc, &ivschat.ListRoomsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, item := range page.Rooms {
			id := StringValue(item.Arn)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(item.Arn), "aws_ivschat_room", "aws", defaultAllowEmptyValues))
		}
	}

	for lc := ivschat.NewListLoggingConfigurationsPaginator(svc, &ivschat.ListLoggingConfigurationsInput{}); lc.HasMorePages(); {
		page, err := lc.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, item := range page.LoggingConfigurations {
			arn := StringValue(item.Arn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(item.Name), "aws_ivschat_logging_configuration", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
