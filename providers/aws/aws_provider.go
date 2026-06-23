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
	"os"
	"strconv"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/pkg/errors"
	"github.com/zclconf/go-cty/cty"
)

type AWSProvider struct { //nolint
	terraformutils.Provider
	region  string
	profile string
}

const GlobalRegion = "aws-global"
const MainRegionPublicPartition = "us-east-1"
const NoRegion = ""

// SupportedGlobalResources should be bound to a default region. AWS doesn't specify in which region default services are
// placed (see  https://docs.aws.amazon.com/general/latest/gr/rande.html), so we shouldn't assume any region as well
var SupportedGlobalResources = []string{
	"account",
	"savingsplans",
	"billing",
	"invoicing",
	"notifications",
	"notificationscontacts",
	"budgets",
	"cloudfront",
	"ecrpublic",
	"globalaccelerator",
	"iam",
	"networkmanager",
	"organization",
	"route53",
	"shield",
	"waf",
}

// SupportedEastOnlyResources should be bound to us-east-1 region only, and does not work in any other region.
var SupportedEastOnlyResources = []string{
	"cur",
	"route53domains",
	"wafv2_cloudfront",
}

func (p AWSProvider) GetResourceConnections() map[string]map[string][]string {
	return map[string]map[string][]string{
		"alb": {
			"sg":     []string{"security_groups", "id"},
			"subnet": []string{"subnets", "id"},
			"alb": []string{
				"load_balancer_arn", "id",
				"listener_arn", "id",
				// TF ALB TG attachment logic doesn't work well with references (doesn't interpolate)
			},
		},
		"auto_scaling": {
			"sg":     []string{"security_groups", "id"},
			"subnet": []string{"vpc_zone_identifier", "id"},
		},
		"ec2_instance": {
			"sg":     []string{"vpc_security_group_ids", "id"},
			"subnet": []string{"subnet_id", "id"},
			"ebs":    []string{"ebs_block_device", "id"},
		},
		"elasticache": {
			"vpc":    []string{"vpc_id", "id"},
			"subnet": []string{"subnet_ids", "id"},
			"sg":     []string{"security_group_ids", "id"},
		},
		"ebs": {
			// TF EBS attachment logic doesn't work well with references (doesn't interpolate)
		},
		"ecs": {
			// ECS is not able anymore to support references (doesn't interpolate)
			"subnet": []string{"network_configuration.subnets", "id"},
			"sg":     []string{"network_configuration.security_groups", "id"},
		},
		"eks": {
			"subnet": []string{"vpc_config.subnet_ids", "id"},
			"sg":     []string{"vpc_config.security_group_ids", "id"},
		},
		"elb": {
			"sg":     []string{"security_groups", "id"},
			"subnet": []string{"subnets", "id"},
		},
		"igw": {"vpc": []string{"vpc_id", "id"}},
		"identitystore": {
			"identitystore": []string{
				"group_id", "id",
				"member_id", "id",
			},
		},
		"msk": {
			"subnet": []string{"broker_node_group_info.client_subnets", "id"},
			"sg":     []string{"broker_node_group_info.security_groups", "id"},
		},
		"nacl": {
			"subnet": []string{"subnet_ids", "id"},
			"vpc":    []string{"vpc_id", "id"},
		},
		"organization": {
			"organization": []string{
				"policy_id", "id",
				"parent_id", "id",
				"target_id", "id",
			},
		},
		"rds": {
			"subnet": []string{"subnet_ids", "id"},
			"sg":     []string{"vpc_security_group_ids", "id"},
		},
		"route_table": {
			"route_table": []string{"route_table_id", "id"},
			"subnet":      []string{"subnet_id", "id"},
			"vpc":         []string{"vpc_id", "id"},
		},
		"sns": {
			"sns": []string{"topic_arn", "id"},
			"sqs": []string{"endpoint", "arn"},
		},
		"sg": {
			"sg": []string{
				"egress.security_groups", "id",
				"ingress.security_groups", "id",
				"security_group_id", "id",
				"source_security_group_id", "id",
			},
		},
		"subnet": {"vpc": []string{"vpc_id", "id"}},
		"transit_gateway": {
			"vpc":             []string{"vpc_id", "id"},
			"transit_gateway": []string{"transit_gateway_id", "id"},
			"subnet":          []string{"subnet_ids", "id"},
			"vpn_connection":  []string{"vpn_connection_id", "id"},
		},
		"vpn_gateway": {"vpc": []string{"vpc_id", "id"}},
		"vpn_connection": {
			"customer_gateway": []string{"customer_gateway_id", "id"},
			"vpn_gateway":      []string{"vpn_gateway_id", "id"},
		},
	}
}

func (p AWSProvider) GetProviderData(arg ...string) map[string]interface{} {
	awsConfig := map[string]interface{}{}

	if p.region == GlobalRegion {
		awsConfig["region"] = MainRegionPublicPartition // For TF to workaround terraform-providers/terraform-provider-aws#1043
	} else if p.region != NoRegion {
		awsConfig["region"] = p.region
	}

	return map[string]interface{}{
		"provider": map[string]interface{}{
			"aws": awsConfig,
		},
	}
}

func (p *AWSProvider) GetConfig() cty.Value {
	if p.region != GlobalRegion {
		return cty.ObjectVal(map[string]cty.Value{
			"region":                 cty.StringVal(p.region),
			"skip_region_validation": cty.True,
		})
	}
	return cty.ObjectVal(map[string]cty.Value{
		"region":                 cty.StringVal(""),
		"skip_region_validation": cty.True,
	})
}

func (p *AWSProvider) GetBasicConfig() cty.Value {
	return p.GetConfig()
}

// check projectName in env params
func (p *AWSProvider) Init(args []string) error {
	p.region = args[0]
	p.profile = args[1]

	// Terraformer accepts region and profile configuration, so we must detect what env variables to adjust to make Go SDK rely on them. AWS_SDK_LOAD_CONFIG here must be checked to determine correct variable to set.
	enableSharedConfig, _ := strconv.ParseBool(os.Getenv("AWS_SDK_LOAD_CONFIG"))
	var err error
	if p.region != GlobalRegion && p.region != NoRegion {
		if enableSharedConfig {
			err = os.Setenv("AWS_DEFAULT_REGION", p.region)
		} else {
			err = os.Setenv("AWS_REGION", p.region)
		}
		if err != nil {
			return err
		}
	}

	if p.profile != "" && p.profile != "default" {
		envVar := "AWS_PROFILE"
		if enableSharedConfig {
			envVar = "AWS_DEFAULT_PROFILE"
		}

		if err := os.Setenv(envVar, p.profile); err != nil {
			return err
		}
	}
	return nil
}

func (p *AWSProvider) GetName() string {
	return "aws"
}

func (p *AWSProvider) InitService(serviceName string, verbose bool) error {
	var isSupported bool
	if _, isSupported = p.GetSupportedService()[serviceName]; !isSupported {
		return errors.New("aws: " + serviceName + " not supported service")
	}
	p.Service = p.GetSupportedService()[serviceName]
	p.Service.SetName(serviceName)
	p.Service.SetVerbose(verbose)
	p.Service.SetProviderName(p.GetName())
	p.Service.SetArgs(map[string]interface{}{
		"region":                 p.region,
		"profile":                p.profile,
		"skip_region_validation": true,
	})
	return nil
}

// GetAWSSupportService return map of support service for AWS
func (p *AWSProvider) GetSupportedService() map[string]terraformutils.ServiceGenerator {
	return map[string]terraformutils.ServiceGenerator{
		"accessanalyzer":               &AwsFacade{service: &AccessAnalyzerGenerator{}},
		"acm":                          &AwsFacade{service: &ACMGenerator{}},
		"acm-pca":                      &AwsFacade{service: &ACMPCAGenerator{}},
		"alb":                          &AwsFacade{service: &AlbGenerator{}},
		"amplify":                      &AwsFacade{service: &AmplifyGenerator{}},
		"api_gateway":                  &AwsFacade{service: &APIGatewayGenerator{}},
		"api_gatewayv2":                &AwsFacade{service: &APIGatewayV2Generator{}},
		"application-autoscaling":      &AwsFacade{service: &AppAutoScalingGenerator{}},
		"appconfig":                    &AwsFacade{service: &AppConfigGenerator{}},
		"appflow":                      &AwsFacade{service: &AppFlowGenerator{}},
		"appstream":                    &AwsFacade{service: &AppStreamGenerator{}},
		"appmesh":                      &AwsFacade{service: &AppMeshGenerator{}},
		"apprunner":                    &AwsFacade{service: &AppRunnerGenerator{}},
		"appintegrations":              &AwsFacade{service: &AppIntegrationsGenerator{}},
		"appsync":                      &AwsFacade{service: &AppSyncGenerator{}},
		"athena":                       &AwsFacade{service: &AthenaGenerator{}},
		"auto_scaling":                 &AwsFacade{service: &AutoScalingGenerator{}},
		"backup":                       &AwsFacade{service: &BackupGenerator{}},
		"batch":                        &AwsFacade{service: &BatchGenerator{}},
		"budgets":                      &AwsFacade{service: &BudgetsGenerator{}},
		"cloud9":                       &AwsFacade{service: &Cloud9Generator{}},
		"cloudformation":               &AwsFacade{service: &CloudFormationGenerator{}},
		"cloudfront":                   &AwsFacade{service: &CloudFrontGenerator{}},
		"cloudhsm":                     &AwsFacade{service: &CloudHsmGenerator{}},
		"cloudtrail":                   &AwsFacade{service: &CloudTrailGenerator{}},
		"cloudwatch":                   &AwsFacade{service: &CloudWatchGenerator{}},
		"codebuild":                    &AwsFacade{service: &CodeBuildGenerator{}},
		"codecommit":                   &AwsFacade{service: &CodeCommitGenerator{}},
		"codeartifact":                 &AwsFacade{service: &CodeArtifactGenerator{}},
		"codeguru-profiler":            &AwsFacade{service: &CodeGuruProfilerGenerator{}},
		"codeguru-reviewer":            &AwsFacade{service: &CodeGuruReviewerGenerator{}},
		"codestar-notifications":       &AwsFacade{service: &CodeStarNotificationsGenerator{}},
		"controltower":                 &AwsFacade{service: &ControlTowerGenerator{}},
		"chime-sdk-mediapipelines":     &AwsFacade{service: &ChimeSDKMediaPipelinesGenerator{}},
		"chime-sdk-voice":              &AwsFacade{service: &ChimeSDKVoiceGenerator{}},
		"autoscaling-plans":            &AwsFacade{service: &AutoScalingPlansGenerator{}},
		"codedeploy":                   &AwsFacade{service: &CodeDeployGenerator{}},
		"codepipeline":                 &AwsFacade{service: &CodePipelineGenerator{}},
		"codestar-connections":         &AwsFacade{service: &CodeStarConnectionsGenerator{}},
		"connect":                      &AwsFacade{service: &ConnectGenerator{}},
		"customer-profiles":            &AwsFacade{service: &CustomerProfilesGenerator{}},
		"dataexchange":                 &AwsFacade{service: &DataExchangeGenerator{}},
		"datazone":                     &AwsFacade{service: &DataZoneGenerator{}},
		"cognito":                      &AwsFacade{service: &CognitoGenerator{}},
		"config":                       &AwsFacade{service: &ConfigGenerator{}},
		"customer_gateway":             &AwsFacade{service: &CustomerGatewayGenerator{}},
		"datapipeline":                 &AwsFacade{service: &DataPipelineGenerator{}},
		"datasync":                     &AwsFacade{service: &DataSyncGenerator{}},
		"dax":                          &AwsFacade{service: &DaxGenerator{}},
		"detective":                    &AwsFacade{service: &DetectiveGenerator{}},
		"devicefarm":                   &AwsFacade{service: &DeviceFarmGenerator{}},
		"dlm":                          &AwsFacade{service: &DlmGenerator{}},
		"dms":                          &AwsFacade{service: &DmsGenerator{}},
		"docdb":                        &AwsFacade{service: &DocDBGenerator{}},
		"docdb-elastic":                &AwsFacade{service: &DocDBElasticGenerator{}},
		"ds":                           &AwsFacade{service: &DirectoryServiceGenerator{}},
		"dx":                           &AwsFacade{service: &DirectConnectGenerator{}},
		"dynamodb":                     &AwsFacade{service: &DynamoDbGenerator{}},
		"ebs":                          &AwsFacade{service: &EbsGenerator{}},
		"ec2_instance":                 &AwsFacade{service: &Ec2Generator{}},
		"ecr":                          &AwsFacade{service: &EcrGenerator{}},
		"ecrpublic":                    &AwsFacade{service: &EcrPublicGenerator{}},
		"ecs":                          &AwsFacade{service: &EcsGenerator{}},
		"efs":                          &AwsFacade{service: &EfsGenerator{}},
		"eks":                          &AwsFacade{service: &EksGenerator{}},
		"eip":                          &AwsFacade{service: &ElasticIPGenerator{}},
		"elasticache":                  &AwsFacade{service: &ElastiCacheGenerator{}},
		"elastic_beanstalk":            &AwsFacade{service: &BeanstalkGenerator{}},
		"elb":                          &AwsFacade{service: &ElbGenerator{}},
		"emr":                          &AwsFacade{service: &EmrGenerator{}},
		"emr-containers":               &AwsFacade{service: &EMRContainersGenerator{}},
		"emr-serverless":               &AwsFacade{service: &EMRServerlessGenerator{}},
		"eni":                          &AwsFacade{service: &EniGenerator{}},
		"es":                           &AwsFacade{service: &EsGenerator{}},
		"finspace":                     &AwsFacade{service: &FinspaceGenerator{}},
		"firehose":                     &AwsFacade{service: &FirehoseGenerator{}},
		"fis":                          &AwsFacade{service: &FISGenerator{}},
		"fms":                          &AwsFacade{service: &FmsGenerator{}},
		"fsx":                          &AwsFacade{service: &FsxGenerator{}},
		"glacier":                      &AwsFacade{service: &GlacierGenerator{}},
		"amp":                          &AwsFacade{service: &PrometheusGenerator{}},
		"appfabric":                    &AwsFacade{service: &AppFabricGenerator{}},
		"applicationinsights":          &AwsFacade{service: &ApplicationInsightsGenerator{}},
		"auditmanager":                 &AwsFacade{service: &AuditManagerGenerator{}},
		"bedrock":                      &AwsFacade{service: &BedrockGenerator{}},
		"account":                      &AwsFacade{service: &AccountGenerator{}},
		"s3vectors":                    &AwsFacade{service: &S3VectorsGenerator{}},
		"arc-region-switch":            &AwsFacade{service: &ARCRegionSwitchGenerator{}},
		"arc-zonal-shift":              &AwsFacade{service: &ARCZonalShiftGenerator{}},
		"dsql":                         &AwsFacade{service: &DSQLGenerator{}},
		"workmail":                     &AwsFacade{service: &WorkMailGenerator{}},
		"odb":                          &AwsFacade{service: &ODBGenerator{}},
		"networkflowmonitor":           &AwsFacade{service: &NetworkFlowMonitorGenerator{}},
		"notifications":                &AwsFacade{service: &NotificationsGenerator{}},
		"notificationscontacts":        &AwsFacade{service: &NotificationsContactsGenerator{}},
		"savingsplans":                 &AwsFacade{service: &SavingsPlansGenerator{}},
		"timestream-query":             &AwsFacade{service: &TimestreamQueryGenerator{}},
		"billing":                      &AwsFacade{service: &BillingGenerator{}},
		"invoicing":                    &AwsFacade{service: &InvoicingGenerator{}},
		"observabilityadmin":           &AwsFacade{service: &ObservabilityAdminGenerator{}},
		"bcmdataexports":               &AwsFacade{service: &BCMDataExportsGenerator{}},
		"bedrockagent":                 &AwsFacade{service: &BedrockAgentGenerator{}},
		"ce":                           &AwsFacade{service: &CostExplorerGenerator{}},
		"chatbot":                      &AwsFacade{service: &ChatbotGenerator{}},
		"cleanrooms":                   &AwsFacade{service: &CleanRoomsGenerator{}},
		"comprehend":                   &AwsFacade{service: &ComprehendGenerator{}},
		"compute-optimizer":            &AwsFacade{service: &ComputeOptimizerGenerator{}},
		"cost-optimization-hub":        &AwsFacade{service: &CostOptimizationHubGenerator{}},
		"cur":                          &AwsFacade{service: &CURGenerator{}},
		"devops-guru":                  &AwsFacade{service: &DevOpsGuruGenerator{}},
		"cloudsearch":                  &AwsFacade{service: &CloudSearchGenerator{}},
		"elastictranscoder":            &AwsFacade{service: &ElasticTranscoderGenerator{}},
		"evidently":                    &AwsFacade{service: &EvidentlyGenerator{}},
		"globalaccelerator":            &AwsFacade{service: &GlobalAcceleratorGenerator{}},
		"grafana":                      &AwsFacade{service: &GrafanaGenerator{}},
		"gamelift":                     &AwsFacade{service: &GameLiftGenerator{}},
		"greengrassv2":                 &AwsFacade{service: &GreengrassV2Generator{}},
		"groundstation":                &AwsFacade{service: &GroundStationGenerator{}},
		"mediaconvert":                 &AwsFacade{service: &MediaConvertGenerator{}},
		"quicksight":                   &AwsFacade{service: &QuickSightGenerator{}},
		"s3control":                    &AwsFacade{service: &S3ControlGenerator{}},
		"s3outposts":                   &AwsFacade{service: &S3OutpostsGenerator{}},
		"glue":                         &AwsFacade{service: &GlueGenerator{}},
		"guardduty":                    &AwsFacade{service: &GuardDutyGenerator{}},
		"healthlake":                   &AwsFacade{service: &HealthLakeGenerator{}},
		"iam":                          &AwsFacade{service: &IamGenerator{}},
		"imagebuilder":                 &AwsFacade{service: &ImageBuilderGenerator{}},
		"ivschat":                      &AwsFacade{service: &IVSChatGenerator{}},
		"ivs":                          &AwsFacade{service: &IVSGenerator{}},
		"identitystore":                &AwsFacade{service: &IdentityStoreGenerator{}},
		"igw":                          &AwsFacade{service: &IgwGenerator{}},
		"inspector":                    &AwsFacade{service: &InspectorGenerator{}},
		"inspector2":                   &AwsFacade{service: &Inspector2Generator{}},
		"iot":                          &AwsFacade{service: &IotGenerator{}},
		"iotevents":                    &AwsFacade{service: &IoTEventsGenerator{}},
		"iotsitewise":                  &AwsFacade{service: &IoTSiteWiseGenerator{}},
		"iottwinmaker":                 &AwsFacade{service: &IoTTwinMakerGenerator{}},
		"kafkaconnect":                 &AwsFacade{service: &KafkaConnectGenerator{}},
		"kendra":                       &AwsFacade{service: &KendraGenerator{}},
		"lexv2-models":                 &AwsFacade{service: &LexModelsV2Generator{}},
		"license-manager":              &AwsFacade{service: &LicenseManagerGenerator{}},
		"location":                     &AwsFacade{service: &LocationGenerator{}},
		"lightsail":                    &AwsFacade{service: &LightsailGenerator{}},
		"m2":                           &AwsFacade{service: &M2Generator{}},
		"keyspaces":                    &AwsFacade{service: &KeyspacesGenerator{}},
		"kinesis":                      &AwsFacade{service: &KinesisGenerator{}},
		"kinesisanalyticsv2":           &AwsFacade{service: &KinesisAnalyticsV2Generator{}},
		"kinesisvideo":                 &AwsFacade{service: &KinesisVideoGenerator{}},
		"lakeformation":                &AwsFacade{service: &LakeFormationGenerator{}},
		"kms":                          &AwsFacade{service: &KmsGenerator{}},
		"lambda":                       &AwsFacade{service: &LambdaGenerator{}},
		"logs":                         &AwsFacade{service: &LogsGenerator{}},
		"macie2":                       &AwsFacade{service: &Macie2Generator{}},
		"media_package":                &AwsFacade{service: &MediaPackageGenerator{}},
		"media_store":                  &AwsFacade{service: &MediaStoreGenerator{}},
		"medialive":                    &AwsFacade{service: &MediaLiveGenerator{}},
		"mediapackagev2":               &AwsFacade{service: &MediaPackageV2Generator{}},
		"mq":                           &AwsFacade{service: &MQGenerator{}},
		"memorydb":                     &AwsFacade{service: &MemoryDBGenerator{}},
		"mwaa":                         &AwsFacade{service: &MWAAGenerator{}},
		"msk":                          &AwsFacade{service: &MskGenerator{}},
		"neptune":                      &AwsFacade{service: &NeptuneGenerator{}},
		"neptune-graph":                &AwsFacade{service: &NeptuneGraphGenerator{}},
		"opensearch":                   &AwsFacade{service: &OpenSearchGenerator{}},
		"opensearchserverless":         &AwsFacade{service: &OpenSearchServerlessGenerator{}},
		"osis":                         &AwsFacade{service: &OSISGenerator{}},
		"pinpointsmsvoicev2":           &AwsFacade{service: &PinpointSMSVoiceV2Generator{}},
		"pinpoint":                     &AwsFacade{service: &PinpointGenerator{}},
		"pipes":                        &AwsFacade{service: &PipesGenerator{}},
		"payment-cryptography":         &AwsFacade{service: &PaymentCryptographyGenerator{}},
		"pca-connector-ad":             &AwsFacade{service: &PCAConnectorADGenerator{}},
		"pcs":                          &AwsFacade{service: &PCSGenerator{}},
		"proton":                       &AwsFacade{service: &ProtonGenerator{}},
		"qbusiness":                    &AwsFacade{service: &QBusinessGenerator{}},
		"rbin":                         &AwsFacade{service: &RbinGenerator{}},
		"nacl":                         &AwsFacade{service: &NaclGenerator{}},
		"networkmanager":               &AwsFacade{service: &NetworkManagerGenerator{}},
		"network-firewall":             &AwsFacade{service: &NetworkFirewallGenerator{}},
		"rolesanywhere":                &AwsFacade{service: &RolesAnywhereGenerator{}},
		"securitylake":                 &AwsFacade{service: &SecurityLakeGenerator{}},
		"nat":                          &AwsFacade{service: &NatGatewayGenerator{}},
		"networkmonitor":               &AwsFacade{service: &NetworkMonitorGenerator{}},
		"internetmonitor":              &AwsFacade{service: &InternetMonitorGenerator{}},
		"oam":                          &AwsFacade{service: &OAMGenerator{}},
		"rum":                          &AwsFacade{service: &RumGenerator{}},
		"synthetics":                   &AwsFacade{service: &SyntheticsGenerator{}},
		"opsworks":                     &AwsFacade{service: &OpsworksGenerator{}},
		"organization":                 &AwsFacade{service: &OrganizationGenerator{}},
		"qldb":                         &AwsFacade{service: &QLDBGenerator{}},
		"ram":                          &AwsFacade{service: &RamGenerator{}},
		"rds":                          &AwsFacade{service: &RDSGenerator{}},
		"redshift":                     &AwsFacade{service: &RedshiftGenerator{}},
		"redshift-serverless":          &AwsFacade{service: &RedshiftServerlessGenerator{}},
		"sagemaker":                    &AwsFacade{service: &SageMakerGenerator{}},
		"scheduler":                    &AwsFacade{service: &SchedulerGenerator{}},
		"schemas":                      &AwsFacade{service: &SchemasGenerator{}},
		"rekognition":                  &AwsFacade{service: &RekognitionGenerator{}},
		"resiliencehub":                &AwsFacade{service: &ResilienceHubGenerator{}},
		"resource-explorer-2":          &AwsFacade{service: &ResourceExplorer2Generator{}},
		"resourcegroups":               &AwsFacade{service: &ResourceGroupsGenerator{}},
		"route53":                      &AwsFacade{service: &Route53Generator{}},
		"route53recoveryreadiness":     &AwsFacade{service: &Route53RecoveryReadinessGenerator{}},
		"route53recoverycontrolconfig": &AwsFacade{service: &Route53RecoveryControlConfigGenerator{}},
		"route53profiles":              &AwsFacade{service: &Route53ProfilesGenerator{}},
		"route53domains":               &AwsFacade{service: &Route53DomainsGenerator{}},
		"route53resolver":              &AwsFacade{service: &Route53ResolverGenerator{}},
		"route_table":                  &AwsFacade{service: &RouteTableGenerator{}},
		"s3":                           &AwsFacade{service: &S3Generator{}},
		"s3tables":                     &AwsFacade{service: &S3TablesGenerator{}},
		"secretsmanager":               &AwsFacade{service: &SecretsManagerGenerator{}},
		"securityhub":                  &AwsFacade{service: &SecurityhubGenerator{}},
		"servicecatalog":               &AwsFacade{service: &ServiceCatalogGenerator{}},
		"servicecatalog-appregistry":   &AwsFacade{service: &ServiceCatalogAppRegistryGenerator{}},
		"servicediscovery":             &AwsFacade{service: &ServiceDiscoveryGenerator{}},
		"servicequotas":                &AwsFacade{service: &ServiceQuotasGenerator{}},
		"shield":                       &AwsFacade{service: &ShieldGenerator{}},
		"signer":                       &AwsFacade{service: &SignerGenerator{}},
		"ses":                          &AwsFacade{service: &SesGenerator{}},
		"sesv2":                        &AwsFacade{service: &SESv2Generator{}},
		"sfn":                          &AwsFacade{service: &SfnGenerator{}},
		"sg":                           &AwsFacade{service: &SecurityGenerator{}},
		"sqs":                          &AwsFacade{service: &SqsGenerator{}},
		"sns":                          &AwsFacade{service: &SnsGenerator{}},
		"sso-admin":                    &AwsFacade{service: &SSOAdminGenerator{}},
		"ssm":                          &AwsFacade{service: &SsmGenerator{}},
		"ssmquicksetup":                &AwsFacade{service: &SSMQuickSetupGenerator{}},
		"ssm-contacts":                 &AwsFacade{service: &SSMContactsGenerator{}},
		"ssm-incidents":                &AwsFacade{service: &SSMIncidentsGenerator{}},
		"storagegateway":               &AwsFacade{service: &StorageGatewayGenerator{}},
		"subnet":                       &AwsFacade{service: &SubnetGenerator{}},
		"swf":                          &AwsFacade{service: &SWFGenerator{}},
		"timestream-influxdb":          &AwsFacade{service: &TimestreamInfluxDBGenerator{}},
		"timestream-write":             &AwsFacade{service: &TimestreamWriteGenerator{}},
		"transcribe":                   &AwsFacade{service: &TranscribeGenerator{}},
		"transfer":                     &AwsFacade{service: &TransferGenerator{}},
		"transit_gateway":              &AwsFacade{service: &TransitGatewayGenerator{}},
		"waf":                          &AwsFacade{service: &WafGenerator{}},
		"wellarchitected":              &AwsFacade{service: &WellArchitectedGenerator{}},
		"waf_regional":                 &AwsFacade{service: &WafRegionalGenerator{}},
		"wafv2_cloudfront":             &AwsFacade{service: NewWafv2CloudfrontGenerator()},
		"wafv2_regional":               &AwsFacade{service: NewWafv2RegionalGenerator()},
		"verifiedaccess":               &AwsFacade{service: &VerifiedAccessGenerator{}},
		"verifiedpermissions":          &AwsFacade{service: &VerifiedPermissionsGenerator{}},
		"vpc":                          &AwsFacade{service: &VpcGenerator{}},
		"vpc-lattice":                  &AwsFacade{service: &VPCLatticeGenerator{}},
		"vpc_endpoint":                 &AwsFacade{service: &VpcEndpointGenerator{}},
		"vpc_peering":                  &AwsFacade{service: &VpcPeeringConnectionGenerator{}},
		"vpn_connection":               &AwsFacade{service: &VpnConnectionGenerator{}},
		"vpn_gateway":                  &AwsFacade{service: &VpnGatewayGenerator{}},
		"workspaces":                   &AwsFacade{service: &WorkspacesGenerator{}},
		"workspaces-web":               &AwsFacade{service: &WorkSpacesWebGenerator{}},
		"xray":                         &AwsFacade{service: &XrayGenerator{}},
	}
}

func StringValue(value *string) string {
	if value != nil {
		return *value
	}
	return ""
}
