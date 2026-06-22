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

	"github.com/aws/aws-sdk-go-v2/service/appstream"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type AppStreamGenerator struct {
	AWSService
}

// InitResources enumerates AppStream fleets. Import ID is the fleet name.
func (g *AppStreamGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := appstream.NewFromConfig(config)

	var token *string
	for {
		out, err := svc.DescribeFleets(context.TODO(), &appstream.DescribeFleetsInput{NextToken: token})
		if err != nil {
			return err
		}
		for _, fleet := range out.Fleets {
			name := StringValue(fleet.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_appstream_fleet", "aws", defaultAllowEmptyValues))
			fleetName := name
			if assoc, err := svc.ListAssociatedStacks(context.TODO(), &appstream.ListAssociatedStacksInput{FleetName: &fleetName}); err == nil {
				for _, stackName := range assoc.Names {
					if stackName == "" {
						continue
					}
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						fleetName+"/"+stackName, fleetName+"_"+stackName, "aws_appstream_fleet_stack_association", "aws", defaultAllowEmptyValues))
				}
			}
		}
		if out.NextToken == nil {
			break
		}
		token = out.NextToken
	}

	if stacks, err := svc.DescribeStacks(context.TODO(), &appstream.DescribeStacksInput{}); err == nil {
		for _, s := range stacks.Stacks {
			name := StringValue(s.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_appstream_stack", "aws", defaultAllowEmptyValues))
		}
	}
	if builders, err := svc.DescribeImageBuilders(context.TODO(), &appstream.DescribeImageBuildersInput{}); err == nil {
		for _, b := range builders.ImageBuilders {
			name := StringValue(b.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_appstream_image_builder", "aws", defaultAllowEmptyValues))
		}
	}
	if dcs, err := svc.DescribeDirectoryConfigs(context.TODO(), &appstream.DescribeDirectoryConfigsInput{}); err == nil {
		for _, dc := range dcs.DirectoryConfigs {
			name := StringValue(dc.DirectoryName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_appstream_directory_config", "aws", defaultAllowEmptyValues))
		}
	}
	if users, err := svc.DescribeUsers(context.TODO(), &appstream.DescribeUsersInput{AuthenticationType: "USERPOOL"}); err == nil {
		for _, u := range users.Users {
			name := StringValue(u.UserName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name+"/USERPOOL", name, "aws_appstream_user", "aws", defaultAllowEmptyValues))
		}
	}
	if usa, err := svc.DescribeUserStackAssociations(context.TODO(), &appstream.DescribeUserStackAssociationsInput{}); err == nil {
		for _, a := range usa.UserStackAssociations {
			userName, stackName := StringValue(a.UserName), StringValue(a.StackName)
			if userName == "" || stackName == "" {
				continue
			}
			id := userName + "/" + string(a.AuthenticationType) + "/" + stackName
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, userName+"_"+stackName, "aws_appstream_user_stack_association", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
