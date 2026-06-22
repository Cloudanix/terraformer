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

	"github.com/aws/aws-sdk-go-v2/service/customerprofiles"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type CustomerProfilesGenerator struct {
	AWSService
}

// InitResources enumerates Customer Profiles domains. Import ID is the domain name.
func (g *CustomerProfilesGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := customerprofiles.NewFromConfig(config)

	var token *string
	for {
		out, err := svc.ListDomains(context.TODO(), &customerprofiles.ListDomainsInput{NextToken: token})
		if err != nil {
			return err
		}
		for _, domain := range out.Items {
			name := StringValue(domain.DomainName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_customerprofiles_domain", "aws", defaultAllowEmptyValues))
		}
		if out.NextToken == nil {
			return nil
		}
		token = out.NextToken
	}
}
