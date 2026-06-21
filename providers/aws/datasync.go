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
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/datasync"
	"github.com/aws/aws-sdk-go-v2/service/datasync/types"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type DataSyncGenerator struct {
	AWSService
}

// datasyncLocationResourceType maps a DataSync LocationUri scheme to its
// Terraform resource type. ListLocations returns every location regardless of
// type, so the type is recovered from the URI scheme (the part before "://").
// Unknown schemes are skipped rather than guessed.
func datasyncLocationResourceType(uri string) string {
	scheme, _, ok := strings.Cut(uri, "://")
	if !ok {
		return ""
	}
	switch scheme {
	case "s3":
		return "aws_datasync_location_s3"
	case "nfs":
		return "aws_datasync_location_nfs"
	case "smb":
		return "aws_datasync_location_smb"
	case "efs":
		return "aws_datasync_location_efs"
	case "hdfs":
		return "aws_datasync_location_hdfs"
	case "object-storage":
		return "aws_datasync_location_object_storage"
	case "azure-blob":
		return "aws_datasync_location_azure_blob"
	case "fsxw":
		return "aws_datasync_location_fsx_windows_file_system"
	case "fsxl":
		return "aws_datasync_location_fsx_lustre_file_system"
	case "fsxz":
		return "aws_datasync_location_fsx_openzfs_file_system"
	case "fsxn":
		return "aws_datasync_location_fsx_ontap_file_system"
	default:
		return ""
	}
}

// InitResources enumerates DataSync tasks, agents, and locations. The Terraform
// import ID for each is the resource ARN; the location's concrete resource type
// is derived from its URI scheme.
func (g *DataSyncGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := datasync.NewFromConfig(config)
	ctx := context.TODO()

	tasks := datasync.NewListTasksPaginator(svc, &datasync.ListTasksInput{})
	for tasks.HasMorePages() {
		page, err := tasks.NextPage(ctx)
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.Tasks, "aws_datasync_task",
			defaultAllowEmptyValues,
			func(t types.TaskListEntry) string { return StringValue(t.TaskArn) },
			func(t types.TaskListEntry) string { return StringValue(t.Name) })
	}

	agents := datasync.NewListAgentsPaginator(svc, &datasync.ListAgentsInput{})
	for agents.HasMorePages() {
		page, err := agents.NextPage(ctx)
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.Agents, "aws_datasync_agent",
			defaultAllowEmptyValues,
			func(a types.AgentListEntry) string { return StringValue(a.AgentArn) },
			func(a types.AgentListEntry) string { return StringValue(a.Name) })
	}

	locations := datasync.NewListLocationsPaginator(svc, &datasync.ListLocationsInput{})
	for locations.HasMorePages() {
		page, err := locations.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, loc := range page.Locations {
			arn := StringValue(loc.LocationArn)
			resourceType := datasyncLocationResourceType(StringValue(loc.LocationUri))
			if arn == "" || resourceType == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, resourceType, "aws", defaultAllowEmptyValues))
		}
	}

	return nil
}
