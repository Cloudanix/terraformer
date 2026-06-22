// Copyright 2019 The Terraformer Authors.
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
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform/helper/hashcode"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

var ebsAllowEmptyValues = []string{"tags."}

type EbsGenerator struct {
	AWSService
}

func (g *EbsGenerator) volumeAttachmentID(device, volumeID, instanceID string) string {
	return fmt.Sprintf("vai-%d", hashcode.String(fmt.Sprintf("%s-%s-%s-", device, instanceID, volumeID)))
}

func (g *EbsGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := ec2.NewFromConfig(config)
	var filters []types.Filter
	for _, filter := range g.Filter {
		if strings.HasPrefix(filter.FieldPath, "tags.") && filter.IsApplicable("ebs_volume") {
			filters = append(filters, types.Filter{
				Name:   aws.String("tag:" + strings.TrimPrefix(filter.FieldPath, "tags.")),
				Values: filter.AcceptableValues,
			})
		}
	}
	// Account/region-level EBS settings. Their Terraform import ID is the region.
	region := config.Region
	if region != "" {
		if enc, err := svc.GetEbsEncryptionByDefault(context.TODO(), &ec2.GetEbsEncryptionByDefaultInput{}); err == nil && enc.EbsEncryptionByDefault != nil && *enc.EbsEncryptionByDefault {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				region, region, "aws_ebs_encryption_by_default", "aws", ebsAllowEmptyValues))
		}
		if kms, err := svc.GetEbsDefaultKmsKeyId(context.TODO(), &ec2.GetEbsDefaultKmsKeyIdInput{}); err == nil && StringValue(kms.KmsKeyId) != "" {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				region, region, "aws_ebs_default_kms_key", "aws", ebsAllowEmptyValues))
		}
		if bpa, err := svc.GetSnapshotBlockPublicAccessState(context.TODO(), &ec2.GetSnapshotBlockPublicAccessStateInput{}); err == nil &&
			bpa.State != "" && bpa.State != types.SnapshotBlockPublicAccessStateUnblocked {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				region, region, "aws_ebs_snapshot_block_public_access", "aws", ebsAllowEmptyValues))
		}
	}

	p := ec2.NewDescribeVolumesPaginator(svc, &ec2.DescribeVolumesInput{
		Filters: filters,
	})
	for p.HasMorePages() {
		page, e := p.NextPage(context.TODO())
		if e != nil {
			return e
		}
		for _, volume := range page.Volumes {
			isRootDevice := false // Let's leave root device configuration to be done in ec2_instance resources

			for _, attachment := range volume.Attachments {
				instances, _ := svc.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{
					InstanceIds: []string{StringValue(attachment.InstanceId)},
				})
				for _, reservation := range instances.Reservations {
					for _, instance := range reservation.Instances {
						if StringValue(instance.RootDeviceName) == StringValue(attachment.Device) {
							isRootDevice = true
						}
					}
				}
			}

			if !isRootDevice {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					StringValue(volume.VolumeId),
					StringValue(volume.VolumeId),
					"aws_ebs_volume",
					"aws",
					ebsAllowEmptyValues,
				))

				for _, attachment := range volume.Attachments {
					if attachment.State == types.VolumeAttachmentStateAttached {
						attachmentID := g.volumeAttachmentID(
							StringValue(attachment.Device),
							StringValue(attachment.VolumeId),
							StringValue(attachment.InstanceId))
						g.Resources = append(g.Resources, terraformutils.NewResource(
							attachmentID,
							StringValue(attachment.InstanceId)+":"+StringValue(attachment.Device),
							"aws_volume_attachment",
							"aws",
							map[string]string{
								"device_name": StringValue(attachment.Device),
								"volume_id":   StringValue(attachment.VolumeId),
								"instance_id": StringValue(attachment.InstanceId),
							},
							[]string{},
							map[string]interface{}{},
						))
					}
				}
			}
		}
	}
	return nil
}
