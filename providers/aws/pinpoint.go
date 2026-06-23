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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pinpoint"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type PinpointGenerator struct {
	AWSService
}

// InitResources enumerates Pinpoint apps. GetApps is token-paginated manually
// (no v2 paginator). Import ID is the application id.
func (g *PinpointGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := pinpoint.NewFromConfig(config)

	var appIDs []string
	var token *string
	for {
		out, err := svc.GetApps(awsContext(), &pinpoint.GetAppsInput{Token: token})
		if err != nil {
			return err
		}
		if out.ApplicationsResponse == nil {
			break
		}
		for _, app := range out.ApplicationsResponse.Item {
			id := StringValue(app.Id)
			if id == "" {
				continue
			}
			appIDs = append(appIDs, id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(app.Name), "aws_pinpoint_app", "aws", defaultAllowEmptyValues))
		}
		if out.ApplicationsResponse.NextToken == nil {
			break
		}
		token = aws.String(StringValue(out.ApplicationsResponse.NextToken))
	}

	for _, appID := range appIDs {
		g.loadPinpointChannels(svc, appID)
	}
	return nil
}

// loadPinpointChannels probes each channel and event stream for an app. Every
// channel is a singleton; its Terraform import ID is the application id.
func (g *PinpointGenerator) loadPinpointChannels(svc *pinpoint.Client, appID string) {
	ctx := awsContext()
	add := func(tfType string) {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			appID, appID, tfType, "aws", defaultAllowEmptyValues))
	}
	if _, err := svc.GetAdmChannel(ctx, &pinpoint.GetAdmChannelInput{ApplicationId: &appID}); err == nil {
		add("aws_pinpoint_adm_channel")
	}
	if _, err := svc.GetApnsChannel(ctx, &pinpoint.GetApnsChannelInput{ApplicationId: &appID}); err == nil {
		add("aws_pinpoint_apns_channel")
	}
	if _, err := svc.GetApnsSandboxChannel(ctx, &pinpoint.GetApnsSandboxChannelInput{ApplicationId: &appID}); err == nil {
		add("aws_pinpoint_apns_sandbox_channel")
	}
	if _, err := svc.GetApnsVoipChannel(ctx, &pinpoint.GetApnsVoipChannelInput{ApplicationId: &appID}); err == nil {
		add("aws_pinpoint_apns_voip_channel")
	}
	if _, err := svc.GetApnsVoipSandboxChannel(ctx, &pinpoint.GetApnsVoipSandboxChannelInput{ApplicationId: &appID}); err == nil {
		add("aws_pinpoint_apns_voip_sandbox_channel")
	}
	if _, err := svc.GetBaiduChannel(ctx, &pinpoint.GetBaiduChannelInput{ApplicationId: &appID}); err == nil {
		add("aws_pinpoint_baidu_channel")
	}
	if _, err := svc.GetEmailChannel(ctx, &pinpoint.GetEmailChannelInput{ApplicationId: &appID}); err == nil {
		add("aws_pinpoint_email_channel")
	}
	if _, err := svc.GetGcmChannel(ctx, &pinpoint.GetGcmChannelInput{ApplicationId: &appID}); err == nil {
		add("aws_pinpoint_gcm_channel")
	}
	if _, err := svc.GetSmsChannel(ctx, &pinpoint.GetSmsChannelInput{ApplicationId: &appID}); err == nil {
		add("aws_pinpoint_sms_channel")
	}
	if _, err := svc.GetEventStream(ctx, &pinpoint.GetEventStreamInput{ApplicationId: &appID}); err == nil {
		add("aws_pinpoint_event_stream")
	}
}
