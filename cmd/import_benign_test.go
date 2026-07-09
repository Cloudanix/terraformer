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

package cmd

import (
	"errors"
	"testing"
)

func TestIsBenignServiceInitErr(t *testing.T) {
	// Real InitResources errors that mean "service not available here" — skip.
	benign := []string{
		"operation error Shield: ListProtections, StatusCode: 400, ResourceNotFoundException: The subscription does not exist.",
		"operation error odb: ListOdbNetworks, AccessDeniedException: Can't process the request because the account isn't onboarded to the service.",
		"operation error drs: DescribeReplicationConfigurationTemplates, UninitializedAccountException: Account not initialized",
		"operation error FMS: ListPolicies, AccessDeniedException: No default admin could be found for account 774118602354 in Region ap-south-2",
		"operation error FSx: DescribeFileCaches, BadRequest: The requested feature is not enabled for this AWS account.",
		"operation error License Manager: ListLicenseConfigurations, AccessDeniedException: Service role not found.",
		"operation error ObservabilityAdmin: ListTelemetryRules, ValidationException: Telemetry evaluation is not enabled for the requester account",
		"operation error Service Catalog: ListTagOptions, TagOptionNotMigratedException: TagOption Migration not complete",
		"operation error DAX: DescribeClusters, InvalidParameterValueException: Access Denied to API Version: DAX_V3",
		"operation error Bedrock: ListCustomModels, ValidationException: Unknown operation",
		"operation error EC2: DescribeCarrierGateways, api error UnsupportedOperation: The functionality you requested is not available in this region.",
		"operation error Glue: ListRegistries, InvalidInputException: Glue Schema Registry is not available in this region.",
		"operation error SageMaker: ListCodeRepositories, UnknownOperationException: The requested operation is not supported in the called region.",
		"operation error SES: ListReceiptRuleSets, api error InvalidAction: UnknownError",
		"operation error SNS: ListPlatformApplications, api error InvalidAction: Operation (ListPlatformApplications) is not supported in this region",
		"operation error EC2: DescribeVerifiedAccessInstances, api error InvalidAction: The action DescribeVerifiedAccessInstances is not valid for this web service.",
		"operation error EFS: DescribeReplicationConfigurations, ReplicationNotFound: No replications found.",
		"operation error QuickSight: ListAnalyses, exceeded maximum number of attempts, 5, StatusCode: 503, deserialization failed, failed to decode response body, invalid character '<' looking for beginning of value",
		// service/feature/account not enabled, onboarded, subscribed, allowlisted, retired
		"operation error Audit Manager: ListAssessments, AccessDeniedException: Please complete AWS Audit Manager setup from home page to enable this action in this account.",
		"operation error Chime SDK Media Pipelines: ListMediaInsightsPipelineConfigurations, ForbiddenException: This feature is not available",
		"operation error Database Migration Service: DescribeReplicationConfigs, AccessDeniedException: Feature is not accessible",
		"operation error DSQL: ListClusters, AccessDeniedException: Caller is not allowlisted",
		"operation error IoT Events: ListDetectorModels, AccessDeniedException: AWS IoT Events is no longer available to any customers.",
		"operation error Macie2: ListClassificationJobs, AccessDeniedException: Macie is not enabled.",
		"operation error QuickSight: ListAnalyses, UnsupportedUserEditionException: Account 774118602354 is not subscribed for QuickSight",
		"operation error Timestream Write: ListDatabases, api error AccessDeniedException: Only existing Timestream for LiveAnalytics customers can access this API.",
		"operation error SSM Contacts: ListRotations, ValidationException: Invalid value provided - Account not found for the request",
		"operation error CloudSearch: DescribeDomains, NotAuthorized: New domain creation not supported on this account.",
		"operation error Glue: GetDevEndpoints, InvalidInputException: Operation is not supported in <eu-south-2> region",
		"operation error Lightsail: GetInstances, RegionSetupInProgressException: Lightsail is setting up this region for your account, please try again later.",
		"operation error MemoryDB: DescribeMultiRegionClusters, InvalidParameterValueException: This API operation is currently unavailable.",
		"operation error OpenSearch: ListApplications, DisabledOperationException: This Operation is currently not supported for your account",
		"operation error Comprehend: ListDocumentClassifiers, NotAuthorizedException: Your account is not authorized to make this call.",
		"operation error Transcribe: ListLanguageModels, BadRequestException: Your account isn't authorized to call this operation.",
		"operation error SecurityLake: ListDataLakes, AccessDeniedException: The request failed because your account is not authorized to perform this operation.",
		"operation error MediaPackage: ListChannels, ForbiddenException: error Action index not authorized",
		"operation error Lambda: ListFunctionUrlConfigs, AccessDeniedException: Unable to determine service/operation name to be authorized",
	}
	for _, msg := range benign {
		if !isBenignServiceInitErr(errors.New(msg)) {
			t.Errorf("expected benign (skip), got fatal for: %s", msg)
		}
	}

	// Errors that must still surface: genuine IAM/authz denials (which name the
	// principal + action), ambiguous empty AccessDenied, real throttles/5xx, and
	// infra errors. These must NOT be swallowed.
	fatal := []string{
		"operation error synthetics: ListGroups, StatusCode: 401, UnauthorizedException: Unauthorized",
		"operation error synthetics: ListGroups, AccessDeniedException: User: arn:aws:sts::774118602354:assumed-role/AWSReservedSSO_AdministratorAccess/x is not authorized to perform: synthetics:ListGroups on resource: arn:aws:synthetics:...",
		"operation error Bedrock AgentCore Control: ListAgentRuntimes, AccessDeniedException: User: arn:aws:sts::774118602354:assumed-role/x is not authorized to perform: bedrock-agentcore:ListAgentRuntimes",
		"operation error Backup: ListFrameworks, AccessDeniedException: Insufficient privileges to perform this action.",
		"operation error Timestream InfluxDB: ListDbInstances, AccessDeniedException: ",
		"operation error QuickSight: ListAnalyses, exceeded maximum number of attempts, 5, StatusCode: 503",
		"operation error S3 Control: ListStorageLensConfigurations, StatusCode: 403, UnknownError: UnknownError",
		"dial tcp: lookup sts.amazonaws.com: no such host",
	}
	for _, msg := range fatal {
		if isBenignServiceInitErr(errors.New(msg)) {
			t.Errorf("expected fatal (surface), got benign for: %s", msg)
		}
	}
}

func TestIsUnrecoverableAuthErr(t *testing.T) {
	// Expired SSO/OIDC session — must abort the whole run.
	dead := []string{
		"operation error CloudWatch Logs: DescribeLogGroups, OIDC: CreateToken, https response error StatusCode: 400, RequestID: x, InvalidGrantException: ",
		"operation error EC2: DescribeInstances, OIDC: CreateToken, StatusCode: 400, InvalidGrantException:",
		"the SSO session associated with this profile has expired or is otherwise invalid",
		"error loading SSO token: the SSO session's token has expired",
	}
	for _, msg := range dead {
		if !isUnrecoverableAuthErr(errors.New(msg)) {
			t.Errorf("expected unrecoverable-auth for: %s", msg)
		}
	}

	// Per-service failures that must NOT abort the whole run.
	survivable := []string{
		"operation error Macie2: ListClassificationJobs, AccessDeniedException: Macie is not enabled.",
		"operation error EC2: DescribeCarrierGateways, UnsupportedOperation: not available in this region.",
		"operation error S3: ListBuckets, ExpiredToken: token expired", // role-cred expiry: refreshable, not fatal
		"dial tcp: i/o timeout",
		"",
	}
	for _, msg := range survivable {
		if isUnrecoverableAuthErr(errors.New(msg)) {
			t.Errorf("did not expect unrecoverable-auth for: %q", msg)
		}
	}
}
