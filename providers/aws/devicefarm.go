// Copyright 2020 The Terraformer Authors.
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
	"github.com/aws/aws-sdk-go-v2/service/devicefarm"
)

var devicefarmAllowEmptyValues = []string{"tags."}

type DeviceFarmGenerator struct {
	AWSService
}

func (g *DeviceFarmGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := devicefarm.NewFromConfig(config)
	ctx := context.TODO()
	p := devicefarm.NewListProjectsPaginator(svc, &devicefarm.ListProjectsInput{})
	var resources []terraformutils.Resource
	var projectArns []string
	for p.HasMorePages() {
		page, e := p.NextPage(ctx)
		if e != nil {
			return e
		}
		for _, project := range page.Projects {
			projectArn := StringValue(project.Arn)
			projectName := StringValue(project.Name)
			projectArns = append(projectArns, projectArn)
			resources = append(resources, terraformutils.NewSimpleResource(
				projectArn,
				projectName,
				"aws_devicefarm_project",
				"aws",
				devicefarmAllowEmptyValues))
		}
	}

	for _, projectArn := range projectArns {
		arn := projectArn
		for dp := devicefarm.NewListDevicePoolsPaginator(svc, &devicefarm.ListDevicePoolsInput{Arn: &arn}); dp.HasMorePages(); {
			page, err := dp.NextPage(ctx)
			if err != nil {
				break
			}
			for _, x := range page.DevicePools {
				if a := StringValue(x.Arn); a != "" {
					resources = append(resources, terraformutils.NewSimpleResource(
						a, StringValue(x.Name), "aws_devicefarm_device_pool", "aws", devicefarmAllowEmptyValues))
				}
			}
		}
		for up := devicefarm.NewListUploadsPaginator(svc, &devicefarm.ListUploadsInput{Arn: &arn}); up.HasMorePages(); {
			page, err := up.NextPage(ctx)
			if err != nil {
				break
			}
			for _, x := range page.Uploads {
				if a := StringValue(x.Arn); a != "" {
					resources = append(resources, terraformutils.NewSimpleResource(
						a, StringValue(x.Name), "aws_devicefarm_upload", "aws", devicefarmAllowEmptyValues))
				}
			}
		}
		if np, err := svc.ListNetworkProfiles(ctx, &devicefarm.ListNetworkProfilesInput{Arn: &arn}); err == nil {
			for _, x := range np.NetworkProfiles {
				if a := StringValue(x.Arn); a != "" {
					resources = append(resources, terraformutils.NewSimpleResource(
						a, StringValue(x.Name), "aws_devicefarm_network_profile", "aws", devicefarmAllowEmptyValues))
				}
			}
		}
	}
	if profiles, err := svc.ListInstanceProfiles(context.TODO(), &devicefarm.ListInstanceProfilesInput{}); err == nil {
		for _, ip := range profiles.InstanceProfiles {
			arn := StringValue(ip.Arn)
			if arn == "" {
				continue
			}
			resources = append(resources, terraformutils.NewSimpleResource(
				arn, StringValue(ip.Name), "aws_devicefarm_instance_profile", "aws", devicefarmAllowEmptyValues))
		}
	}
	tgp := devicefarm.NewListTestGridProjectsPaginator(svc, &devicefarm.ListTestGridProjectsInput{})
	for tgp.HasMorePages() {
		page, err := tgp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, t := range page.TestGridProjects {
			arn := StringValue(t.Arn)
			if arn == "" {
				continue
			}
			resources = append(resources, terraformutils.NewSimpleResource(
				arn, StringValue(t.Name), "aws_devicefarm_test_grid_project", "aws", devicefarmAllowEmptyValues))
		}
	}
	g.Resources = resources
	return nil
}
