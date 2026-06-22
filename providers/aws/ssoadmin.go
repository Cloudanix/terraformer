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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type SSOAdminGenerator struct {
	AWSService
}

// InitResources enumerates IAM Identity Center (SSO) permission sets and their
// managed-policy attachments for every Identity Center instance. Import IDs are
// comma-joined composites the aws provider expects:
//   - aws_ssoadmin_permission_set            → "<permission_set_arn>,<instance_arn>"
//   - aws_ssoadmin_managed_policy_attachment → "<managed_policy_arn>,<permission_set_arn>,<instance_arn>"
//
// Account assignments are intentionally skipped: enumerating them needs the org
// account list + principal IDs (Organizations dependency), out of scope here.
func (g *SSOAdminGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := ssoadmin.NewFromConfig(config)
	ctx := context.TODO()

	instances, err := svc.ListInstances(ctx, &ssoadmin.ListInstancesInput{})
	if err != nil {
		return err
	}

	for _, instance := range instances.Instances {
		instanceArn := StringValue(instance.InstanceArn)
		if instanceArn == "" {
			continue
		}

		permissionSets := ssoadmin.NewListPermissionSetsPaginator(svc, &ssoadmin.ListPermissionSetsInput{InstanceArn: aws.String(instanceArn)})
		for permissionSets.HasMorePages() {
			page, err := permissionSets.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, psArn := range page.PermissionSets {
				if psArn == "" {
					continue
				}
				name := psArn
				if described, err := svc.DescribePermissionSet(ctx, &ssoadmin.DescribePermissionSetInput{
					InstanceArn: aws.String(instanceArn), PermissionSetArn: aws.String(psArn),
				}); err == nil && described.PermissionSet != nil {
					name = StringValue(described.PermissionSet.Name)
				}
				g.Resources = append(g.Resources, terraformutils.NewResource(
					psArn+","+instanceArn,
					name,
					"aws_ssoadmin_permission_set",
					"aws",
					map[string]string{"instance_arn": instanceArn, "arn": psArn},
					defaultAllowEmptyValues,
					map[string]interface{}{},
				))

				policies := ssoadmin.NewListManagedPoliciesInPermissionSetPaginator(svc, &ssoadmin.ListManagedPoliciesInPermissionSetInput{
					InstanceArn: aws.String(instanceArn), PermissionSetArn: aws.String(psArn),
				})
				for policies.HasMorePages() {
					polPage, err := policies.NextPage(ctx)
					if err != nil {
						return err
					}
					for _, pol := range polPage.AttachedManagedPolicies {
						policyArn := StringValue(pol.Arn)
						if policyArn == "" {
							continue
						}
						g.Resources = append(g.Resources, terraformutils.NewResource(
							policyArn+","+psArn+","+instanceArn,
							StringValue(pol.Name),
							"aws_ssoadmin_managed_policy_attachment",
							"aws",
							map[string]string{"managed_policy_arn": policyArn, "permission_set_arn": psArn, "instance_arn": instanceArn},
							defaultAllowEmptyValues,
							map[string]interface{}{},
						))
					}
				}
			}
		}

		apps := ssoadmin.NewListApplicationsPaginator(svc, &ssoadmin.ListApplicationsInput{InstanceArn: aws.String(instanceArn)})
		for apps.HasMorePages() {
			page, err := apps.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, app := range page.Applications {
				arn := StringValue(app.ApplicationArn)
				if arn == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					arn, StringValue(app.Name), "aws_ssoadmin_application", "aws", defaultAllowEmptyValues))
			}
		}

		issuers := ssoadmin.NewListTrustedTokenIssuersPaginator(svc, &ssoadmin.ListTrustedTokenIssuersInput{InstanceArn: aws.String(instanceArn)})
		for issuers.HasMorePages() {
			page, err := issuers.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, issuer := range page.TrustedTokenIssuers {
				arn := StringValue(issuer.TrustedTokenIssuerArn)
				if arn == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					arn, StringValue(issuer.Name), "aws_ssoadmin_trusted_token_issuer", "aws", defaultAllowEmptyValues))
			}
		}
	}

	return nil
}
