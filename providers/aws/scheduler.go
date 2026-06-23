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
	"github.com/aws/aws-sdk-go-v2/service/scheduler"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type SchedulerGenerator struct {
	AWSService
}

// InitResources enumerates EventBridge Scheduler schedule groups and schedules.
// Import IDs: group name; "<group-name>/<schedule-name>" for schedules.
func (g *SchedulerGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := scheduler.NewFromConfig(config)
	ctx := awsContext()

	groups := scheduler.NewListScheduleGroupsPaginator(svc, &scheduler.ListScheduleGroupsInput{})
	for groups.HasMorePages() {
		page, err := groups.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, grp := range page.ScheduleGroups {
			name := StringValue(grp.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_scheduler_schedule_group", "aws", defaultAllowEmptyValues))
		}
	}

	schedules := scheduler.NewListSchedulesPaginator(svc, &scheduler.ListSchedulesInput{})
	for schedules.HasMorePages() {
		page, err := schedules.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, sch := range page.Schedules {
			name := StringValue(sch.Name)
			if name == "" {
				continue
			}
			group := StringValue(sch.GroupName)
			if group == "" {
				group = "default"
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				group+"/"+name, group+"_"+name, "aws_scheduler_schedule", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
