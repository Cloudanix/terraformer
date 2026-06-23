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

	"github.com/aws/aws-sdk-go-v2/service/networkflowmonitor"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type NetworkFlowMonitorGenerator struct {
	AWSService
}

func (g *NetworkFlowMonitorGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := networkflowmonitor.NewFromConfig(config)
	for p := networkflowmonitor.NewListMonitorsPaginator(svc, &networkflowmonitor.ListMonitorsInput{}); p.HasMorePages(); {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, m := range page.Monitors {
			name := StringValue(m.MonitorName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_networkflowmonitor_monitor", "aws", defaultAllowEmptyValues))
		}
	}
	for p := networkflowmonitor.NewListScopesPaginator(svc, &networkflowmonitor.ListScopesInput{}); p.HasMorePages(); {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, s := range page.Scopes {
			id := StringValue(s.ScopeId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_networkflowmonitor_scope", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
