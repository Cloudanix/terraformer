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
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type ConnectGenerator struct {
	AWSService
}

// InitResources enumerates Connect instances and their per-instance children
// (queues, routing profiles, contact flows, hours of operation, security
// profiles, users). Child import IDs are "<instance-id>:<resource-id>".
func (g *ConnectGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := connect.NewFromConfig(config)
	ctx := context.TODO()

	var instanceIDs []string
	p := connect.NewListInstancesPaginator(svc, &connect.ListInstancesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, instance := range page.InstanceSummaryList {
			id := StringValue(instance.Id)
			if id == "" {
				continue
			}
			instanceIDs = append(instanceIDs, id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(instance.InstanceAlias), "aws_connect_instance", "aws", defaultAllowEmptyValues))
		}
	}

	for _, instanceID := range instanceIDs {
		g.loadConnectChildren(svc, instanceID)
	}
	return nil
}

func (g *ConnectGenerator) loadConnectChildren(svc *connect.Client, instanceID string) {
	ctx := context.TODO()
	add := func(childID, tfType string) {
		if childID != "" {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				fmt.Sprintf("%s:%s", instanceID, childID),
				fmt.Sprintf("%s_%s", instanceID, childID),
				tfType, "aws", defaultAllowEmptyValues))
		}
	}
	for q := connect.NewListQueuesPaginator(svc, &connect.ListQueuesInput{InstanceId: aws.String(instanceID)}); q.HasMorePages(); {
		page, err := q.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range page.QueueSummaryList {
			add(StringValue(x.Id), "aws_connect_queue")
		}
	}
	for r := connect.NewListRoutingProfilesPaginator(svc, &connect.ListRoutingProfilesInput{InstanceId: aws.String(instanceID)}); r.HasMorePages(); {
		page, err := r.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range page.RoutingProfileSummaryList {
			add(StringValue(x.Id), "aws_connect_routing_profile")
		}
	}
	for c := connect.NewListContactFlowsPaginator(svc, &connect.ListContactFlowsInput{InstanceId: aws.String(instanceID)}); c.HasMorePages(); {
		page, err := c.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range page.ContactFlowSummaryList {
			add(StringValue(x.Id), "aws_connect_contact_flow")
		}
	}
	for h := connect.NewListHoursOfOperationsPaginator(svc, &connect.ListHoursOfOperationsInput{InstanceId: aws.String(instanceID)}); h.HasMorePages(); {
		page, err := h.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range page.HoursOfOperationSummaryList {
			add(StringValue(x.Id), "aws_connect_hours_of_operation")
		}
	}
	for s := connect.NewListSecurityProfilesPaginator(svc, &connect.ListSecurityProfilesInput{InstanceId: aws.String(instanceID)}); s.HasMorePages(); {
		page, err := s.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range page.SecurityProfileSummaryList {
			add(StringValue(x.Id), "aws_connect_security_profile")
		}
	}
	for u := connect.NewListUsersPaginator(svc, &connect.ListUsersInput{InstanceId: aws.String(instanceID)}); u.HasMorePages(); {
		page, err := u.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range page.UserSummaryList {
			add(StringValue(x.Id), "aws_connect_user")
		}
	}
	for q := connect.NewListQuickConnectsPaginator(svc, &connect.ListQuickConnectsInput{InstanceId: aws.String(instanceID)}); q.HasMorePages(); {
		page, err := q.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range page.QuickConnectSummaryList {
			add(StringValue(x.Id), "aws_connect_quick_connect")
		}
	}
	for m := connect.NewListContactFlowModulesPaginator(svc, &connect.ListContactFlowModulesInput{InstanceId: aws.String(instanceID)}); m.HasMorePages(); {
		page, err := m.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range page.ContactFlowModulesSummaryList {
			add(StringValue(x.Id), "aws_connect_contact_flow_module")
		}
	}
	for h := connect.NewListUserHierarchyGroupsPaginator(svc, &connect.ListUserHierarchyGroupsInput{InstanceId: aws.String(instanceID)}); h.HasMorePages(); {
		page, err := h.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range page.UserHierarchyGroupSummaryList {
			add(StringValue(x.Id), "aws_connect_user_hierarchy_group")
		}
	}
}
