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

	"github.com/aws/aws-sdk-go-v2/service/chimesdkvoice"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type ChimeSDKVoiceGenerator struct {
	AWSService
}

// InitResources enumerates Chime SDK Voice connectors. Import ID is the id.
func (g *ChimeSDKVoiceGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := chimesdkvoice.NewFromConfig(config)

	ctx := context.TODO()
	add := func(id, name, tfType string) {
		if id != "" {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, name, tfType, "aws", defaultAllowEmptyValues))
		}
	}

	p := chimesdkvoice.NewListVoiceConnectorsPaginator(svc, &chimesdkvoice.ListVoiceConnectorsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, vc := range page.VoiceConnectors {
			id := StringValue(vc.VoiceConnectorId)
			if id == "" {
				continue
			}
			name := StringValue(vc.Name)
			add(id, name, "aws_chimesdkvoice_voice_connector")
			add(id, name, "aws_chime_voice_connector")

			if o, err := svc.GetVoiceConnectorOrigination(ctx, &chimesdkvoice.GetVoiceConnectorOriginationInput{VoiceConnectorId: vc.VoiceConnectorId}); err == nil && o.Origination != nil {
				add(id, name, "aws_chime_voice_connector_origination")
			}
			if s, err := svc.GetVoiceConnectorStreamingConfiguration(ctx, &chimesdkvoice.GetVoiceConnectorStreamingConfigurationInput{VoiceConnectorId: vc.VoiceConnectorId}); err == nil && s.StreamingConfiguration != nil {
				add(id, name, "aws_chime_voice_connector_streaming")
			}
			if t, err := svc.GetVoiceConnectorTermination(ctx, &chimesdkvoice.GetVoiceConnectorTerminationInput{VoiceConnectorId: vc.VoiceConnectorId}); err == nil && t.Termination != nil {
				add(id, name, "aws_chime_voice_connector_termination")
			}
			if c, err := svc.ListVoiceConnectorTerminationCredentials(ctx, &chimesdkvoice.ListVoiceConnectorTerminationCredentialsInput{VoiceConnectorId: vc.VoiceConnectorId}); err == nil && len(c.Usernames) > 0 {
				add(id, name, "aws_chime_voice_connector_termination_credentials")
			}
		}
	}

	for gp := chimesdkvoice.NewListVoiceConnectorGroupsPaginator(svc, &chimesdkvoice.ListVoiceConnectorGroupsInput{}); gp.HasMorePages(); {
		page, err := gp.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, grp := range page.VoiceConnectorGroups {
			add(StringValue(grp.VoiceConnectorGroupId), StringValue(grp.Name), "aws_chime_voice_connector_group")
		}
	}
	return nil
}
