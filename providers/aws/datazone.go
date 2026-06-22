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

	"github.com/aws/aws-sdk-go-v2/service/datazone"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type DataZoneGenerator struct {
	AWSService
}

// InitResources enumerates DataZone domains. Import ID is the domain id.
func (g *DataZoneGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := datazone.NewFromConfig(config)

	p := datazone.NewListDomainsPaginator(svc, &datazone.ListDomainsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, domain := range page.Items {
			id := StringValue(domain.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(domain.Name), "aws_datazone_domain", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
