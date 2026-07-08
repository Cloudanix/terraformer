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
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type AWSService struct { //nolint
	terraformutils.Service
}

// awsImportCtx holds the current import run's context (deadline + Ctrl-C) so the
// AWS generators — including the many receiver-less helper functions — can pass
// it to their SDK calls via awsContext(), instead of context.TODO(). Safe as a
// package var because per-service discovery (InitResources) runs sequentially
// in cmd/import.go's initAllServicesResources, and SetContext is called once per
// service immediately before it. ponytail: a package global beats threading ctx
// through 700+ existing call sites and free funcs; revisit only if InitResources
// ever becomes concurrent (then store on the generator and use g.Context()).
var awsImportCtx context.Context

// awsContext returns the import-run context, or context.Background() when unset
// (plan replay, tests, or before SetContext). Generators call this instead of
// context.TODO() so a hung SDK call honours --timeout / Ctrl-C.
func awsContext() context.Context {
	if awsImportCtx != nil {
		return awsImportCtx
	}
	return context.Background()
}

// SetContext captures the import-run context for awsContext() in addition to the
// embedded Service's own storage.
func (s *AWSService) SetContext(ctx context.Context) {
	awsImportCtx = ctx
	s.Service.SetContext(ctx)
}

var awsVariable = regexp.MustCompile(`(\${[0-9A-Za-z:]+})`)

// defaultAllowEmptyValues lists attribute prefixes that may legitimately be
// empty without being dropped from generated HCL. Almost every AWS resource
// carries tags, so "tags." is the universal default. Generators use this shared
// var instead of redeclaring a per-file `var xAllowEmptyValues`.
var defaultAllowEmptyValues = []string{"tags."}

// appendSimpleResources appends one terraformutils.Resource per item to dst,
// deriving the Terraform import ID and resource name via the id/name
// extractors. Items whose import ID is empty are skipped (un-importable).
//
// It centralizes the nil-safe page→item→NewSimpleResource loop so the AWS
// service generators don't copy-paste it (and its bugs). Mirrors append()
// semantics — dst is returned with the new resources appended:
//
//	g.Resources = appendSimpleResources(g.Resources, page.Things, "aws_"+"thing",
//	    defaultAllowEmptyValues,
//	    func(t types.Thing) string { return aws.ToString(t.Id) },
//	    func(t types.Thing) string { return aws.ToString(t.Name) })
//
// Extractors must be nil-safe (use aws.ToString, never *t.Field) — a nil
// pointer deref here would kill the whole import run.
func appendSimpleResources[T any](dst []terraformutils.Resource, items []T,
	resourceType string, allowEmptyValues []string, id, name func(T) string,
) []terraformutils.Resource {
	for _, item := range items {
		resourceID := id(item)
		if resourceID == "" {
			continue
		}
		dst = append(dst, terraformutils.NewSimpleResource(
			resourceID, name(item), resourceType, "aws", allowEmptyValues))
	}
	return dst
}

// wrapPolicyAttribute rewrites attr on every resource whose type is in types
// into a heredoc-wrapped, interpolation-escaped policy document — the form
// terraform-provider-aws expects for inline JSON policies. Resources lacking
// attr, or whose attr is not a string, are left untouched. Factored out of the
// per-generator PostConvertHook duplication (see ecr.go).
func (s *AWSService) wrapPolicyAttribute(resources []terraformutils.Resource, attr string, types ...string) {
	typeSet := make(map[string]bool, len(types))
	for _, t := range types {
		typeSet[t] = true
	}
	for i := range resources {
		if !typeSet[resources[i].InstanceInfo.Type] {
			continue
		}
		val, ok := resources[i].Item[attr]
		if !ok {
			continue
		}
		str, ok := val.(string)
		if !ok {
			continue
		}
		resources[i].Item[attr] = fmt.Sprintf("<<POLICY\n%s\nPOLICY", s.escapeAwsInterpolation(str))
	}
}

// configCache is keyed by region: terraformer imports global resources first
// (region "aws-global") then loops the real regions in the same process. A single
// shared config would freeze every later region to the first pass's endpoint,
// causing wrong-region signing / aws-global DNS failures. Cache per region instead.
var (
	configCache   = map[string]*aws.Config{}
	configCacheMu sync.Mutex
)

func (s *AWSService) generateConfig() (aws.Config, error) {
	region, _ := s.GetArgs()["region"].(string)

	configCacheMu.Lock()
	defer configCacheMu.Unlock()
	if c, ok := configCache[region]; ok {
		return *c, nil
	}

	baseConfig, e := s.buildBaseConfig()

	if e != nil {
		return baseConfig, e
	}
	if s.Verbose {
		baseConfig.ClientLogMode = aws.LogRequestWithBody & aws.LogResponseWithBody
	}

	creds, e := baseConfig.Credentials.Retrieve(awsContext())

	if e != nil {
		return baseConfig, e
	}

	// Log the principal the API calls run as (once per region, on cache miss),
	// so it's clear from the logs which role/session's privileges are in effect.
	logCallerIdentity(region, baseConfig, creds)

	// terraform cannot ask for MFA token, so we need to pass STS session token, which might contain credentials with MFA requirement
	accessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if accessKey == "" {
		os.Setenv("AWS_ACCESS_KEY_ID", creds.AccessKeyID)
		os.Setenv("AWS_SECRET_ACCESS_KEY", creds.SecretAccessKey)

		if creds.SessionToken != "" {
			os.Setenv("AWS_SESSION_TOKEN", creds.SessionToken)
		}
	}
	configCache[region] = &baseConfig
	return baseConfig, nil
}

func (s *AWSService) buildBaseConfig() (aws.Config, error) {
	var loadOptions []func(*config.LoadOptions) error
	if s.GetArgs()["profile"].(string) != "" {
		loadOptions = append(loadOptions, config.WithSharedConfigProfile(s.GetArgs()["profile"].(string)))
	}
	if s.GetArgs()["region"].(string) != "" {
		os.Setenv("AWS_REGION", s.GetArgs()["region"].(string))
	}
	loadOptions = append(loadOptions, config.WithAssumeRoleCredentialOptions(func(options *stscreds.AssumeRoleOptions) {
		options.TokenProvider = stscreds.StdinTokenProvider
	}))
	// Adaptive retry backs off under throttling (expected when importing many
	// services across regions); cap attempts so a throttled large import
	// degrades instead of retrying unboundedly. ponytail: 5 attempts, bump if
	// large accounts still see ThrottlingException surface as errors.
	loadOptions = append(loadOptions, config.WithRetryer(func() aws.Retryer {
		return retry.AddWithMaxAttempts(retry.NewAdaptiveMode(), 5)
	}))
	return config.LoadDefaultConfig(awsContext(), loadOptions...)
}

// for CF interpolation and IAM Policy variables
func (*AWSService) escapeAwsInterpolation(str string) string {
	return awsVariable.ReplaceAllString(str, "$$$1")
}

func (s *AWSService) getAccountNumber(config aws.Config) (*string, error) {
	stsSvc := sts.NewFromConfig(config)
	identity, err := stsSvc.GetCallerIdentity(awsContext(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, err
	}
	return identity.Account, nil
}

// maskAccessKeyID keeps the first 4 chars (ASIA = temporary session, AKIA =
// long-lived IAM user) and the last 4, masking the middle. The access key ID is
// identifying but not itself secret; the secret key and session token are never
// logged.
func maskAccessKeyID(id string) string {
	if len(id) <= 8 {
		return "****"
	}
	return id[:4] + strings.Repeat("*", len(id)-8) + id[len(id)-4:]
}

// logCallerIdentity records which principal the API calls run as (assumed-role
// ARN, account, user id) plus whether the credentials are temporary and when
// they expire. Called once per region on config cache-miss. Best-effort: an STS
// failure is logged but never aborts the import.
func logCallerIdentity(region string, cfg aws.Config, creds aws.Credentials) {
	credType := "long-lived"
	if creds.SessionToken != "" || strings.HasPrefix(creds.AccessKeyID, "ASIA") {
		credType = "temporary"
	}
	expiry := "none"
	if creds.CanExpire {
		expiry = creds.Expires.UTC().Format(time.RFC3339)
	}
	source := creds.Source
	if source == "" {
		source = "unknown"
	}

	identity, err := sts.NewFromConfig(cfg).GetCallerIdentity(awsContext(), &sts.GetCallerIdentityInput{})
	if err != nil {
		log.Printf("aws [%s] caller identity lookup failed (%s creds, source=%s): %v",
			region, credType, source, err)
		return
	}
	log.Printf("aws [%s] API calls run as: arn=%s account=%s userId=%s | creds=%s accessKeyId=%s source=%s expires=%s",
		region, StringValue(identity.Arn), StringValue(identity.Account), StringValue(identity.UserId),
		credType, maskAccessKeyID(creds.AccessKeyID), source, expiry)
}
