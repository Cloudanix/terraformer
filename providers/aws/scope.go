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

// regionScope is the region binding of an AWS service generator.
//
//	scopeRegional  imported once per requested region
//	scopeGlobal    bound to a default region (IAM, Route53, CloudFront…);
//	               imported in the "aws-global" pass
//	scopeEastOnly  only works in us-east-1 (e.g. WAFv2 for CloudFront)
type regionScope int

const (
	scopeRegional regionScope = iota
	scopeGlobal
	scopeEastOnly
)

// serviceScope is the single declarative source of truth for every registered
// AWS service's region binding. scope_test.go asserts it against the registry
// (GetSupportedService) and against SupportedGlobalResources /
// SupportedEastOnlyResources — so adding a service to the registry without a
// scope, or misclassifying one, fails the build instead of silently importing
// into the wrong region path / failing to sign (the bug class fixed in
// commit c127700b). Every key needs an explicit entry, including regional, so
// the assertion can prove completeness.
var serviceScope = map[string]regionScope{
	// global — bound to a default region, imported in the aws-global pass
	"budgets":           scopeGlobal,
	"cloudfront":        scopeGlobal,
	"ecrpublic":         scopeGlobal,
	"globalaccelerator": scopeGlobal,
	"iam":               scopeGlobal,
	"organization":      scopeGlobal,
	"route53":           scopeGlobal,
	"shield":            scopeGlobal,
	"waf":               scopeGlobal,

	// us-east-1 only
	"wafv2_cloudfront": scopeEastOnly,

	// regional (default)
	"accessanalyzer":          scopeRegional,
	"acm":                     scopeRegional,
	"alb":                     scopeRegional,
	"api_gateway":             scopeRegional,
	"api_gatewayv2":           scopeRegional,
	"application-autoscaling": scopeRegional,
	"appsync":                 scopeRegional,
	"auto_scaling":            scopeRegional,
	"backup":                  scopeRegional,
	"batch":                   scopeRegional,
	"cloud9":                  scopeRegional,
	"cloudformation":          scopeRegional,
	"cloudhsm":                scopeRegional,
	"cloudtrail":              scopeRegional,
	"cloudwatch":              scopeRegional,
	"codebuild":               scopeRegional,
	"codecommit":              scopeRegional,
	"codedeploy":              scopeRegional,
	"codepipeline":            scopeRegional,
	"cognito":                 scopeRegional,
	"config":                  scopeRegional,
	"customer_gateway":        scopeRegional,
	"datapipeline":            scopeRegional,
	"datasync":                scopeRegional,
	"dax":                     scopeRegional,
	"devicefarm":              scopeRegional,
	"dlm":                     scopeRegional,
	"dms":                     scopeRegional,
	"docdb":                   scopeRegional,
	"ds":                      scopeRegional,
	"dx":                      scopeRegional,
	"dynamodb":                scopeRegional,
	"ebs":                     scopeRegional,
	"ec2_instance":            scopeRegional,
	"ecr":                     scopeRegional,
	"ecs":                     scopeRegional,
	"efs":                     scopeRegional,
	"eip":                     scopeRegional,
	"eks":                     scopeRegional,
	"elastic_beanstalk":       scopeRegional,
	"elasticache":             scopeRegional,
	"elb":                     scopeRegional,
	"emr":                     scopeRegional,
	"eni":                     scopeRegional,
	"es":                      scopeRegional,
	"firehose":                scopeRegional,
	"fsx":                     scopeRegional,
	"glacier":                 scopeRegional,
	"glue":                    scopeRegional,
	"guardduty":               scopeRegional,
	"identitystore":           scopeRegional,
	"igw":                     scopeRegional,
	"iot":                     scopeRegional,
	"kinesis":                 scopeRegional,
	"kms":                     scopeRegional,
	"lambda":                  scopeRegional,
	"logs":                    scopeRegional,
	"macie2":                  scopeRegional,
	"media_package":           scopeRegional,
	"media_store":             scopeRegional,
	"medialive":               scopeRegional,
	"mq":                      scopeRegional,
	"msk":                     scopeRegional,
	"nacl":                    scopeRegional,
	"nat":                     scopeRegional,
	"opsworks":                scopeRegional,
	"qldb":                    scopeRegional,
	"rds":                     scopeRegional,
	"redshift":                scopeRegional,
	"resourcegroups":          scopeRegional,
	"route53resolver":         scopeRegional,
	"route_table":             scopeRegional,
	"s3":                      scopeRegional,
	"secretsmanager":          scopeRegional,
	"securityhub":             scopeRegional,
	"servicecatalog":          scopeRegional,
	"servicediscovery":        scopeRegional,
	"servicequotas":           scopeRegional,
	"ses":                     scopeRegional,
	"sesv2":                   scopeRegional,
	"sfn":                     scopeRegional,
	"sg":                      scopeRegional,
	"sns":                     scopeRegional,
	"sqs":                     scopeRegional,
	"sso-admin":               scopeRegional,
	"ssm":                     scopeRegional,
	"storagegateway":          scopeRegional,
	"subnet":                  scopeRegional,
	"swf":                     scopeRegional,
	"transfer":                scopeRegional,
	"transit_gateway":         scopeRegional,
	"vpc":                     scopeRegional,
	"vpc_endpoint":            scopeRegional,
	"vpc_peering":             scopeRegional,
	"vpn_connection":          scopeRegional,
	"vpn_gateway":             scopeRegional,
	"waf_regional":            scopeRegional,
	"wafv2_regional":          scopeRegional,
	"workspaces":              scopeRegional,
	"xray":                    scopeRegional,
}
