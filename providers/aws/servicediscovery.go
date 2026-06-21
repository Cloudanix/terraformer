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
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery/types"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type ServiceDiscoveryGenerator struct {
	AWSService
}

// namespaceResourceType maps a Cloud Map namespace type to its Terraform
// resource type. ListNamespaces returns all namespaces in one stream, so the
// type has to be branched per item rather than per API call.
func namespaceResourceType(t types.NamespaceType) string {
	switch t {
	case types.NamespaceTypeHttp:
		return "aws_service_discovery_http_namespace"
	case types.NamespaceTypeDnsPublic:
		return "aws_service_discovery_public_dns_namespace"
	case types.NamespaceTypeDnsPrivate:
		return "aws_service_discovery_private_dns_namespace"
	default:
		return ""
	}
}

// InitResources enumerates Cloud Map (Service Discovery) namespaces, services,
// and registered instances. Import IDs:
//   - namespaces / service → the resource Id
//   - aws_service_discovery_instance → "<service-id>/<instance-id>"
func (g *ServiceDiscoveryGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := servicediscovery.NewFromConfig(config)
	ctx := context.TODO()

	namespaces := servicediscovery.NewListNamespacesPaginator(svc, &servicediscovery.ListNamespacesInput{})
	for namespaces.HasMorePages() {
		page, err := namespaces.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, ns := range page.Namespaces {
			id := StringValue(ns.Id)
			resourceType := namespaceResourceType(ns.Type)
			if id == "" || resourceType == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(ns.Name), resourceType, "aws", defaultAllowEmptyValues))
		}
	}

	var serviceIDs []string
	services := servicediscovery.NewListServicesPaginator(svc, &servicediscovery.ListServicesInput{})
	for services.HasMorePages() {
		page, err := services.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, s := range page.Services {
			id := StringValue(s.Id)
			if id == "" {
				continue
			}
			serviceIDs = append(serviceIDs, id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(s.Name), "aws_service_discovery_service", "aws", defaultAllowEmptyValues))
		}
	}

	for _, serviceID := range serviceIDs {
		instances := servicediscovery.NewListInstancesPaginator(svc, &servicediscovery.ListInstancesInput{ServiceId: aws.String(serviceID)})
		for instances.HasMorePages() {
			page, err := instances.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, inst := range page.Instances {
				instanceID := StringValue(inst.Id)
				if instanceID == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					fmt.Sprintf("%s/%s", serviceID, instanceID),
					instanceID,
					"aws_service_discovery_instance", "aws", defaultAllowEmptyValues))
			}
		}
	}

	return nil
}
