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
	"github.com/aws/aws-sdk-go-v2/service/networkmonitor"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type NetworkMonitorGenerator struct {
	AWSService
}

// InitResources enumerates Network Monitor monitors. Import ID is the name.
func (g *NetworkMonitorGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := networkmonitor.NewFromConfig(config)

	p := networkmonitor.NewListMonitorsPaginator(svc, &networkmonitor.ListMonitorsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, monitor := range page.Monitors {
			name := StringValue(monitor.MonitorName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_networkmonitor_monitor", "aws", defaultAllowEmptyValues))

			if mon, err := svc.GetMonitor(awsContext(), &networkmonitor.GetMonitorInput{MonitorName: monitor.MonitorName}); err == nil {
				for _, probe := range mon.Probes {
					probeID := StringValue(probe.ProbeId)
					if probeID == "" {
						continue
					}
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						name+","+probeID, name+"_"+probeID, "aws_networkmonitor_probe", "aws", defaultAllowEmptyValues))
				}
			}
		}
	}
	return nil
}
