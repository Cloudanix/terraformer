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
	"github.com/aws/aws-sdk-go-v2/service/cloudsearch"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type CloudSearchGenerator struct {
	AWSService
}

// InitResources enumerates CloudSearch domains (DescribeDomains is a single
// call, no paginator). Import ID is the domain name.
func (g *CloudSearchGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := cloudsearch.NewFromConfig(config)

	out, err := svc.DescribeDomains(awsContext(), &cloudsearch.DescribeDomainsInput{})
	if err != nil {
		return err
	}
	for _, d := range out.DomainStatusList {
		name := StringValue(d.DomainName)
		if name == "" {
			continue
		}
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			name, name, "aws_cloudsearch_domain", "aws", defaultAllowEmptyValues))
		if pol, err := svc.DescribeServiceAccessPolicies(awsContext(), &cloudsearch.DescribeServiceAccessPoliciesInput{DomainName: d.DomainName}); err == nil &&
			pol.AccessPolicies != nil && StringValue(pol.AccessPolicies.Options) != "" {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_cloudsearch_domain_service_access_policy", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
