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

	"github.com/aws/aws-sdk-go-v2/service/redshiftserverless"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type RedshiftServerlessGenerator struct {
	AWSService
}

// InitResources enumerates Redshift Serverless namespaces and workgroups.
// Import IDs are the namespace / workgroup name.
func (g *RedshiftServerlessGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := redshiftserverless.NewFromConfig(config)
	ctx := context.TODO()

	namespaces := redshiftserverless.NewListNamespacesPaginator(svc, &redshiftserverless.ListNamespacesInput{})
	for namespaces.HasMorePages() {
		page, err := namespaces.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, ns := range page.Namespaces {
			name := StringValue(ns.NamespaceName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_redshiftserverless_namespace", "aws", defaultAllowEmptyValues))
		}
	}

	workgroups := redshiftserverless.NewListWorkgroupsPaginator(svc, &redshiftserverless.ListWorkgroupsInput{})
	for workgroups.HasMorePages() {
		page, err := workgroups.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, wg := range page.Workgroups {
			name := StringValue(wg.WorkgroupName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_redshiftserverless_workgroup", "aws", defaultAllowEmptyValues))
		}
	}

	for sp := redshiftserverless.NewListSnapshotsPaginator(svc, &redshiftserverless.ListSnapshotsInput{}); sp.HasMorePages(); {
		page, err := sp.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, s := range page.Snapshots {
			name := StringValue(s.SnapshotName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_redshiftserverless_snapshot", "aws", defaultAllowEmptyValues))
		}
	}
	for ep := redshiftserverless.NewListEndpointAccessPaginator(svc, &redshiftserverless.ListEndpointAccessInput{}); ep.HasMorePages(); {
		page, err := ep.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, e := range page.Endpoints {
			name := StringValue(e.EndpointName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_redshiftserverless_endpoint_access", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
