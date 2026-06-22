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
	"github.com/aws/aws-sdk-go-v2/service/storagegateway"
	"github.com/aws/aws-sdk-go-v2/service/storagegateway/types"
)

type StorageGatewayGenerator struct {
	AWSService
}

// InitResources enumerates Storage Gateway gateways. The Terraform import ID for
// aws_storagegateway_gateway is the gateway ARN.
func (g *StorageGatewayGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := storagegateway.NewFromConfig(config)

	ctx := context.TODO()
	var gatewayArns []string
	p := storagegateway.NewListGatewaysPaginator(svc, &storagegateway.ListGatewaysInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, gw := range page.Gateways {
			gatewayArns = append(gatewayArns, StringValue(gw.GatewayARN))
		}
		g.Resources = appendSimpleResources(g.Resources, page.Gateways, "aws_storagegateway_gateway",
			defaultAllowEmptyValues,
			func(gw types.GatewayInfo) string { return StringValue(gw.GatewayARN) },
			func(gw types.GatewayInfo) string { return StringValue(gw.GatewayName) })
	}

	for _, gwArn := range gatewayArns {
		arn := gwArn
		if arn == "" {
			continue
		}
		// Per-gateway config singletons; import ID is the gateway ARN.
		if _, err := svc.DescribeCache(ctx, &storagegateway.DescribeCacheInput{GatewayARN: &arn}); err == nil {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_storagegateway_cache", "aws", defaultAllowEmptyValues))
		}
		if _, err := svc.DescribeUploadBuffer(ctx, &storagegateway.DescribeUploadBufferInput{GatewayARN: &arn}); err == nil {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_storagegateway_upload_buffer", "aws", defaultAllowEmptyValues))
		}
		if _, err := svc.DescribeWorkingStorage(ctx, &storagegateway.DescribeWorkingStorageInput{GatewayARN: &arn}); err == nil {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_storagegateway_working_storage", "aws", defaultAllowEmptyValues))
		}
	}

	for fsa := storagegateway.NewListFileSystemAssociationsPaginator(svc, &storagegateway.ListFileSystemAssociationsInput{}); fsa.HasMorePages(); {
		page, err := fsa.NextPage(ctx)
		if err != nil {
			break
		}
		for _, f := range page.FileSystemAssociationSummaryList {
			arn := StringValue(f.FileSystemAssociationARN)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_storagegateway_file_system_association", "aws", defaultAllowEmptyValues))
		}
	}

	pools := storagegateway.NewListTapePoolsPaginator(svc, &storagegateway.ListTapePoolsInput{})
	for pools.HasMorePages() {
		page, err := pools.NextPage(context.TODO())
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.PoolInfos, "aws_storagegateway_tape_pool",
			defaultAllowEmptyValues,
			func(p types.PoolInfo) string { return StringValue(p.PoolARN) },
			func(p types.PoolInfo) string { return StringValue(p.PoolName) })
	}

	for fs := storagegateway.NewListFileSharesPaginator(svc, &storagegateway.ListFileSharesInput{}); fs.HasMorePages(); {
		page, err := fs.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, f := range page.FileShareInfoList {
			arn := StringValue(f.FileShareARN)
			if arn == "" {
				continue
			}
			tfType := ""
			switch f.FileShareType {
			case "NFS":
				tfType = "aws_storagegateway_nfs_file_share"
			case "SMB":
				tfType = "aws_storagegateway_smb_file_share"
			default:
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, tfType, "aws", defaultAllowEmptyValues))
		}
	}

	for vol := storagegateway.NewListVolumesPaginator(svc, &storagegateway.ListVolumesInput{}); vol.HasMorePages(); {
		page, err := vol.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, v := range page.VolumeInfos {
			arn := StringValue(v.VolumeARN)
			if arn == "" {
				continue
			}
			tfType := ""
			switch StringValue(v.VolumeType) {
			case "CACHED":
				tfType = "aws_storagegateway_cached_iscsi_volume"
			case "STORED":
				tfType = "aws_storagegateway_stored_iscsi_volume"
			default:
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, tfType, "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
