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
	"fmt"
	"log"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var S3AllowEmptyValues = []string{"tags."}

var S3AdditionalFields = map[string]interface{}{}

type S3Generator struct {
	AWSService
}

// createResources iterate on all buckets
// for each bucket we check region and choose only bucket from set region
// for each bucket try get bucket policy, if policy exist create additional NewTerraformResource for policy
func (g *S3Generator) createResources(config aws.Config, buckets *s3.ListBucketsOutput, region string) []terraformutils.Resource {
	var resources []terraformutils.Resource
	svc := s3.NewFromConfig(config)
	for _, bucket := range buckets.Buckets {
		resourceName := StringValue(bucket.Name)
		location, err := svc.GetBucketLocation(context.TODO(), &s3.GetBucketLocationInput{Bucket: bucket.Name})
		if err != nil {
			log.Println(err)
			continue
		}
		// check if bucket in region
		constraintString := string(location.LocationConstraint)
		if constraintString == region || (constraintString == "" && region == "us-east-1") {
			attributes := map[string]string{
				"force_destroy": "false",
				"acl":           "private",
			}
			// try get policy
			var policy *s3.GetBucketPolicyOutput
			policy, err = svc.GetBucketPolicy(context.TODO(), &s3.GetBucketPolicyInput{
				Bucket: bucket.Name,
			})

			if err == nil && policy.Policy != nil {
				attributes["policy"] = *policy.Policy
				resources = append(resources, terraformutils.NewResource(
					resourceName,
					resourceName,
					"aws_s3_bucket_policy",
					"aws",
					nil,
					S3AllowEmptyValues,
					S3AdditionalFields))
			}
			resources = append(resources, terraformutils.NewResource(
				resourceName,
				resourceName,
				"aws_s3_bucket",
				"aws",
				attributes,
				S3AllowEmptyValues,
				S3AdditionalFields))
			resources = append(resources, g.bucketSubresources(svc, resourceName)...)
		}
	}
	return resources
}

// bucketSubresources probes each split-out S3 bucket configuration and emits a
// resource only where the configuration actually exists — mirroring how the
// bucket policy is handled above. All these resources import by bucket name.
// Most "Get*" calls error (NoSuch*Configuration) when unset, which we treat as
// "not configured" and skip.
func (g *S3Generator) bucketSubresources(svc *s3.Client, bucket string) []terraformutils.Resource {
	ctx := context.TODO()
	var resources []terraformutils.Resource
	add := func(tfType string) {
		resources = append(resources, terraformutils.NewSimpleResource(
			bucket, bucket+"_"+tfType, tfType, "aws", S3AllowEmptyValues))
	}

	if out, err := svc.GetBucketVersioning(ctx, &s3.GetBucketVersioningInput{Bucket: &bucket}); err == nil && out.Status != "" {
		add("aws_s3_bucket_versioning")
	}
	if out, err := svc.GetBucketLifecycleConfiguration(ctx, &s3.GetBucketLifecycleConfigurationInput{Bucket: &bucket}); err == nil && len(out.Rules) > 0 {
		add("aws_s3_bucket_lifecycle_configuration")
	}
	if _, err := svc.GetBucketEncryption(ctx, &s3.GetBucketEncryptionInput{Bucket: &bucket}); err == nil {
		add("aws_s3_bucket_server_side_encryption_configuration")
	}
	if _, err := svc.GetPublicAccessBlock(ctx, &s3.GetPublicAccessBlockInput{Bucket: &bucket}); err == nil {
		add("aws_s3_bucket_public_access_block")
	}
	if out, err := svc.GetBucketCors(ctx, &s3.GetBucketCorsInput{Bucket: &bucket}); err == nil && len(out.CORSRules) > 0 {
		add("aws_s3_bucket_cors_configuration")
	}
	if out, err := svc.GetBucketLogging(ctx, &s3.GetBucketLoggingInput{Bucket: &bucket}); err == nil && out.LoggingEnabled != nil {
		add("aws_s3_bucket_logging")
	}
	if _, err := svc.GetBucketWebsite(ctx, &s3.GetBucketWebsiteInput{Bucket: &bucket}); err == nil {
		add("aws_s3_bucket_website_configuration")
	}
	if out, err := svc.GetBucketOwnershipControls(ctx, &s3.GetBucketOwnershipControlsInput{Bucket: &bucket}); err == nil && out.OwnershipControls != nil {
		add("aws_s3_bucket_ownership_controls")
	}
	if out, err := svc.GetObjectLockConfiguration(ctx, &s3.GetObjectLockConfigurationInput{Bucket: &bucket}); err == nil && out.ObjectLockConfiguration != nil {
		add("aws_s3_bucket_object_lock_configuration")
	}
	if _, err := svc.GetBucketReplication(ctx, &s3.GetBucketReplicationInput{Bucket: &bucket}); err == nil {
		add("aws_s3_bucket_replication_configuration")
	}
	if out, err := svc.GetBucketAccelerateConfiguration(ctx, &s3.GetBucketAccelerateConfigurationInput{Bucket: &bucket}); err == nil && out.Status != "" {
		add("aws_s3_bucket_accelerate_configuration")
	}
	if out, err := svc.GetBucketNotificationConfiguration(ctx, &s3.GetBucketNotificationConfigurationInput{Bucket: &bucket}); err == nil &&
		(out.EventBridgeConfiguration != nil || len(out.LambdaFunctionConfigurations) > 0 || len(out.QueueConfigurations) > 0 || len(out.TopicConfigurations) > 0) {
		add("aws_s3_bucket_notification")
	}
	if out, err := svc.GetBucketRequestPayment(ctx, &s3.GetBucketRequestPaymentInput{Bucket: &bucket}); err == nil && out.Payer == "Requester" {
		add("aws_s3_bucket_request_payment_configuration")
	}
	if _, err := svc.GetBucketAcl(ctx, &s3.GetBucketAclInput{Bucket: &bucket}); err == nil {
		add("aws_s3_bucket_acl")
	}

	// Named, listable per-bucket configurations; import ID is "<bucket>:<id>".
	addNamed := func(id, tfType string) {
		if id != "" {
			resources = append(resources, terraformutils.NewSimpleResource(
				bucket+":"+id, bucket+"_"+id, tfType, "aws", S3AllowEmptyValues))
		}
	}
	if out, err := svc.ListBucketAnalyticsConfigurations(ctx, &s3.ListBucketAnalyticsConfigurationsInput{Bucket: &bucket}); err == nil {
		for _, c := range out.AnalyticsConfigurationList {
			addNamed(StringValue(c.Id), "aws_s3_bucket_analytics_configuration")
		}
	}
	if out, err := svc.ListBucketMetricsConfigurations(ctx, &s3.ListBucketMetricsConfigurationsInput{Bucket: &bucket}); err == nil {
		for _, c := range out.MetricsConfigurationList {
			addNamed(StringValue(c.Id), "aws_s3_bucket_metric")
		}
	}
	if out, err := svc.ListBucketInventoryConfigurations(ctx, &s3.ListBucketInventoryConfigurationsInput{Bucket: &bucket}); err == nil {
		for _, c := range out.InventoryConfigurationList {
			addNamed(StringValue(c.Id), "aws_s3_bucket_inventory")
		}
	}
	if out, err := svc.ListBucketIntelligentTieringConfigurations(ctx, &s3.ListBucketIntelligentTieringConfigurationsInput{Bucket: &bucket}); err == nil {
		for _, c := range out.IntelligentTieringConfigurationList {
			addNamed(StringValue(c.Id), "aws_s3_bucket_intelligent_tiering_configuration")
		}
	}
	return resources
}

// Generate TerraformResources from AWS API,
// Need bucket name as ID for terraform resource
func (g *S3Generator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := s3.NewFromConfig(config)

	buckets, err := svc.ListBuckets(context.TODO(), nil)
	if err != nil {
		return err
	}
	g.Resources = g.createResources(config, buckets, g.GetArgs()["region"].(string))

	// S3 Express directory buckets (separate API).
	for p := s3.NewListDirectoryBucketsPaginator(svc, &s3.ListDirectoryBucketsInput{}); p.HasMorePages(); {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			break
		}
		for _, b := range page.Buckets {
			name := StringValue(b.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_s3_directory_bucket", "aws", S3AllowEmptyValues))
		}
	}
	return nil
}

// PostGenerateHook for add bucket policy json as heredoc
// support only bucket with policy
func (g *S3Generator) PostConvertHook() error {
	for i, resource := range g.Resources {
		if resource.InstanceInfo.Type == "aws_s3_bucket" {
			if val, ok := g.Resources[i].Item["acl"]; ok && val == "private" {
				delete(g.Resources[i].Item, "acl")
			}
			if val, ok := g.Resources[i].Item["policy"]; ok {
				g.Resources[i].Item["policy"] = fmt.Sprintf(`<<POLICY
%s
POLICY`, g.escapeAwsInterpolation(val.(string)))
			}
		}
	}
	return nil
}
