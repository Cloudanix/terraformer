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

	"github.com/aws/aws-sdk-go-v2/service/schemas"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type SchemasGenerator struct {
	AWSService
}

// InitResources enumerates EventBridge Schemas registries. Import ID is the name.
func (g *SchemasGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := schemas.NewFromConfig(config)

	ctx := context.TODO()
	var registryNames []string
	p := schemas.NewListRegistriesPaginator(svc, &schemas.ListRegistriesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, reg := range page.Registries {
			name := StringValue(reg.RegistryName)
			if name == "" {
				continue
			}
			registryNames = append(registryNames, name)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_schemas_registry", "aws", defaultAllowEmptyValues))
		}
	}

	for d := schemas.NewListDiscoverersPaginator(svc, &schemas.ListDiscoverersInput{}); d.HasMorePages(); {
		page, err := d.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, disc := range page.Discoverers {
			id := StringValue(disc.DiscovererId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_schemas_discoverer", "aws", defaultAllowEmptyValues))
		}
	}

	for _, registry := range registryNames {
		reg := registry
		if _, err := svc.GetResourcePolicy(ctx, &schemas.GetResourcePolicyInput{RegistryName: &reg}); err == nil {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				reg, reg, "aws_schemas_registry_policy", "aws", defaultAllowEmptyValues))
		}
		for sp := schemas.NewListSchemasPaginator(svc, &schemas.ListSchemasInput{RegistryName: &reg}); sp.HasMorePages(); {
			page, err := sp.NextPage(ctx)
			if err != nil {
				break
			}
			for _, s := range page.Schemas {
				name := StringValue(s.SchemaName)
				if name == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					reg+":"+name, reg+"_"+name, "aws_schemas_schema", "aws", defaultAllowEmptyValues))
			}
		}
	}
	return nil
}
