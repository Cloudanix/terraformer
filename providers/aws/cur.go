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
	"github.com/aws/aws-sdk-go-v2/service/costandusagereportservice"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type CURGenerator struct {
	AWSService
}

// InitResources enumerates Cost & Usage Report definitions (us-east-1 only).
// Import ID is the report name.
func (g *CURGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := costandusagereportservice.NewFromConfig(config)
	p := costandusagereportservice.NewDescribeReportDefinitionsPaginator(svc, &costandusagereportservice.DescribeReportDefinitionsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, r := range page.ReportDefinitions {
			name := StringValue(r.ReportName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_cur_report_definition", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
