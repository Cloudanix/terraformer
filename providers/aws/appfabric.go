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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appfabric"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type AppFabricGenerator struct {
	AWSService
}

func (g *AppFabricGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := appfabric.NewFromConfig(config)
	ctx := awsContext()
	var bundleArns []string
	p := appfabric.NewListAppBundlesPaginator(svc, &appfabric.ListAppBundlesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, item := range page.AppBundleSummaryList {
			id := StringValue(item.Arn)
			if id == "" {
				continue
			}
			bundleArns = append(bundleArns, id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_appfabric_app_bundle", "aws", defaultAllowEmptyValues))
		}
	}

	for _, bundleArn := range bundleArns {
		bundle := bundleArn
		for ap := appfabric.NewListAppAuthorizationsPaginator(svc, &appfabric.ListAppAuthorizationsInput{AppBundleIdentifier: &bundle}); ap.HasMorePages(); {
			page, err := ap.NextPage(ctx)
			if err != nil {
				break
			}
			for _, a := range page.AppAuthorizationSummaryList {
				arn := StringValue(a.AppAuthorizationArn)
				if arn == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					arn+","+bundle, arn, "aws_appfabric_app_authorization", "aws", defaultAllowEmptyValues))
			}
		}
		for ip := appfabric.NewListIngestionsPaginator(svc, &appfabric.ListIngestionsInput{AppBundleIdentifier: &bundle}); ip.HasMorePages(); {
			page, err := ip.NextPage(ctx)
			if err != nil {
				break
			}
			for _, ing := range page.Ingestions {
				ingArn := StringValue(ing.Arn)
				if ingArn == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					ingArn+","+bundle, ingArn, "aws_appfabric_ingestion", "aws", defaultAllowEmptyValues))
				for dp := appfabric.NewListIngestionDestinationsPaginator(svc, &appfabric.ListIngestionDestinationsInput{AppBundleIdentifier: &bundle, IngestionIdentifier: aws.String(ingArn)}); dp.HasMorePages(); {
					dpage, err := dp.NextPage(ctx)
					if err != nil {
						break
					}
					for _, d := range dpage.IngestionDestinations {
						dArn := StringValue(d.Arn)
						if dArn == "" {
							continue
						}
						g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
							dArn+","+ingArn+","+bundle, dArn, "aws_appfabric_ingestion_destination", "aws", defaultAllowEmptyValues))
					}
				}
			}
		}
	}
	return nil
}
