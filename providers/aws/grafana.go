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

	"github.com/aws/aws-sdk-go-v2/service/grafana"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type GrafanaGenerator struct {
	AWSService
}

// InitResources enumerates Managed Grafana workspaces. Import ID is the workspace id.
func (g *GrafanaGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := grafana.NewFromConfig(config)

	p := grafana.NewListWorkspacesPaginator(svc, &grafana.ListWorkspacesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, ws := range page.Workspaces {
			id := StringValue(ws.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(ws.Name), "aws_grafana_workspace", "aws", defaultAllowEmptyValues))
			if ws.LicenseType != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, id, "aws_grafana_license_association", "aws", defaultAllowEmptyValues))
			}
			g.loadWorkspaceChildren(svc, id)
		}
	}
	return nil
}

// loadWorkspaceChildren enumerates a workspace's SAML config and service accounts.
func (g *GrafanaGenerator) loadWorkspaceChildren(svc *grafana.Client, workspaceID string) {
	ctx := context.TODO()
	if auth, err := svc.DescribeWorkspaceAuthentication(ctx, &grafana.DescribeWorkspaceAuthenticationInput{WorkspaceId: &workspaceID}); err == nil &&
		auth.Authentication != nil && auth.Authentication.Saml != nil {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			workspaceID, workspaceID, "aws_grafana_workspace_saml_configuration", "aws", defaultAllowEmptyValues))
	}
	seenRole := map[string]bool{}
	for pp := grafana.NewListPermissionsPaginator(svc, &grafana.ListPermissionsInput{WorkspaceId: &workspaceID}); pp.HasMorePages(); {
		page, err := pp.NextPage(ctx)
		if err != nil {
			break
		}
		for _, perm := range page.Permissions {
			role := string(perm.Role)
			if role == "" || seenRole[role] {
				continue
			}
			seenRole[role] = true
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				role+"/"+workspaceID, role+"_"+workspaceID, "aws_grafana_role_association", "aws", defaultAllowEmptyValues))
		}
	}
	for p := grafana.NewListWorkspaceServiceAccountsPaginator(svc, &grafana.ListWorkspaceServiceAccountsInput{WorkspaceId: &workspaceID}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, sa := range page.ServiceAccounts {
			said := StringValue(sa.Id)
			if said == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				workspaceID+"/"+said, workspaceID+"_"+said, "aws_grafana_workspace_service_account", "aws", defaultAllowEmptyValues))
			for tp := grafana.NewListWorkspaceServiceAccountTokensPaginator(svc, &grafana.ListWorkspaceServiceAccountTokensInput{WorkspaceId: &workspaceID, ServiceAccountId: sa.Id}); tp.HasMorePages(); {
				tpage, err := tp.NextPage(ctx)
				if err != nil {
					break
				}
				for _, t := range tpage.ServiceAccountTokens {
					tid := StringValue(t.Id)
					if tid == "" {
						continue
					}
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						workspaceID+"/"+said+"/"+tid, workspaceID+"_"+said+"_"+tid, "aws_grafana_workspace_service_account_token", "aws", defaultAllowEmptyValues))
				}
			}
		}
	}
}
