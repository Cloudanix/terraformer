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
	"github.com/aws/aws-sdk-go-v2/service/fsx"
	"github.com/aws/aws-sdk-go-v2/service/fsx/types"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type FsxGenerator struct {
	AWSService
}

// fsxFileSystemResourceType maps an FSx file-system engine to its Terraform
// resource type. DescribeFileSystems returns every engine in one stream, so the
// type is branched per item.
func fsxFileSystemResourceType(t types.FileSystemType) string {
	switch t {
	case types.FileSystemTypeWindows:
		return "aws_fsx_windows_file_system"
	case types.FileSystemTypeLustre:
		return "aws_fsx_lustre_file_system"
	case types.FileSystemTypeOntap:
		return "aws_fsx_ontap_file_system"
	case types.FileSystemTypeOpenzfs:
		return "aws_fsx_openzfs_file_system"
	default:
		return ""
	}
}

func fsxVolumeResourceType(t types.VolumeType) string {
	switch t {
	case types.VolumeTypeOntap:
		return "aws_fsx_ontap_volume"
	case types.VolumeTypeOpenzfs:
		return "aws_fsx_openzfs_volume"
	default:
		return ""
	}
}

// InitResources enumerates FSx file systems, backups, ONTAP storage virtual
// machines, and volumes. Import IDs are the resource IDs; file systems and
// volumes resolve their concrete Terraform type from the engine type.
func (g *FsxGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := fsx.NewFromConfig(config)
	ctx := awsContext()

	fileSystems := fsx.NewDescribeFileSystemsPaginator(svc, &fsx.DescribeFileSystemsInput{})
	for fileSystems.HasMorePages() {
		page, err := fileSystems.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, fileSystem := range page.FileSystems {
			id := StringValue(fileSystem.FileSystemId)
			resourceType := fsxFileSystemResourceType(fileSystem.FileSystemType)
			if id == "" || resourceType == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, resourceType, "aws", defaultAllowEmptyValues))
		}
	}

	backups := fsx.NewDescribeBackupsPaginator(svc, &fsx.DescribeBackupsInput{})
	for backups.HasMorePages() {
		page, err := backups.NextPage(ctx)
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.Backups, "aws_fsx_backup",
			defaultAllowEmptyValues,
			func(b types.Backup) string { return StringValue(b.BackupId) },
			func(b types.Backup) string { return StringValue(b.BackupId) })
	}

	svms := fsx.NewDescribeStorageVirtualMachinesPaginator(svc, &fsx.DescribeStorageVirtualMachinesInput{})
	for svms.HasMorePages() {
		page, err := svms.NextPage(ctx)
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.StorageVirtualMachines, "aws_fsx_ontap_storage_virtual_machine",
			defaultAllowEmptyValues,
			func(s types.StorageVirtualMachine) string { return StringValue(s.StorageVirtualMachineId) },
			func(s types.StorageVirtualMachine) string { return StringValue(s.StorageVirtualMachineId) })
	}

	volumes := fsx.NewDescribeVolumesPaginator(svc, &fsx.DescribeVolumesInput{})
	for volumes.HasMorePages() {
		page, err := volumes.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, volume := range page.Volumes {
			id := StringValue(volume.VolumeId)
			resourceType := fsxVolumeResourceType(volume.VolumeType)
			if id == "" || resourceType == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, resourceType, "aws", defaultAllowEmptyValues))
		}
	}

	dras := fsx.NewDescribeDataRepositoryAssociationsPaginator(svc, &fsx.DescribeDataRepositoryAssociationsInput{})
	for dras.HasMorePages() {
		page, err := dras.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, a := range page.Associations {
			id := StringValue(a.AssociationId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_fsx_data_repository_association", "aws", defaultAllowEmptyValues))
		}
	}

	caches := fsx.NewDescribeFileCachesPaginator(svc, &fsx.DescribeFileCachesInput{})
	for caches.HasMorePages() {
		page, err := caches.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, c := range page.FileCaches {
			id := StringValue(c.FileCacheId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_fsx_file_cache", "aws", defaultAllowEmptyValues))
		}
	}

	snaps := fsx.NewDescribeSnapshotsPaginator(svc, &fsx.DescribeSnapshotsInput{})
	for snaps.HasMorePages() {
		page, err := snaps.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, sn := range page.Snapshots {
			id := StringValue(sn.SnapshotId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(sn.Name), "aws_fsx_openzfs_snapshot", "aws", defaultAllowEmptyValues))
		}
	}

	return nil
}
