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

	"github.com/aws/aws-sdk-go-v2/service/licensemanager"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type LicenseManagerGenerator struct {
	AWSService
}

// InitResources enumerates License Manager license configurations. Import ID is
// the configuration ARN.
func (g *LicenseManagerGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := licensemanager.NewFromConfig(config)

	var token *string
	for {
		out, err := svc.ListLicenseConfigurations(context.TODO(), &licensemanager.ListLicenseConfigurationsInput{NextToken: token})
		if err != nil {
			return err
		}
		for _, lc := range out.LicenseConfigurations {
			arn := StringValue(lc.LicenseConfigurationArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(lc.Name), "aws_licensemanager_license_configuration", "aws", defaultAllowEmptyValues))
		}
		if out.NextToken == nil {
			break
		}
		token = out.NextToken
	}

	// Distributed (granted-out) and received grants.
	var gToken *string
	for {
		out, err := svc.ListDistributedGrants(context.TODO(), &licensemanager.ListDistributedGrantsInput{NextToken: gToken})
		if err != nil {
			break
		}
		for _, grant := range out.Grants {
			arn := StringValue(grant.GrantArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_licensemanager_grant", "aws", defaultAllowEmptyValues))
		}
		if out.NextToken == nil {
			break
		}
		gToken = out.NextToken
	}
	return nil
}
