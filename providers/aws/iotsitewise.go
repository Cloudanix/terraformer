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
	"github.com/aws/aws-sdk-go-v2/service/iotsitewise"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type IoTSiteWiseGenerator struct {
	AWSService
}

// InitResources enumerates IoT SiteWise asset models. Import ID is the model id.
func (g *IoTSiteWiseGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := iotsitewise.NewFromConfig(config)

	p := iotsitewise.NewListAssetModelsPaginator(svc, &iotsitewise.ListAssetModelsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, am := range page.AssetModelSummaries {
			id := StringValue(am.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(am.Name), "aws_iotsitewise_asset_model", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
