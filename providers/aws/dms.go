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
	dms "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"
	"github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
)

type DmsGenerator struct {
	AWSService
}

// InitResources enumerates Database Migration Service resources. Import IDs are
// the resource identifiers (not ARNs):
//   - aws_dms_replication_instance     → replication instance identifier
//   - aws_dms_endpoint                 → endpoint identifier
//   - aws_dms_replication_task         → replication task identifier
//   - aws_dms_replication_subnet_group → subnet group identifier
//   - aws_dms_certificate              → certificate identifier
//   - aws_dms_event_subscription       → subscription name
func (g *DmsGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := dms.NewFromConfig(config)
	ctx := context.TODO()

	instances := dms.NewDescribeReplicationInstancesPaginator(svc, &dms.DescribeReplicationInstancesInput{})
	for instances.HasMorePages() {
		page, err := instances.NextPage(ctx)
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.ReplicationInstances, "aws_dms_replication_instance",
			defaultAllowEmptyValues,
			func(r types.ReplicationInstance) string { return StringValue(r.ReplicationInstanceIdentifier) },
			func(r types.ReplicationInstance) string { return StringValue(r.ReplicationInstanceIdentifier) })
	}

	endpoints := dms.NewDescribeEndpointsPaginator(svc, &dms.DescribeEndpointsInput{})
	for endpoints.HasMorePages() {
		page, err := endpoints.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, en := range page.Endpoints {
			id := StringValue(en.EndpointIdentifier)
			if id == "" {
				continue
			}
			// S3 endpoints have their own dedicated TF resource type.
			tfType := "aws_dms_endpoint"
			if StringValue(en.EngineName) == "s3" {
				tfType = "aws_dms_s3_endpoint"
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, tfType, "aws", defaultAllowEmptyValues))
		}
	}

	tasks := dms.NewDescribeReplicationTasksPaginator(svc, &dms.DescribeReplicationTasksInput{})
	for tasks.HasMorePages() {
		page, err := tasks.NextPage(ctx)
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.ReplicationTasks, "aws_dms_replication_task",
			defaultAllowEmptyValues,
			func(t types.ReplicationTask) string { return StringValue(t.ReplicationTaskIdentifier) },
			func(t types.ReplicationTask) string { return StringValue(t.ReplicationTaskIdentifier) })
	}

	subnetGroups := dms.NewDescribeReplicationSubnetGroupsPaginator(svc, &dms.DescribeReplicationSubnetGroupsInput{})
	for subnetGroups.HasMorePages() {
		page, err := subnetGroups.NextPage(ctx)
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.ReplicationSubnetGroups, "aws_dms_replication_subnet_group",
			defaultAllowEmptyValues,
			func(s types.ReplicationSubnetGroup) string { return StringValue(s.ReplicationSubnetGroupIdentifier) },
			func(s types.ReplicationSubnetGroup) string { return StringValue(s.ReplicationSubnetGroupIdentifier) })
	}

	certs := dms.NewDescribeCertificatesPaginator(svc, &dms.DescribeCertificatesInput{})
	for certs.HasMorePages() {
		page, err := certs.NextPage(ctx)
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.Certificates, "aws_dms_certificate",
			defaultAllowEmptyValues,
			func(c types.Certificate) string { return StringValue(c.CertificateIdentifier) },
			func(c types.Certificate) string { return StringValue(c.CertificateIdentifier) })
	}

	eventSubs := dms.NewDescribeEventSubscriptionsPaginator(svc, &dms.DescribeEventSubscriptionsInput{})
	for eventSubs.HasMorePages() {
		page, err := eventSubs.NextPage(ctx)
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.EventSubscriptionsList, "aws_dms_event_subscription",
			defaultAllowEmptyValues,
			func(es types.EventSubscription) string { return StringValue(es.CustSubscriptionId) },
			func(es types.EventSubscription) string { return StringValue(es.CustSubscriptionId) })
	}

	replConfigs := dms.NewDescribeReplicationConfigsPaginator(svc, &dms.DescribeReplicationConfigsInput{})
	for replConfigs.HasMorePages() {
		page, err := replConfigs.NextPage(ctx)
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.ReplicationConfigs, "aws_dms_replication_config",
			defaultAllowEmptyValues,
			func(rc types.ReplicationConfig) string { return StringValue(rc.ReplicationConfigArn) },
			func(rc types.ReplicationConfig) string { return StringValue(rc.ReplicationConfigIdentifier) })
	}

	return nil
}
