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
	"fmt"
	"log"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

var IamAllowEmptyValues = []string{"tags."}

var IamAdditionalFields = map[string]interface{}{}

type IamGenerator struct {
	AWSService
}

// isServiceLinkedRolePath reports whether an IAM role path identifies a
// service-linked role (which maps to aws_iam_service_linked_role, not
// aws_iam_role).
func isServiceLinkedRolePath(path string) bool {
	return strings.HasPrefix(path, "/aws-service-role/")
}

func (g *IamGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := iam.NewFromConfig(config)
	g.Resources = []terraformutils.Resource{}
	err := g.getUsers(svc)
	if err != nil {
		log.Println(err)
	}

	err = g.getGroups(svc)
	if err != nil {
		log.Println(err)
	}

	err = g.getPolicies(svc)
	if err != nil {
		log.Println(err)
	}

	err = g.getRoles(svc)
	if err != nil {
		log.Println(err)
	}

	err = g.getInstanceProfiles(svc)
	if err != nil {
		log.Println(err)
	}

	g.getAccountResources(svc)

	return nil
}

// getAccountResources enumerates account-level IAM resources that are simple
// global lists (no parent): account alias, OIDC/SAML providers, server
// certificates, virtual MFA devices, and the singleton account password policy.
// Errors are logged, not fatal — a missing password policy (NoSuchEntity) or a
// permission gap on one list shouldn't abort the whole IAM import.
func (g *IamGenerator) getAccountResources(svc *iam.Client) {
	ctx := awsContext()

	if feats, err := svc.ListOrganizationsFeatures(ctx, &iam.ListOrganizationsFeaturesInput{}); err == nil && len(feats.EnabledFeatures) > 0 {
		orgID := StringValue(feats.OrganizationId)
		if orgID != "" {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				orgID, orgID, "aws_iam_organizations_features", "aws", IamAllowEmptyValues))
		}
	}

	aliases := iam.NewListAccountAliasesPaginator(svc, &iam.ListAccountAliasesInput{})
	for aliases.HasMorePages() {
		page, err := aliases.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, alias := range page.AccountAliases {
			if alias == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				alias, alias, "aws_iam_account_alias", "aws", IamAllowEmptyValues))
		}
	}

	if oidc, err := svc.ListOpenIDConnectProviders(ctx, &iam.ListOpenIDConnectProvidersInput{}); err != nil {
		log.Println(err)
	} else {
		for _, p := range oidc.OpenIDConnectProviderList {
			arn := StringValue(p.Arn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_iam_openid_connect_provider", "aws", IamAllowEmptyValues))
		}
	}

	if saml, err := svc.ListSAMLProviders(ctx, &iam.ListSAMLProvidersInput{}); err != nil {
		log.Println(err)
	} else {
		for _, p := range saml.SAMLProviderList {
			arn := StringValue(p.Arn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_iam_saml_provider", "aws", IamAllowEmptyValues))
		}
	}

	certs := iam.NewListServerCertificatesPaginator(svc, &iam.ListServerCertificatesInput{})
	for certs.HasMorePages() {
		page, err := certs.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, c := range page.ServerCertificateMetadataList {
			name := StringValue(c.ServerCertificateName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_iam_server_certificate", "aws", IamAllowEmptyValues))
		}
	}

	mfa := iam.NewListVirtualMFADevicesPaginator(svc, &iam.ListVirtualMFADevicesInput{})
	for mfa.HasMorePages() {
		page, err := mfa.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, d := range page.VirtualMFADevices {
			arn := StringValue(d.SerialNumber)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_iam_virtual_mfa_device", "aws", IamAllowEmptyValues))
		}
	}

	// Singleton: present only if a custom password policy is set (else NoSuchEntity).
	if _, err := svc.GetAccountPasswordPolicy(ctx, &iam.GetAccountPasswordPolicyInput{}); err == nil {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			"iam-account-password-policy", "iam-account-password-policy",
			"aws_iam_account_password_policy", "aws", IamAllowEmptyValues))
	}
}

func (g *IamGenerator) getRoles(svc *iam.Client) error {
	p := iam.NewListRolesPaginator(svc, &iam.ListRolesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, role := range page.Roles {
			roleName := StringValue(role.RoleName)
			// Service-linked roles can't be managed as aws_iam_role; they map
			// to aws_iam_service_linked_role (import by ARN).
			if isServiceLinkedRolePath(StringValue(role.Path)) {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					StringValue(role.Arn), roleName, "aws_iam_service_linked_role", "aws", IamAllowEmptyValues))
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				roleName,
				roleName,
				"aws_iam_role",
				"aws",
				IamAllowEmptyValues))
			rolePoliciesPage := iam.NewListRolePoliciesPaginator(svc, &iam.ListRolePoliciesInput{RoleName: role.RoleName})
			for rolePoliciesPage.HasMorePages() {
				rolePoliciesNextPage, err := rolePoliciesPage.NextPage(awsContext())
				if err != nil {
					log.Println(err)
					continue
				}
				for _, policyName := range rolePoliciesNextPage.PolicyNames {
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						roleName+":"+policyName,
						roleName+"_"+policyName,
						"aws_iam_role_policy",
						"aws",
						IamAllowEmptyValues))
				}
			}
			roleAttachedPoliciesPage := iam.NewListAttachedRolePoliciesPaginator(svc, &iam.ListAttachedRolePoliciesInput{
				RoleName: &roleName,
			})
			for roleAttachedPoliciesPage.HasMorePages() {
				roleAttachedPoliciesNextPage, err := roleAttachedPoliciesPage.NextPage(awsContext())
				if err != nil {
					log.Println(err)
					continue
				}
				for _, attachedPolicy := range roleAttachedPoliciesNextPage.AttachedPolicies {
					g.Resources = append(g.Resources, terraformutils.NewResource(
						roleName+"/"+*attachedPolicy.PolicyArn,
						roleName+"_"+*attachedPolicy.PolicyName,
						"aws_iam_role_policy_attachment",
						"aws",
						map[string]string{
							"role":       roleName,
							"policy_arn": *attachedPolicy.PolicyArn,
						},
						IamAllowEmptyValues,
						map[string]interface{}{}))
				}
			}
		}
	}
	return nil
}

func (g *IamGenerator) getUsers(svc *iam.Client) error {
	p := iam.NewListUsersPaginator(svc, &iam.ListUsersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, user := range page.Users {
			resourceName := StringValue(user.UserName)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				resourceName,
				StringValue(user.UserId),
				"aws_iam_user",
				"aws",
				map[string]string{
					"force_destroy": "false",
				},
				IamAllowEmptyValues,
				map[string]interface{}{}))
			err := g.getUserPolices(svc, user.UserName)
			if err != nil {
				log.Println(err)
			}
			err = g.getUserPolicyAttachment(svc, user.UserName)
			if err != nil {
				log.Println(err)
			}
			err = g.getUserGroup(svc, user.UserName)
			if err != nil {
				log.Println(err)
			}
			err = g.getUserAccessKey(svc, user.UserName, StringValue(user.UserId))
			if err != nil {
				log.Println(err)
			}
			g.getUserCredentials(svc, StringValue(user.UserName))
		}
	}
	return nil
}

// getUserCredentials enumerates a user's SSH public keys, signing certificates,
// and service-specific credentials. Composite import IDs per the aws provider.
func (g *IamGenerator) getUserCredentials(svc *iam.Client, userName string) {
	if userName == "" {
		return
	}
	ctx := awsContext()
	if out, err := svc.ListSSHPublicKeys(ctx, &iam.ListSSHPublicKeysInput{UserName: &userName}); err == nil {
		for _, k := range out.SSHPublicKeys {
			id := StringValue(k.SSHPublicKeyId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				userName+":"+id+":SSH", userName+"_"+id, "aws_iam_user_ssh_key", "aws", IamAllowEmptyValues))
		}
	}
	if out, err := svc.ListSigningCertificates(ctx, &iam.ListSigningCertificatesInput{UserName: &userName}); err == nil {
		for _, c := range out.Certificates {
			id := StringValue(c.CertificateId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id+":"+userName, userName+"_"+id, "aws_iam_signing_certificate", "aws", IamAllowEmptyValues))
		}
	}
	if out, err := svc.ListServiceSpecificCredentials(ctx, &iam.ListServiceSpecificCredentialsInput{UserName: &userName}); err == nil {
		for _, c := range out.ServiceSpecificCredentials {
			id := StringValue(c.ServiceSpecificCredentialId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(c.ServiceName)+":"+userName+":"+id, userName+"_"+id,
				"aws_iam_service_specific_credential", "aws", IamAllowEmptyValues))
		}
	}
}

func (g *IamGenerator) getUserGroup(svc *iam.Client, userName *string) error {
	p := iam.NewListGroupsForUserPaginator(svc, &iam.ListGroupsForUserInput{UserName: userName})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, group := range page.Groups {
			userGroupMembership := *userName + "/" + *group.GroupName
			g.Resources = append(g.Resources, terraformutils.NewResource(
				userGroupMembership,
				userGroupMembership,
				"aws_iam_user_group_membership",
				"aws",
				map[string]string{
					"user":     *userName,
					"groups.#": "1",
					"groups.0": *group.GroupName,
				},
				IamAllowEmptyValues,
				IamAdditionalFields,
			))
		}
	}
	return nil
}

func (g *IamGenerator) getUserPolices(svc *iam.Client, userName *string) error {
	p := iam.NewListUserPoliciesPaginator(svc, &iam.ListUserPoliciesInput{UserName: userName})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, policy := range page.PolicyNames {
			resourceName := StringValue(userName) + "_" + policy
			resourceName = strings.ReplaceAll(resourceName, "@", "")
			policyID := StringValue(userName) + ":" + policy
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				policyID,
				resourceName,
				"aws_iam_user_policy",
				"aws",
				IamAllowEmptyValues))
		}
	}
	return nil
}

func (g *IamGenerator) getUserPolicyAttachment(svc *iam.Client, userName *string) error {
	p := iam.NewListAttachedUserPoliciesPaginator(svc, &iam.ListAttachedUserPoliciesInput{
		UserName: userName,
	})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, attachedPolicy := range page.AttachedPolicies {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				*userName+"/"+*attachedPolicy.PolicyArn,
				*userName+"_"+*attachedPolicy.PolicyName,
				"aws_iam_user_policy_attachment",
				"aws",
				map[string]string{
					"user":       *userName,
					"policy_arn": *attachedPolicy.PolicyArn,
				},
				IamAllowEmptyValues,
				map[string]interface{}{}))
		}
	}
	return nil
}

func (g *IamGenerator) getPolicies(svc *iam.Client) error {
	p := iam.NewListPoliciesPaginator(svc, &iam.ListPoliciesInput{Scope: types.PolicyScopeTypeLocal})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, policy := range page.Policies {
			resourceName := StringValue(policy.PolicyName)
			policyARN := StringValue(policy.Arn)

			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				policyARN,
				resourceName,
				"aws_iam_policy",
				"aws",
				IamAllowEmptyValues))
		}
	}
	return nil
}

func (g *IamGenerator) getGroups(svc *iam.Client) error {
	p := iam.NewListGroupsPaginator(svc, &iam.ListGroupsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, group := range page.Groups {
			resourceName := StringValue(group.GroupName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				resourceName,
				resourceName,
				"aws_iam_group",
				"aws",
				IamAllowEmptyValues))
			g.getGroupPolicies(svc, group)
			g.getAttachedGroupPolicies(svc, group)
		}
	}
	return nil
}

func (g *IamGenerator) getGroupPolicies(svc *iam.Client, group types.Group) {
	groupPoliciesPage := iam.NewListGroupPoliciesPaginator(svc, &iam.ListGroupPoliciesInput{GroupName: group.GroupName})
	for groupPoliciesPage.HasMorePages() {
		groupPoliciesNextPage, err := groupPoliciesPage.NextPage(awsContext())
		if err != nil {
			log.Println(err)
			continue
		}
		for _, policy := range groupPoliciesNextPage.PolicyNames {
			id := *group.GroupName + ":" + policy
			groupPolicyName := *group.GroupName + "_" + policy
			g.Resources = append(g.Resources, terraformutils.NewResource(
				id,
				groupPolicyName,
				"aws_iam_group_policy",
				"aws",
				map[string]string{},
				IamAllowEmptyValues,
				IamAdditionalFields))
		}
	}
}

func (g *IamGenerator) getAttachedGroupPolicies(svc *iam.Client, group types.Group) {
	groupAttachedPoliciesPage := iam.NewListAttachedGroupPoliciesPaginator(svc,
		&iam.ListAttachedGroupPoliciesInput{GroupName: group.GroupName})
	for groupAttachedPoliciesPage.HasMorePages() {
		groupAttachedPoliciesNextPage, err := groupAttachedPoliciesPage.NextPage(awsContext())
		if err != nil {
			log.Println(err)
			continue
		}
		for _, attachedPolicy := range groupAttachedPoliciesNextPage.AttachedPolicies {
			if !strings.Contains(*attachedPolicy.PolicyArn, "arn:aws:iam::aws") {
				continue // map only AWS managed policies since others should be managed by
			}
			id := *group.GroupName + "/" + *attachedPolicy.PolicyArn
			g.Resources = append(g.Resources, terraformutils.NewResource(
				id,
				*group.GroupName+"_"+*attachedPolicy.PolicyName,
				"aws_iam_group_policy_attachment",
				"aws",
				map[string]string{
					"group":      *group.GroupName,
					"policy_arn": *attachedPolicy.PolicyArn,
				},
				IamAllowEmptyValues,
				IamAdditionalFields))
		}
	}
}

func (g *IamGenerator) getInstanceProfiles(svc *iam.Client) error {
	p := iam.NewListInstanceProfilesPaginator(svc, &iam.ListInstanceProfilesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, instanceProfile := range page.InstanceProfiles {
			resourceName := *instanceProfile.InstanceProfileName

			g.Resources = append(g.Resources, terraformutils.NewResource(
				resourceName,
				resourceName,
				"aws_iam_instance_profile",
				"aws",
				map[string]string{
					"name": resourceName,
				},
				IamAllowEmptyValues,
				IamAdditionalFields))
		}
	}
	return nil
}

func (g *IamGenerator) getUserAccessKey(svc *iam.Client, userName *string, userID string) error {
	p := iam.NewListAccessKeysPaginator(svc, &iam.ListAccessKeysInput{UserName: userName})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, key := range page.AccessKeyMetadata {
			accessKeyID := StringValue(key.AccessKeyId)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				accessKeyID,
				accessKeyID,
				"aws_iam_access_key",
				"aws",
				map[string]string{
					"user": *userName,
				},
				IamAllowEmptyValues,
				map[string]interface{}{
					"depends_on": []string{"aws_iam_user.tfer--" + userID},
				}))
		}
	}
	return nil
}

// PostGenerateHook for add policy json as heredoc
func (g *IamGenerator) PostConvertHook() error {
	for i, resource := range g.Resources {
		switch {
		case resource.InstanceInfo.Type == "aws_iam_policy" ||
			resource.InstanceInfo.Type == "aws_iam_user_policy" ||
			resource.InstanceInfo.Type == "aws_iam_group_policy" ||
			resource.InstanceInfo.Type == "aws_iam_role_policy":
			policy := g.escapeAwsInterpolation(resource.Item["policy"].(string))
			resource.Item["policy"] = fmt.Sprintf(`<<POLICY
%s
POLICY`, policy)
		case resource.InstanceInfo.Type == "aws_iam_role":
			policy := g.escapeAwsInterpolation(resource.Item["assume_role_policy"].(string))
			g.Resources[i].Item["assume_role_policy"] = fmt.Sprintf(`<<POLICY
%s
POLICY`, policy)
		case resource.InstanceInfo.Type == "aws_iam_instance_profile":
			delete(resource.Item, "roles")
		}
	}
	return nil
}
