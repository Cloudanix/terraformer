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
	"github.com/aws/aws-sdk-go-v2/service/appmesh"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type AppMeshGenerator struct {
	AWSService
}

// InitResources enumerates App Mesh service meshes. Import ID is the mesh name.
func (g *AppMeshGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := appmesh.NewFromConfig(config)

	p := appmesh.NewListMeshesPaginator(svc, &appmesh.ListMeshesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, mesh := range page.Meshes {
			name := StringValue(mesh.MeshName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_appmesh_mesh", "aws", defaultAllowEmptyValues))
			g.loadMeshChildren(svc, name)
		}
	}
	return nil
}

// loadMeshChildren enumerates a mesh's virtual nodes/routers/services/gateways.
// Import IDs are "<mesh-name>/<resource-name>".
func (g *AppMeshGenerator) loadMeshChildren(svc *appmesh.Client, mesh string) {
	ctx := context.TODO()
	add := func(name, tfType string) {
		if name != "" {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				fmt.Sprintf("%s/%s", mesh, name), fmt.Sprintf("%s_%s", mesh, name),
				tfType, "aws", defaultAllowEmptyValues))
		}
	}
	for p := appmesh.NewListVirtualNodesPaginator(svc, &appmesh.ListVirtualNodesInput{MeshName: aws.String(mesh)}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range page.VirtualNodes {
			add(StringValue(x.VirtualNodeName), "aws_appmesh_virtual_node")
		}
	}
	for p := appmesh.NewListVirtualRoutersPaginator(svc, &appmesh.ListVirtualRoutersInput{MeshName: aws.String(mesh)}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range page.VirtualRouters {
			add(StringValue(x.VirtualRouterName), "aws_appmesh_virtual_router")
		}
	}
	for p := appmesh.NewListVirtualServicesPaginator(svc, &appmesh.ListVirtualServicesInput{MeshName: aws.String(mesh)}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range page.VirtualServices {
			add(StringValue(x.VirtualServiceName), "aws_appmesh_virtual_service")
		}
	}
	for p := appmesh.NewListVirtualGatewaysPaginator(svc, &appmesh.ListVirtualGatewaysInput{MeshName: aws.String(mesh)}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range page.VirtualGateways {
			add(StringValue(x.VirtualGatewayName), "aws_appmesh_virtual_gateway")
		}
	}
}
