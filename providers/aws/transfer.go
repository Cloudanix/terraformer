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

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/transfer"
	"github.com/aws/aws-sdk-go-v2/service/transfer/types"
)

type TransferGenerator struct {
	AWSService
}

// InitResources enumerates AWS Transfer Family servers. The Terraform import ID
// for aws_transfer_server is the server ID.
func (g *TransferGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := transfer.NewFromConfig(config)

	ctx := context.TODO()
	var serverIDs []string
	p := transfer.NewListServersPaginator(svc, &transfer.ListServersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, s := range page.Servers {
			serverIDs = append(serverIDs, StringValue(s.ServerId))
		}
		g.Resources = appendSimpleResources(g.Resources, page.Servers, "aws_transfer_server",
			defaultAllowEmptyValues,
			func(s types.ListedServer) string { return StringValue(s.ServerId) },
			func(s types.ListedServer) string { return StringValue(s.ServerId) })
	}

	for _, serverID := range serverIDs {
		if serverID == "" {
			continue
		}
		sid := serverID
		for up := transfer.NewListUsersPaginator(svc, &transfer.ListUsersInput{ServerId: &sid}); up.HasMorePages(); {
			page, err := up.NextPage(ctx)
			if err != nil {
				break
			}
			for _, u := range page.Users {
				name := StringValue(u.UserName)
				if name == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					sid+"/"+name, sid+"_"+name, "aws_transfer_user", "aws", defaultAllowEmptyValues))
				userName := name
				if du, err := svc.DescribeUser(ctx, &transfer.DescribeUserInput{ServerId: &sid, UserName: &userName}); err == nil && du.User != nil {
					for _, k := range du.User.SshPublicKeys {
						keyID := StringValue(k.SshPublicKeyId)
						if keyID == "" {
							continue
						}
						g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
							sid+"/"+userName+"/"+keyID, userName+"_"+keyID, "aws_transfer_ssh_key", "aws", defaultAllowEmptyValues))
					}
				}
			}
		}
		for ap := transfer.NewListAccessesPaginator(svc, &transfer.ListAccessesInput{ServerId: &sid}); ap.HasMorePages(); {
			page, err := ap.NextPage(ctx)
			if err != nil {
				break
			}
			for _, a := range page.Accesses {
				ext := StringValue(a.ExternalId)
				if ext == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					sid+"/"+ext, sid+"_"+ext, "aws_transfer_access", "aws", defaultAllowEmptyValues))
			}
		}
		for agp := transfer.NewListAgreementsPaginator(svc, &transfer.ListAgreementsInput{ServerId: &sid}); agp.HasMorePages(); {
			page, err := agp.NextPage(ctx)
			if err != nil {
				break
			}
			for _, ag := range page.Agreements {
				aid := StringValue(ag.AgreementId)
				if aid == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					sid+"/"+aid, sid+"_"+aid, "aws_transfer_agreement", "aws", defaultAllowEmptyValues))
			}
		}
	}

	for cp := transfer.NewListCertificatesPaginator(svc, &transfer.ListCertificatesInput{}); cp.HasMorePages(); {
		page, err := cp.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, c := range page.Certificates {
			id := StringValue(c.CertificateId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_transfer_certificate", "aws", defaultAllowEmptyValues))
		}
	}

	conns := transfer.NewListConnectorsPaginator(svc, &transfer.ListConnectorsInput{})
	for conns.HasMorePages() {
		page, err := conns.NextPage(context.TODO())
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.Connectors, "aws_transfer_connector",
			defaultAllowEmptyValues,
			func(c types.ListedConnector) string { return StringValue(c.ConnectorId) },
			func(c types.ListedConnector) string { return StringValue(c.ConnectorId) })
	}

	profiles := transfer.NewListProfilesPaginator(svc, &transfer.ListProfilesInput{})
	for profiles.HasMorePages() {
		page, err := profiles.NextPage(context.TODO())
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.Profiles, "aws_transfer_profile",
			defaultAllowEmptyValues,
			func(p types.ListedProfile) string { return StringValue(p.ProfileId) },
			func(p types.ListedProfile) string { return StringValue(p.ProfileId) })
	}

	workflows := transfer.NewListWorkflowsPaginator(svc, &transfer.ListWorkflowsInput{})
	for workflows.HasMorePages() {
		page, err := workflows.NextPage(context.TODO())
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.Workflows, "aws_transfer_workflow",
			defaultAllowEmptyValues,
			func(w types.ListedWorkflow) string { return StringValue(w.WorkflowId) },
			func(w types.ListedWorkflow) string { return StringValue(w.WorkflowId) })
	}
	return nil
}
