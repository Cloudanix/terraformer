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

	"github.com/aws/aws-sdk-go-v2/service/workspacesweb"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type WorkSpacesWebGenerator struct {
	AWSService
}

// InitResources enumerates WorkSpaces Web portals. Import ID is the portal ARN.
func (g *WorkSpacesWebGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := workspacesweb.NewFromConfig(config)
	ctx := context.TODO()
	add := func(arn, tfType string) {
		if arn != "" {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(arn, arn, tfType, "aws", defaultAllowEmptyValues))
		}
	}

	for p := workspacesweb.NewListPortalsPaginator(svc, &workspacesweb.ListPortalsInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, portal := range page.Portals {
			arn := StringValue(portal.PortalArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(portal.DisplayName), "aws_workspacesweb_portal", "aws", defaultAllowEmptyValues))
			for ip := workspacesweb.NewListIdentityProvidersPaginator(svc, &workspacesweb.ListIdentityProvidersInput{PortalArn: &arn}); ip.HasMorePages(); {
				ipage, err := ip.NextPage(ctx)
				if err != nil {
					break
				}
				for _, idp := range ipage.IdentityProviders {
					add(StringValue(idp.IdentityProviderArn), "aws_workspacesweb_identity_provider")
				}
			}
		}
	}
	for p := workspacesweb.NewListBrowserSettingsPaginator(svc, &workspacesweb.ListBrowserSettingsInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range page.BrowserSettings {
			add(StringValue(x.BrowserSettingsArn), "aws_workspacesweb_browser_settings")
		}
	}
	for p := workspacesweb.NewListNetworkSettingsPaginator(svc, &workspacesweb.ListNetworkSettingsInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range page.NetworkSettings {
			add(StringValue(x.NetworkSettingsArn), "aws_workspacesweb_network_settings")
		}
	}
	for p := workspacesweb.NewListIpAccessSettingsPaginator(svc, &workspacesweb.ListIpAccessSettingsInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range page.IpAccessSettings {
			add(StringValue(x.IpAccessSettingsArn), "aws_workspacesweb_ip_access_settings")
		}
	}
	for p := workspacesweb.NewListUserSettingsPaginator(svc, &workspacesweb.ListUserSettingsInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range page.UserSettings {
			add(StringValue(x.UserSettingsArn), "aws_workspacesweb_user_settings")
		}
	}
	for p := workspacesweb.NewListDataProtectionSettingsPaginator(svc, &workspacesweb.ListDataProtectionSettingsInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range page.DataProtectionSettings {
			add(StringValue(x.DataProtectionSettingsArn), "aws_workspacesweb_data_protection_settings")
		}
	}
	for p := workspacesweb.NewListUserAccessLoggingSettingsPaginator(svc, &workspacesweb.ListUserAccessLoggingSettingsInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range page.UserAccessLoggingSettings {
			add(StringValue(x.UserAccessLoggingSettingsArn), "aws_workspacesweb_user_access_logging_settings")
		}
	}
	for p := workspacesweb.NewListTrustStoresPaginator(svc, &workspacesweb.ListTrustStoresInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range page.TrustStores {
			add(StringValue(x.TrustStoreArn), "aws_workspacesweb_trust_store")
		}
	}
	for p := workspacesweb.NewListSessionLoggersPaginator(svc, &workspacesweb.ListSessionLoggersInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range page.SessionLoggers {
			add(StringValue(x.SessionLoggerArn), "aws_workspacesweb_session_logger")
		}
	}
	return nil
}
