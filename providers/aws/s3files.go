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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3files"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type S3FilesGenerator struct {
	AWSService
}

// InitResources enumerates S3 file systems plus their access points, mount
// targets, policy, and synchronization configuration.
func (g *S3FilesGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := s3files.NewFromConfig(config)
	ctx := context.TODO()
	for p := s3files.NewListFileSystemsPaginator(svc, &s3files.ListFileSystemsInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, fs := range page.FileSystems {
			id := StringValue(fs.FileSystemId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_s3files_file_system", "aws", defaultAllowEmptyValues))
			if _, err := svc.GetFileSystemPolicy(ctx, &s3files.GetFileSystemPolicyInput{FileSystemId: aws.String(id)}); err == nil {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, id, "aws_s3files_file_system_policy", "aws", defaultAllowEmptyValues))
			}
			if _, err := svc.GetSynchronizationConfiguration(ctx, &s3files.GetSynchronizationConfigurationInput{FileSystemId: aws.String(id)}); err == nil {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, id, "aws_s3files_synchronization_configuration", "aws", defaultAllowEmptyValues))
			}
			for ap := s3files.NewListAccessPointsPaginator(svc, &s3files.ListAccessPointsInput{FileSystemId: aws.String(id)}); ap.HasMorePages(); {
				apage, err := ap.NextPage(ctx)
				if err != nil {
					break
				}
				for _, a := range apage.AccessPoints {
					arn := StringValue(a.AccessPointArn)
					if arn == "" {
						continue
					}
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						arn, arn, "aws_s3files_access_point", "aws", defaultAllowEmptyValues))
				}
			}
			for mp := s3files.NewListMountTargetsPaginator(svc, &s3files.ListMountTargetsInput{FileSystemId: aws.String(id)}); mp.HasMorePages(); {
				mpage, err := mp.NextPage(ctx)
				if err != nil {
					break
				}
				for _, m := range mpage.MountTargets {
					mid := StringValue(m.MountTargetId)
					if mid == "" {
						continue
					}
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						mid, mid, "aws_s3files_mount_target", "aws", defaultAllowEmptyValues))
				}
			}
		}
	}
	return nil
}
