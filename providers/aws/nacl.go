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
	"strconv"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

var NaclAllowEmptyValues = []string{"tags."}

type NaclGenerator struct {
	AWSService
}

// naclRuleImportID builds the aws_network_acl_rule import ID
// ("<nacl>:<rule>:<protocol>:<egress>") for one network ACL entry, returning
// ok=false for entries that must be skipped (no rule number, or the implicit
// 32767 default-deny rule which is not a manageable resource).
func naclRuleImportID(naclID string, ruleNumber *int32, protocol string, egress *bool) (string, bool) {
	if ruleNumber == nil || *ruleNumber == 32767 {
		return "", false
	}
	egressStr := "false"
	if egress != nil && *egress {
		egressStr = "true"
	}
	return naclID + ":" + strconv.Itoa(int(*ruleNumber)) + ":" + protocol + ":" + egressStr, true
}

func (NaclGenerator) createResources(nacls *ec2.DescribeNetworkAclsOutput) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	var resourceType string
	for _, nacl := range nacls.NetworkAcls {
		isDefault := nacl.IsDefault != nil && *nacl.IsDefault
		if isDefault {
			resourceType = "aws_default_network_acl"
		} else {
			resourceType = "aws_network_acl"
		}
		naclID := StringValue(nacl.NetworkAclId)
		resources = append(resources, terraformutils.NewSimpleResource(
			naclID,
			naclID,
			resourceType,
			"aws",
			NaclAllowEmptyValues))
		// Standalone rules only for non-default ACLs; skip the implicit 32767 deny.
		if isDefault {
			continue
		}
		for _, entry := range nacl.Entries {
			id, ok := naclRuleImportID(naclID, entry.RuleNumber, StringValue(entry.Protocol), entry.Egress)
			if !ok {
				continue
			}
			resources = append(resources, terraformutils.NewSimpleResource(
				id, id, "aws_network_acl_rule", "aws", NaclAllowEmptyValues))
		}
	}
	return resources
}

// Generate TerraformResources from AWS API,
// from each network ACL create 1 TerraformResource.
// Need NetworkAclId as ID for terraform resource
func (g *NaclGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := ec2.NewFromConfig(config)
	p := ec2.NewDescribeNetworkAclsPaginator(svc, &ec2.DescribeNetworkAclsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		g.Resources = append(g.Resources, g.createResources(page)...)
	}
	return nil
}
