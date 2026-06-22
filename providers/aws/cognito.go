package aws

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentity"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
)

var CognitoAllowEmptyValues = []string{"tags."}

var CognitoAdditionalFields = map[string]interface{}{}

type CognitoGenerator struct {
	AWSService
}

const CognitoMaxResults = 60 // Required field for Cognito API

func (g *CognitoGenerator) loadIdentityPools(svc *cognitoidentity.Client) error {
	p := cognitoidentity.NewListIdentityPoolsPaginator(svc, &cognitoidentity.ListIdentityPoolsInput{
		MaxResults: aws.Int32(CognitoMaxResults),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, pool := range page.IdentityPools {
			var id = *pool.IdentityPoolId
			var resourceName = *pool.IdentityPoolName
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				resourceName+"_"+id,
				"aws_cognito_identity_pool",
				"aws",
				[]string{}))
			poolID := id
			if roles, err := svc.GetIdentityPoolRoles(context.TODO(), &cognitoidentity.GetIdentityPoolRolesInput{IdentityPoolId: &poolID}); err == nil && len(roles.Roles) > 0 {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					poolID, poolID, "aws_cognito_identity_pool_roles_attachment", "aws", []string{}))
			}
		}
	}

	return nil
}

func (g *CognitoGenerator) loadUserPools(svc *cognitoidentityprovider.Client) ([]string, error) {
	p := cognitoidentityprovider.NewListUserPoolsPaginator(svc, &cognitoidentityprovider.ListUserPoolsInput{
		MaxResults: aws.Int32(CognitoMaxResults),
	})

	var userPoolIds []string
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}
		for _, pool := range page.UserPools {
			id := *pool.Id
			resourceName := *pool.Name
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				resourceName+"_"+id,
				"aws_cognito_user_pool",
				"aws",
				[]string{}))

			userPoolIds = append(userPoolIds, *pool.Id)
		}
	}
	return userPoolIds, nil
}

func (g *CognitoGenerator) loadUserPoolClients(svc *cognitoidentityprovider.Client, userPoolIds []string) error {
	for _, userPoolID := range userPoolIds {
		p := cognitoidentityprovider.NewListUserPoolClientsPaginator(svc, &cognitoidentityprovider.ListUserPoolClientsInput{
			UserPoolId: aws.String(userPoolID),
			MaxResults: aws.Int32(CognitoMaxResults),
		})

		for p.HasMorePages() {
			page, err := p.NextPage(context.TODO())
			if err != nil {
				return err
			}
			for _, poolClient := range page.UserPoolClients {
				id := *poolClient.ClientId
				resourceName := *poolClient.ClientName
				g.Resources = append(g.Resources, terraformutils.NewResource(
					id,
					resourceName+"_"+id,
					"aws_cognito_user_pool_client",
					"aws",
					map[string]string{
						"user_pool_id": *poolClient.UserPoolId,
					},
					CognitoAllowEmptyValues,
					CognitoAdditionalFields))
			}
		}
	}
	return nil
}

func (g *CognitoGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}

	svcCognitoIdentity := cognitoidentity.NewFromConfig(config)
	if err := g.loadIdentityPools(svcCognitoIdentity); err != nil {
		return err
	}
	svcCognitoIdentityProvider := cognitoidentityprovider.NewFromConfig(config)

	userPoolIds, err := g.loadUserPools(svcCognitoIdentityProvider)
	if err != nil {
		return err
	}
	if err = g.loadUserPoolClients(svcCognitoIdentityProvider, userPoolIds); err != nil {
		return err
	}
	if err = g.loadUserPoolChildren(svcCognitoIdentityProvider, userPoolIds); err != nil {
		return err
	}

	return nil
}

// loadUserPoolChildren enumerates per-user-pool groups, resource servers, and
// identity providers. Import IDs:
//   - aws_cognito_user_group         → "<user_pool_id>/<group_name>"
//   - aws_cognito_resource_server    → "<user_pool_id>|<identifier>"
//   - aws_cognito_identity_provider  → "<user_pool_id>:<provider_name>"
func (g *CognitoGenerator) loadUserPoolChildren(svc *cognitoidentityprovider.Client, userPoolIds []string) error {
	for _, userPoolID := range userPoolIds {
		if desc, err := svc.DescribeUserPool(context.TODO(), &cognitoidentityprovider.DescribeUserPoolInput{UserPoolId: aws.String(userPoolID)}); err == nil && desc.UserPool != nil {
			domain := StringValue(desc.UserPool.Domain)
			if domain != "" {
				g.Resources = append(g.Resources, terraformutils.NewResource(
					domain, domain, "aws_cognito_user_pool_domain", "aws",
					map[string]string{"user_pool_id": userPoolID}, CognitoAllowEmptyValues, CognitoAdditionalFields))
			}
		}
		groups := cognitoidentityprovider.NewListGroupsPaginator(svc, &cognitoidentityprovider.ListGroupsInput{
			UserPoolId: aws.String(userPoolID), Limit: aws.Int32(CognitoMaxResults),
		})
		for groups.HasMorePages() {
			page, err := groups.NextPage(context.TODO())
			if err != nil {
				return err
			}
			for _, group := range page.Groups {
				name := StringValue(group.GroupName)
				if name == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewResource(
					userPoolID+"/"+name, userPoolID+"_"+name, "aws_cognito_user_group", "aws",
					map[string]string{"user_pool_id": userPoolID}, CognitoAllowEmptyValues, CognitoAdditionalFields))
				for ug := cognitoidentityprovider.NewListUsersInGroupPaginator(svc, &cognitoidentityprovider.ListUsersInGroupInput{
					UserPoolId: aws.String(userPoolID), GroupName: aws.String(name),
				}); ug.HasMorePages(); {
					ugPage, err := ug.NextPage(context.TODO())
					if err != nil {
						break
					}
					for _, u := range ugPage.Users {
						username := StringValue(u.Username)
						if username == "" {
							continue
						}
						g.Resources = append(g.Resources, terraformutils.NewResource(
							userPoolID+"/"+name+"/"+username, userPoolID+"_"+name+"_"+username, "aws_cognito_user_in_group", "aws",
							map[string]string{"user_pool_id": userPoolID}, CognitoAllowEmptyValues, CognitoAdditionalFields))
					}
				}
			}
		}

		servers := cognitoidentityprovider.NewListResourceServersPaginator(svc, &cognitoidentityprovider.ListResourceServersInput{
			UserPoolId: aws.String(userPoolID), MaxResults: aws.Int32(CognitoMaxResults),
		})
		for servers.HasMorePages() {
			page, err := servers.NextPage(context.TODO())
			if err != nil {
				return err
			}
			for _, server := range page.ResourceServers {
				identifier := StringValue(server.Identifier)
				if identifier == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewResource(
					userPoolID+"|"+identifier, userPoolID+"_"+identifier, "aws_cognito_resource_server", "aws",
					map[string]string{"user_pool_id": userPoolID}, CognitoAllowEmptyValues, CognitoAdditionalFields))
			}
		}

		providers := cognitoidentityprovider.NewListIdentityProvidersPaginator(svc, &cognitoidentityprovider.ListIdentityProvidersInput{
			UserPoolId: aws.String(userPoolID), MaxResults: aws.Int32(CognitoMaxResults),
		})
		for providers.HasMorePages() {
			page, err := providers.NextPage(context.TODO())
			if err != nil {
				return err
			}
			for _, provider := range page.Providers {
				name := StringValue(provider.ProviderName)
				if name == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewResource(
					userPoolID+":"+name, userPoolID+"_"+name, "aws_cognito_identity_provider", "aws",
					map[string]string{"user_pool_id": userPoolID}, CognitoAllowEmptyValues, CognitoAdditionalFields))
			}
		}

		users := cognitoidentityprovider.NewListUsersPaginator(svc, &cognitoidentityprovider.ListUsersInput{
			UserPoolId: aws.String(userPoolID),
		})
		for users.HasMorePages() {
			page, err := users.NextPage(context.TODO())
			if err != nil {
				return err
			}
			for _, user := range page.Users {
				username := StringValue(user.Username)
				if username == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewResource(
					userPoolID+"/"+username, userPoolID+"_"+username, "aws_cognito_user", "aws",
					map[string]string{"user_pool_id": userPoolID}, CognitoAllowEmptyValues, CognitoAdditionalFields))
			}
		}

		if ui, err := svc.GetUICustomization(context.TODO(), &cognitoidentityprovider.GetUICustomizationInput{
			UserPoolId: aws.String(userPoolID),
		}); err == nil && ui.UICustomization != nil && StringValue(ui.UICustomization.CSS) != "" {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				userPoolID, userPoolID, "aws_cognito_user_pool_ui_customization", "aws",
				map[string]string{"user_pool_id": userPoolID}, CognitoAllowEmptyValues, CognitoAdditionalFields))
		}

		if rc, err := svc.DescribeRiskConfiguration(context.TODO(), &cognitoidentityprovider.DescribeRiskConfigurationInput{
			UserPoolId: aws.String(userPoolID),
		}); err == nil && rc.RiskConfiguration != nil &&
			(rc.RiskConfiguration.AccountTakeoverRiskConfiguration != nil ||
				rc.RiskConfiguration.CompromisedCredentialsRiskConfiguration != nil ||
				rc.RiskConfiguration.RiskExceptionConfiguration != nil) {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				userPoolID, userPoolID, "aws_cognito_risk_configuration", "aws",
				map[string]string{"user_pool_id": userPoolID}, CognitoAllowEmptyValues, CognitoAdditionalFields))
		}
	}
	return nil
}

func (g *CognitoGenerator) PostConvertHook() error {
	for _, r := range g.Resources {
		if r.InstanceInfo.Type != "aws_cognito_user_pool" {
			continue
		}
		if _, ok := r.InstanceState.Attributes["admin_create_user_config.0.unused_account_validity_days"]; ok {
			if _, okpp := r.InstanceState.Attributes["admin_create_user_config.0.unused_account_validity_days"]; okpp {
				delete(r.Item["admin_create_user_config"].([]interface{})[0].(map[string]interface{}), "unused_account_validity_days")
			}
		}
		if _, ok := r.InstanceState.Attributes["sms_verification_message"]; ok {
			if _, oktmp := r.InstanceState.Attributes["verification_message_template.0.sms_message"]; oktmp {
				delete(r.Item, "sms_verification_message")
			}
		}
		if _, ok := r.InstanceState.Attributes["email_verification_message"]; ok {
			if _, oktmp := r.InstanceState.Attributes["verification_message_template.0.email_message"]; oktmp {
				delete(r.Item, "email_verification_message")
			}
		}
		if _, ok := r.InstanceState.Attributes["email_verification_subject"]; ok {
			if _, oktmp := r.InstanceState.Attributes["verification_message_template.0.email_subject"]; oktmp {
				delete(r.Item, "email_verification_subject")
			}
		}
	}
	return nil
}
