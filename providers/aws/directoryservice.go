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
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/directoryservice"
	"github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
)

type DirectoryServiceGenerator struct {
	AWSService
}

// InitResources enumerates AWS Directory Service directories. The Terraform
// import ID for aws_directory_service_directory is the directory ID.
func (g *DirectoryServiceGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := directoryservice.NewFromConfig(config)

	ctx := awsContext()
	var directoryIDs []string
	p := directoryservice.NewDescribeDirectoriesPaginator(svc, &directoryservice.DescribeDirectoriesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, d := range page.DirectoryDescriptions {
			dirID := StringValue(d.DirectoryId)
			directoryIDs = append(directoryIDs, dirID)
			if d.RadiusSettings != nil && dirID != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					dirID, dirID, "aws_directory_service_radius_settings", "aws", defaultAllowEmptyValues))
			}
		}
		g.Resources = appendSimpleResources(g.Resources, page.DirectoryDescriptions, "aws_directory_service_directory",
			defaultAllowEmptyValues,
			func(d types.DirectoryDescription) string { return StringValue(d.DirectoryId) },
			func(d types.DirectoryDescription) string { return StringValue(d.DirectoryId) })
	}

	add := func(id, name, tfType string) {
		if id != "" {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, name, tfType, "aws", defaultAllowEmptyValues))
		}
	}
	for _, dirID := range directoryIDs {
		if dirID == "" {
			continue
		}
		if out, err := svc.DescribeConditionalForwarders(ctx, &directoryservice.DescribeConditionalForwardersInput{DirectoryId: &dirID}); err == nil {
			for _, cf := range out.ConditionalForwarders {
				domain := StringValue(cf.RemoteDomainName)
				if domain == "" {
					continue
				}
				add(dirID+":"+domain, dirID+"_"+domain, "aws_directory_service_conditional_forwarder")
			}
		}
		if out, err := svc.ListLogSubscriptions(ctx, &directoryservice.ListLogSubscriptionsInput{DirectoryId: &dirID}); err == nil && len(out.LogSubscriptions) > 0 {
			add(dirID, dirID, "aws_directory_service_log_subscription")
		}
		if out, err := svc.DescribeTrusts(ctx, &directoryservice.DescribeTrustsInput{DirectoryId: &dirID}); err == nil {
			for _, t := range out.Trusts {
				add(StringValue(t.TrustId), StringValue(t.TrustId), "aws_directory_service_trust")
			}
		}
		if out, err := svc.DescribeSharedDirectories(ctx, &directoryservice.DescribeSharedDirectoriesInput{OwnerDirectoryId: &dirID}); err == nil {
			for _, sd := range out.SharedDirectories {
				owner := StringValue(sd.OwnerDirectoryId)
				shared := StringValue(sd.SharedDirectoryId)
				if owner == "" || shared == "" {
					continue
				}
				add(owner+"/"+shared, owner+"_"+shared, "aws_directory_service_shared_directory")
			}
		}
		dir := dirID
		for rp := directoryservice.NewDescribeRegionsPaginator(svc, &directoryservice.DescribeRegionsInput{DirectoryId: &dir}); rp.HasMorePages(); {
			page, err := rp.NextPage(ctx)
			if err != nil {
				break
			}
			for _, r := range page.RegionsDescription {
				// Only additional (replicated) regions are separate TF resources.
				if r.RegionType != "Additional" {
					continue
				}
				region := StringValue(r.RegionName)
				if region == "" {
					continue
				}
				add(dir+","+region, dir+"_"+region, "aws_directory_service_region")
			}
		}
	}
	return nil
}
