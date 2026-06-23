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
	"github.com/aws/aws-sdk-go-v2/service/opensearch"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type OpenSearchGenerator struct {
	AWSService
}

// InitResources enumerates OpenSearch domains (the newer service; classic
// Elasticsearch lives under the "es" generator). Import ID is the domain name.
func (g *OpenSearchGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := opensearch.NewFromConfig(config)

	ctx := awsContext()
	out, err := svc.ListDomainNames(ctx, &opensearch.ListDomainNamesInput{})
	if err != nil {
		return err
	}
	for _, domain := range out.DomainNames {
		name := StringValue(domain.DomainName)
		if name == "" {
			continue
		}
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			name, name, "aws_opensearch_domain", "aws", defaultAllowEmptyValues))

		for pp := opensearch.NewListPackagesForDomainPaginator(svc, &opensearch.ListPackagesForDomainInput{DomainName: domain.DomainName}); pp.HasMorePages(); {
			ppage, err := pp.NextPage(ctx)
			if err != nil {
				break
			}
			for _, pkg := range ppage.DomainPackageDetailsList {
				pkgID := StringValue(pkg.PackageID)
				if pkgID == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					pkgID+"-"+name, pkgID+"_"+name, "aws_opensearch_package_association", "aws", defaultAllowEmptyValues))
			}
		}

		cfg, err := svc.DescribeDomainConfig(ctx, &opensearch.DescribeDomainConfigInput{DomainName: domain.DomainName})
		if err != nil || cfg.DomainConfig == nil {
			continue
		}
		if cfg.DomainConfig.AccessPolicies != nil && StringValue(cfg.DomainConfig.AccessPolicies.Options) != "" {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_opensearch_domain_policy", "aws", defaultAllowEmptyValues))
		}
		if cfg.DomainConfig.AdvancedSecurityOptions != nil &&
			cfg.DomainConfig.AdvancedSecurityOptions.Options != nil &&
			cfg.DomainConfig.AdvancedSecurityOptions.Options.SAMLOptions != nil {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_opensearch_domain_saml_options", "aws", defaultAllowEmptyValues))
		}
	}

	if eps, err := svc.ListVpcEndpoints(ctx, &opensearch.ListVpcEndpointsInput{}); err == nil {
		for _, ep := range eps.VpcEndpointSummaryList {
			id := StringValue(ep.VpcEndpointId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_opensearch_vpc_endpoint", "aws", defaultAllowEmptyValues))
		}
	}

	for p := opensearch.NewDescribeOutboundConnectionsPaginator(svc, &opensearch.DescribeOutboundConnectionsInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, c := range page.Connections {
			id := StringValue(c.ConnectionId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_opensearch_outbound_connection", "aws", defaultAllowEmptyValues))
		}
	}

	for p := opensearch.NewDescribePackagesPaginator(svc, &opensearch.DescribePackagesInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, pkg := range page.PackageDetailsList {
			id := StringValue(pkg.PackageID)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_opensearch_package", "aws", defaultAllowEmptyValues))
		}
	}

	for ap := opensearch.NewListApplicationsPaginator(svc, &opensearch.ListApplicationsInput{}); ap.HasMorePages(); {
		page, err := ap.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, a := range page.ApplicationSummaries {
			id := StringValue(a.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(a.Name), "aws_opensearch_application", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
