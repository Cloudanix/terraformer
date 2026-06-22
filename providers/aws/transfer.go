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

	p := transfer.NewListServersPaginator(svc, &transfer.ListServersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.Servers, "aws_transfer_server",
			defaultAllowEmptyValues,
			func(s types.ListedServer) string { return StringValue(s.ServerId) },
			func(s types.ListedServer) string { return StringValue(s.ServerId) })
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
