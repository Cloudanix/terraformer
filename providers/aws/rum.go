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

	"github.com/aws/aws-sdk-go-v2/service/rum"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type RumGenerator struct {
	AWSService
}

// InitResources enumerates CloudWatch RUM app monitors. Import ID is the name.
func (g *RumGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := rum.NewFromConfig(config)

	p := rum.NewListAppMonitorsPaginator(svc, &rum.ListAppMonitorsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, monitor := range page.AppMonitorSummaries {
			name := StringValue(monitor.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_rum_app_monitor", "aws", defaultAllowEmptyValues))

			if dest, err := svc.ListRumMetricsDestinations(context.TODO(), &rum.ListRumMetricsDestinationsInput{AppMonitorName: monitor.Name}); err == nil && len(dest.Destinations) > 0 {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					name, name, "aws_rum_metrics_destination", "aws", defaultAllowEmptyValues))
			}
		}
	}
	return nil
}
