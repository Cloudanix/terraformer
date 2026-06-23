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
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
)

var cloudFrontAllowEmptyValues = []string{"tags."}

type CloudFrontGenerator struct {
	AWSService
}

func (g *CloudFrontGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := cloudfront.NewFromConfig(config)

	if err := g.loadDistribution(svc); err != nil {
		return err
	}

	if err := g.loadCachePolicy(svc); err != nil {
		return err
	}

	if err := g.loadFunctions(svc); err != nil {
		return err
	}

	if err := g.loadOriginAccessControls(svc); err != nil {
		return err
	}

	if err := g.loadOriginAccessIdentities(svc); err != nil {
		return err
	}

	if err := g.loadResponseHeadersPolicies(svc); err != nil {
		return err
	}

	if err := g.loadOriginRequestPolicies(svc); err != nil {
		return err
	}

	if err := g.loadKeyGroups(svc); err != nil {
		return err
	}

	if err := g.loadKeyValueStores(svc); err != nil {
		return err
	}

	if err := g.loadPublicKeys(svc); err != nil {
		return err
	}

	if err := g.loadFieldLevelEncryption(svc); err != nil {
		return err
	}

	if err := g.loadContinuousDeploymentPolicies(svc); err != nil {
		return err
	}

	if err := g.loadRealtimeLogConfigs(svc); err != nil {
		return err
	}

	if err := g.loadVpcOrigins(svc); err != nil {
		return err
	}

	if err := g.loadAnycastIPLists(svc); err != nil {
		return err
	}

	if err := g.loadDistributionTenants(svc); err != nil {
		return err
	}

	if err := g.loadConnectionGroups(svc); err != nil {
		return err
	}

	if err := g.loadTrustStores(svc); err != nil {
		return err
	}

	return nil
}

func (g *CloudFrontGenerator) loadVpcOrigins(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListVpcOrigins(awsContext(), &cloudfront.ListVpcOriginsInput{Marker: marker})
		if err != nil {
			return err
		}
		if out.VpcOriginList == nil {
			return nil
		}
		for _, v := range out.VpcOriginList.Items {
			g.addSimple(StringValue(v.Id), StringValue(v.Name), "aws_cloudfront_vpc_origin")
		}
		marker = out.VpcOriginList.NextMarker
		if marker == nil {
			return nil
		}
	}
}

func (g *CloudFrontGenerator) loadAnycastIPLists(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListAnycastIpLists(awsContext(), &cloudfront.ListAnycastIpListsInput{Marker: marker})
		if err != nil {
			return err
		}
		if out.AnycastIpLists == nil {
			return nil
		}
		for _, a := range out.AnycastIpLists.Items {
			g.addSimple(StringValue(a.Id), StringValue(a.Name), "aws_cloudfront_anycast_ip_list")
		}
		marker = out.AnycastIpLists.NextMarker
		if marker == nil {
			return nil
		}
	}
}

func (g *CloudFrontGenerator) loadDistributionTenants(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListDistributionTenants(awsContext(), &cloudfront.ListDistributionTenantsInput{Marker: marker})
		if err != nil {
			return err
		}
		for _, t := range out.DistributionTenantList {
			g.addSimple(StringValue(t.Id), StringValue(t.Name), "aws_cloudfront_distribution_tenant")
		}
		marker = out.NextMarker
		if marker == nil {
			return nil
		}
	}
}

func (g *CloudFrontGenerator) loadConnectionGroups(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListConnectionGroups(awsContext(), &cloudfront.ListConnectionGroupsInput{Marker: marker})
		if err != nil {
			return err
		}
		for _, c := range out.ConnectionGroups {
			g.addSimple(StringValue(c.Id), StringValue(c.Name), "aws_cloudfront_connection_group")
		}
		marker = out.NextMarker
		if marker == nil {
			return nil
		}
	}
}

func (g *CloudFrontGenerator) loadTrustStores(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListTrustStores(awsContext(), &cloudfront.ListTrustStoresInput{Marker: marker})
		if err != nil {
			return err
		}
		for _, t := range out.TrustStoreList {
			g.addSimple(StringValue(t.Id), StringValue(t.Name), "aws_cloudfront_trust_store")
		}
		marker = out.NextMarker
		if marker == nil {
			return nil
		}
	}
}

func (g *CloudFrontGenerator) loadPublicKeys(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListPublicKeys(awsContext(), &cloudfront.ListPublicKeysInput{Marker: marker})
		if err != nil {
			return err
		}
		if out.PublicKeyList == nil {
			return nil
		}
		for _, k := range out.PublicKeyList.Items {
			g.addSimple(StringValue(k.Id), StringValue(k.Name), "aws_cloudfront_public_key")
		}
		marker = out.PublicKeyList.NextMarker
		if marker == nil {
			return nil
		}
	}
}

func (g *CloudFrontGenerator) loadFieldLevelEncryption(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListFieldLevelEncryptionConfigs(awsContext(), &cloudfront.ListFieldLevelEncryptionConfigsInput{Marker: marker})
		if err != nil {
			return err
		}
		if out.FieldLevelEncryptionList == nil {
			break
		}
		for _, c := range out.FieldLevelEncryptionList.Items {
			id := StringValue(c.Id)
			g.addSimple(id, id, "aws_cloudfront_field_level_encryption_config")
		}
		marker = out.FieldLevelEncryptionList.NextMarker
		if marker == nil {
			break
		}
	}
	marker = nil
	for {
		out, err := svc.ListFieldLevelEncryptionProfiles(awsContext(), &cloudfront.ListFieldLevelEncryptionProfilesInput{Marker: marker})
		if err != nil {
			return err
		}
		if out.FieldLevelEncryptionProfileList == nil {
			return nil
		}
		for _, p := range out.FieldLevelEncryptionProfileList.Items {
			g.addSimple(StringValue(p.Id), StringValue(p.Name), "aws_cloudfront_field_level_encryption_profile")
		}
		marker = out.FieldLevelEncryptionProfileList.NextMarker
		if marker == nil {
			return nil
		}
	}
}

func (g *CloudFrontGenerator) loadContinuousDeploymentPolicies(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListContinuousDeploymentPolicies(awsContext(), &cloudfront.ListContinuousDeploymentPoliciesInput{Marker: marker})
		if err != nil {
			return err
		}
		if out.ContinuousDeploymentPolicyList == nil {
			return nil
		}
		for _, p := range out.ContinuousDeploymentPolicyList.Items {
			if p.ContinuousDeploymentPolicy == nil {
				continue
			}
			id := StringValue(p.ContinuousDeploymentPolicy.Id)
			g.addSimple(id, id, "aws_cloudfront_continuous_deployment_policy")
		}
		marker = out.ContinuousDeploymentPolicyList.NextMarker
		if marker == nil {
			return nil
		}
	}
}

func (g *CloudFrontGenerator) loadRealtimeLogConfigs(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListRealtimeLogConfigs(awsContext(), &cloudfront.ListRealtimeLogConfigsInput{Marker: marker})
		if err != nil {
			return err
		}
		if out.RealtimeLogConfigs == nil {
			return nil
		}
		for _, c := range out.RealtimeLogConfigs.Items {
			g.addSimple(StringValue(c.ARN), StringValue(c.Name), "aws_cloudfront_realtime_log_config")
		}
		marker = out.RealtimeLogConfigs.NextMarker
		if marker == nil {
			return nil
		}
	}
}

func (g *CloudFrontGenerator) loadKeyValueStores(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListKeyValueStores(awsContext(), &cloudfront.ListKeyValueStoresInput{Marker: marker})
		if err != nil {
			return err
		}
		if out.KeyValueStoreList != nil {
			for _, kvs := range out.KeyValueStoreList.Items {
				name := StringValue(kvs.Name)
				g.addSimple(name, name, "aws_cloudfront_key_value_store")
			}
			marker = out.KeyValueStoreList.NextMarker
		}
		if marker == nil {
			return nil
		}
	}
}

func (g *CloudFrontGenerator) addSimple(id, name, tfType string) {
	if id == "" {
		return
	}
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
		id, name, tfType, "aws", cloudFrontAllowEmptyValues))
}

func (g *CloudFrontGenerator) loadFunctions(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListFunctions(awsContext(), &cloudfront.ListFunctionsInput{Marker: marker})
		if err != nil {
			return err
		}
		if out.FunctionList != nil {
			for _, fn := range out.FunctionList.Items {
				name := StringValue(fn.Name)
				g.addSimple(name, name, "aws_cloudfront_function")
			}
			marker = out.FunctionList.NextMarker
		}
		if marker == nil {
			return nil
		}
	}
}

func (g *CloudFrontGenerator) loadOriginAccessControls(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListOriginAccessControls(awsContext(), &cloudfront.ListOriginAccessControlsInput{Marker: marker})
		if err != nil {
			return err
		}
		if out.OriginAccessControlList != nil {
			for _, oac := range out.OriginAccessControlList.Items {
				id := StringValue(oac.Id)
				g.addSimple(id, id, "aws_cloudfront_origin_access_control")
			}
			marker = out.OriginAccessControlList.NextMarker
		}
		if marker == nil {
			return nil
		}
	}
}

func (g *CloudFrontGenerator) loadOriginAccessIdentities(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListCloudFrontOriginAccessIdentities(awsContext(), &cloudfront.ListCloudFrontOriginAccessIdentitiesInput{Marker: marker})
		if err != nil {
			return err
		}
		if out.CloudFrontOriginAccessIdentityList != nil {
			for _, oai := range out.CloudFrontOriginAccessIdentityList.Items {
				id := StringValue(oai.Id)
				g.addSimple(id, id, "aws_cloudfront_origin_access_identity")
			}
			marker = out.CloudFrontOriginAccessIdentityList.NextMarker
		}
		if marker == nil {
			return nil
		}
	}
}

func (g *CloudFrontGenerator) loadResponseHeadersPolicies(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListResponseHeadersPolicies(awsContext(), &cloudfront.ListResponseHeadersPoliciesInput{Marker: marker})
		if err != nil {
			return err
		}
		if out.ResponseHeadersPolicyList != nil {
			for _, p := range out.ResponseHeadersPolicyList.Items {
				if p.ResponseHeadersPolicy == nil {
					continue
				}
				id := StringValue(p.ResponseHeadersPolicy.Id)
				g.addSimple(id, id, "aws_cloudfront_response_headers_policy")
			}
			marker = out.ResponseHeadersPolicyList.NextMarker
		}
		if marker == nil {
			return nil
		}
	}
}

func (g *CloudFrontGenerator) loadOriginRequestPolicies(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListOriginRequestPolicies(awsContext(), &cloudfront.ListOriginRequestPoliciesInput{Marker: marker})
		if err != nil {
			return err
		}
		if out.OriginRequestPolicyList != nil {
			for _, p := range out.OriginRequestPolicyList.Items {
				if p.OriginRequestPolicy == nil {
					continue
				}
				id := StringValue(p.OriginRequestPolicy.Id)
				g.addSimple(id, id, "aws_cloudfront_origin_request_policy")
			}
			marker = out.OriginRequestPolicyList.NextMarker
		}
		if marker == nil {
			return nil
		}
	}
}

func (g *CloudFrontGenerator) loadKeyGroups(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListKeyGroups(awsContext(), &cloudfront.ListKeyGroupsInput{Marker: marker})
		if err != nil {
			return err
		}
		if out.KeyGroupList != nil {
			for _, kg := range out.KeyGroupList.Items {
				if kg.KeyGroup == nil {
					continue
				}
				id := StringValue(kg.KeyGroup.Id)
				g.addSimple(id, id, "aws_cloudfront_key_group")
			}
			marker = out.KeyGroupList.NextMarker
		}
		if marker == nil {
			return nil
		}
	}
}

func (g *CloudFrontGenerator) loadDistribution(svc *cloudfront.Client) error {
	p := cloudfront.NewListDistributionsPaginator(svc, &cloudfront.ListDistributionsInput{})
	for p.HasMorePages() {
		page, e := p.NextPage(awsContext())
		if e != nil {
			return e
		}
		for _, distribution := range page.DistributionList.Items {
			r := terraformutils.NewResource(
				StringValue(distribution.Id),
				StringValue(distribution.Id),
				"aws_cloudfront_distribution",
				"aws",
				map[string]string{
					"retain_on_delete": "false",
				},
				cloudFrontAllowEmptyValues,
				map[string]interface{}{},
			)
			r.IgnoreKeys = append(r.IgnoreKeys, "^active_trusted_signers.(.*)")
			g.Resources = append(g.Resources, r)

			// Realtime metrics subscription is a singleton per distribution.
			if _, err := svc.GetMonitoringSubscription(awsContext(), &cloudfront.GetMonitoringSubscriptionInput{DistributionId: distribution.Id}); err == nil {
				g.addSimple(StringValue(distribution.Id), StringValue(distribution.Id), "aws_cloudfront_monitoring_subscription")
			}
		}
	}
	return nil
}

func (g *CloudFrontGenerator) loadCachePolicy(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListCachePolicies(awsContext(), &cloudfront.ListCachePoliciesInput{
			Marker: marker,
		})
		if err != nil {
			return err
		}
		for _, cachePolicy := range out.CachePolicyList.Items {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(cachePolicy.CachePolicy.Id),
				StringValue(cachePolicy.CachePolicy.Id),
				"aws_cloudfront_cache_policy",
				"aws",
				cloudFrontAllowEmptyValues,
			))
		}
		marker = out.CachePolicyList.NextMarker
		if marker == nil {
			break
		}
	}
	return nil
}

func (g *CloudFrontGenerator) PostConvertHook() error {
	for i, r := range g.Resources {
		if r.InstanceInfo.Type != "aws_cloudfront_distribution" {
			continue
		}

		for _, cachePolicy := range g.Resources {
			if cachePolicy.InstanceInfo.Type != "aws_cloudfront_cache_policy" {
				continue
			}

			if defaultCacheBehavior, ok := r.Item["default_cache_behavior"].([]interface{})[0].(map[string]interface{})["cache_policy_id"]; ok {
				if defaultCacheBehavior.(string) == cachePolicy.InstanceState.Attributes["id"] {
					g.Resources[i].Item["default_cache_behavior"].([]interface{})[0].(map[string]interface{})["cache_policy_id"] = "${aws_cloudfront_cache_policy." + cachePolicy.ResourceName + ".id}"
				}
			}

			if orderedCacheBehavior, ok := r.Item["ordered_cache_behavior"].([]interface{}); ok {
				for j, behavior := range orderedCacheBehavior {
					if behavior, ok := behavior.(map[string]interface{})["cache_policy_id"]; ok && behavior.(string) == cachePolicy.InstanceState.Attributes["id"] {
						g.Resources[i].Item["ordered_cache_behavior"].([]interface{})[j].(map[string]interface{})["cache_policy_id"] = "${aws_cloudfront_cache_policy." + cachePolicy.ResourceName + ".id}"
					}
				}
			}
		}

	}
	return nil
}
