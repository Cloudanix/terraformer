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
	"github.com/aws/aws-sdk-go-v2/service/emrcontainers"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type EMRContainersGenerator struct {
	AWSService
}

// InitResources enumerates EMR on EKS virtual clusters. Import ID is the cluster id.
func (g *EMRContainersGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := emrcontainers.NewFromConfig(config)

	p := emrcontainers.NewListVirtualClustersPaginator(svc, &emrcontainers.ListVirtualClustersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, vc := range page.VirtualClusters {
			id := StringValue(vc.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(vc.Name), "aws_emrcontainers_virtual_cluster", "aws", defaultAllowEmptyValues))
		}
	}

	for jt := emrcontainers.NewListJobTemplatesPaginator(svc, &emrcontainers.ListJobTemplatesInput{}); jt.HasMorePages(); {
		page, err := jt.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, t := range page.Templates {
			id := StringValue(t.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(t.Name), "aws_emrcontainers_job_template", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
