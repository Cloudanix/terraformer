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
	"github.com/aws/aws-sdk-go-v2/service/odb"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type ODBGenerator struct {
	AWSService
}

// InitResources enumerates Oracle Database@AWS networks, VM clusters, Exadata
// infrastructures, autonomous VM clusters, and peering connections (import by id).
func (g *ODBGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := odb.NewFromConfig(config)
	ctx := awsContext()
	for p := odb.NewListOdbNetworksPaginator(svc, &odb.ListOdbNetworksInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range page.OdbNetworks {
			if id := StringValue(x.OdbNetworkId); id != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, id, "aws_odb_network", "aws", defaultAllowEmptyValues))
			}
		}
	}
	for p := odb.NewListCloudVmClustersPaginator(svc, &odb.ListCloudVmClustersInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range page.CloudVmClusters {
			if id := StringValue(x.CloudVmClusterId); id != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, id, "aws_odb_cloud_vm_cluster", "aws", defaultAllowEmptyValues))
			}
		}
	}
	for p := odb.NewListCloudExadataInfrastructuresPaginator(svc, &odb.ListCloudExadataInfrastructuresInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range page.CloudExadataInfrastructures {
			if id := StringValue(x.CloudExadataInfrastructureId); id != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, id, "aws_odb_cloud_exadata_infrastructure", "aws", defaultAllowEmptyValues))
			}
		}
	}
	for p := odb.NewListCloudAutonomousVmClustersPaginator(svc, &odb.ListCloudAutonomousVmClustersInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range page.CloudAutonomousVmClusters {
			if id := StringValue(x.CloudAutonomousVmClusterId); id != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, id, "aws_odb_cloud_autonomous_vm_cluster", "aws", defaultAllowEmptyValues))
			}
		}
	}
	for p := odb.NewListOdbPeeringConnectionsPaginator(svc, &odb.ListOdbPeeringConnectionsInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range page.OdbPeeringConnections {
			if id := StringValue(x.OdbPeeringConnectionId); id != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, id, "aws_odb_network_peering_connection", "aws", defaultAllowEmptyValues))
			}
		}
	}
	return nil
}
