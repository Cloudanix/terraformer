// Copyright 2019 The Terraformer Authors.
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

package azure

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/hashicorp/go-azure-helpers/authentication"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils/providerwrapper"
)

type AzureProvider struct { //nolint
	terraformutils.Provider
	config        authentication.Config
	credential    azcore.TokenCredential
	resourceGroup string
}

func (p *AzureProvider) setEnvConfig() error {
	subscriptionID := os.Getenv("ARM_SUBSCRIPTION_ID")
	if subscriptionID == "" {
		return errors.New("set ARM_SUBSCRIPTION_ID env var")
	}
	var auxTenants []string
	if v := os.Getenv("ARM_AUXILIARY_TENANT_IDS"); v != "" {
		auxTenants = strings.Split(v, ";")
		if len(auxTenants) > 3 {
			return fmt.Errorf("the provider only supports 3 auxiliary tenant IDs for ARM_AUXILIARY_TENANT_IDS")
		}
	}
	builder := &authentication.Builder{
		ClientID:            os.Getenv("ARM_CLIENT_ID"),
		SubscriptionID:      subscriptionID,
		TenantID:            os.Getenv("ARM_TENANT_ID"),
		AuxiliaryTenantIDs:  auxTenants,
		Environment:         os.Getenv("ARM_ENVIRONMENT"),
		MetadataHost:        os.Getenv("ARM_METADATA_HOSTNAME"),
		MsiEndpoint:         os.Getenv("ARM_MSI_ENDPOINT"),
		ClientSecret:        os.Getenv("ARM_CLIENT_SECRET"),
		ClientCertPath:      os.Getenv("ARM_CLIENT_CERTIFICATE_PATH"),
		ClientCertPassword:  os.Getenv("ARM_CLIENT_CERTIFICATE_PASSWORD"),
		IDTokenRequestToken: os.Getenv("ARM_OIDC_REQUEST_TOKEN"),
		IDTokenRequestURL:   os.Getenv("ARM_OIDC_REQUEST_URL"),

		// Feature Toggles
		SupportsAzureCliToken:          true,
		SupportsClientSecretAuth:       true,
		SupportsClientCertAuth:         true,
		SupportsManagedServiceIdentity: os.Getenv("ARM_USE_MSI") != "",
		SupportsOIDCAuth:               os.Getenv("ARM_USE_OIDC") != "",
		UseMicrosoftGraph:              os.Getenv("ARM_USE_ADAL") == "",
	}

	if builder.Environment == "" {
		builder.Environment = "public"
	}
	config, err := builder.Build()
	if err != nil {
		return nil
	}
	p.config = *config

	return nil
}

// getCredential builds the azcore credential used by every (Track 2) generator.
// DefaultAzureCredential reads AZURE_* env vars / MSI / Azure CLI / workload
// identity. We bridge the terraformer ARM_* vars onto AZURE_* first so the
// existing ARM_* credentials keep working. Construction does not authenticate;
// tokens are fetched lazily on first use.
func (p *AzureProvider) getCredential() (azcore.TokenCredential, error) {
	bridge := map[string]string{
		"ARM_CLIENT_ID":                   "AZURE_CLIENT_ID",
		"ARM_TENANT_ID":                   "AZURE_TENANT_ID",
		"ARM_CLIENT_SECRET":               "AZURE_CLIENT_SECRET",
		"ARM_CLIENT_CERTIFICATE_PATH":     "AZURE_CLIENT_CERTIFICATE_PATH",
		"ARM_CLIENT_CERTIFICATE_PASSWORD": "AZURE_CLIENT_CERTIFICATE_PASSWORD",
	}
	for arm, azure := range bridge {
		if os.Getenv(azure) == "" {
			if v := os.Getenv(arm); v != "" {
				os.Setenv(azure, v)
			}
		}
	}
	return azidentity.NewDefaultAzureCredential(nil)
}

func (p *AzureProvider) Init(args []string) error {
	err := p.setEnvConfig()
	if err != nil {
		return err
	}

	credential, err := p.getCredential()
	if err != nil {
		return err
	}
	p.credential = credential

	p.resourceGroup = args[0]

	return nil
}

func (p *AzureProvider) GetName() string {
	return "azurerm"
}

func (p *AzureProvider) GetProviderData(arg ...string) map[string]interface{} {
	version := providerwrapper.GetProviderVersion(p.GetName())
	if strings.Contains(version, "v2.") {
		return map[string]interface{}{
			"provider": map[string]interface{}{
				"azurerm": map[string]interface{}{
					// NOTE:
					// Workaround for azurerm v2 provider changes
					// Tested with azurerm_resource_group under v2.17.0
					// https://github.com/terraform-providers/terraform-provider-azurerm/issues/5866#issuecomment-594239342
					// https://github.com/hashicorp/terraform/issues/24200#issuecomment-594745861
					"features": map[string]interface{}{},
				},
			},
		}
	}
	return map[string]interface{}{
		"provider": map[string]interface{}{
			"azurerm": map[string]interface{}{
				"version": version,
			},
		},
	}
}

func (AzureProvider) GetResourceConnections() map[string]map[string][]string {
	return map[string]map[string][]string{
		"analysis": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"app_service": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"application_gateway": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"cosmosdb": {
			"resource_group": []string{
				"resource_group_name", "name",
				"location", "location",
			},
		},
		"container": {
			"resource_group": []string{
				"resource_group_name", "name",
				"location", "location",
			},
		},
		"database": {
			"resource_group": []string{
				"resource_group_name", "name",
				"location", "location",
			},
		},
		"databricks": {
			"resource_group": []string{
				"resource_group_name", "name",
				"managed_resource_group_name", "name",
				"location", "location",
			},
			"storage_account": []string{"storage_account_name", "name"},
			"subnet": []string{
				"public_subnet_name", "name",
				"private_subnet_name", "name",
			},
			"virtual_network": []string{"virtual_network_id", "id"},
		},
		"data_factory": {
			"resource_group": []string{
				"resource_group_name", "name",
				"location", "location",
			},
			"data_factory": []string{
				"data_factory_name", "name",
				"data_factory_id", "id",
				"linked_service_name", "name",
				"integration_runtime_name", "name",
			},
			"databricks":      []string{"existing_cluster_id", "id"},
			"keyvault":        []string{"keyvault_id", "id"},
			"storage_account": []string{"storage_account_id", "id"},
		},
		"disk": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"dns": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"eventhub": {
			"resource_group": []string{"resource_group_name", "name"},
			"eventhub": []string{
				"eventhub_name", "name",
				"namespace_name", "name",
			},
		},
		"keyvault": {
			"resource_group": []string{
				"resource_group_name", "name",
				"location", "location",
			},
		},
		"load_balancer": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"network_interface": {
			"resource_group": []string{
				"resource_group_name", "name",
				"location", "location",
			},
			"subnet": []string{"subnet_id", "id"},
		},
		"network_security_group": {
			"resource_group": []string{
				"resource_group_name", "name",
				"location", "location",
			},
			"network_security_group": []string{"network_security_group_name", "name"},
		},
		"network_watcher": {
			"resource_group": []string{
				"resource_group_name", "name",
				"location", "location",
			},
			"network_watcher": []string{"network_watcher_name", "name"},
			"storage_account": []string{"storage_account_id", "id"},
		},
		"private_dns": {
			"resource_group":  []string{"resource_group_name", "name"},
			"virtual_network": []string{"virtual_network_id", "id"},
			"private_dns": []string{
				"zone_name", "name",
				"private_dns_zone_name", "name",
			},
		},
		"private_endpoint": {
			"resource_group": []string{
				"resource_group_name", "name",
				"location", "location",
			},
			"subnet": []string{"subnet_id", "id"},
		},
		"public_ip": {
			"resource_group": []string{
				"resource_group_name", "name",
				"location", "location",
			},
		},
		"purview": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"redis": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"route_table": {
			"resource_group": []string{
				"resource_group_name", "name",
				"location", "location",
			},
			"route_table": []string{"route_table_name", "name"},
		},
		"scaleset": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"ssh_public_key": {
			"resource_group": []string{
				"resource_group_name", "name",
				"location", "location",
			},
		},
		"storage_account": {
			"resource_group": []string{
				"resource_group_name", "name",
				"location", "location",
			},
			"virtual_network": []string{"virtual_network_subnet_ids", "id"},
		},
		"storage_blob": {
			"storage_account":   []string{"storage_account_name", "name"},
			"storage_container": []string{"storage_container_name", "name"},
		},
		"storage_container": {
			"storage_account": []string{"storage_account_name", "name"},
		},
		"synapse": {
			"resource_group": []string{
				"resource_group_name", "name",
				"managed_resource_group_name", "name",
			},
			"synapse": []string{"synapse_workspace_id", "id"},
		},
		"subnet": {
			"resource_group":         []string{"resource_group_name", "name"},
			"virtual_network":        []string{"virtual_network_name", "name"},
			"network_security_group": []string{"network_security_group_id", "id"},
			"route_table":            []string{"route_table_id", "id"},
			"subnet":                 []string{"subnet_id", "id"},
		},
		"virtual_machine": {
			"resource_group": []string{
				"resource_group_name", "name",
				"location", "location",
			},
			"network_interface": []string{"network_interface_ids", "id"},
		},
		"virtual_network": {
			"resource_group": []string{"resource_group_name", "name"},
		},
	}
}

func (p *AzureProvider) GetSupportedService() map[string]terraformutils.ServiceGenerator {
	return map[string]terraformutils.ServiceGenerator{
		"analysis":                             &AnalysisGenerator{},
		"apim":                                 &APIManagementGenerator{},
		"app_configuration":                    &AppConfigurationGenerator{},
		"app_service":                          &AppServiceGenerator{},
		"elastic":                              &ElasticGenerator{},
		"iotcentral":                           &IoTCentralGenerator{},
		"maintenance":                          &MaintenanceGenerator{},
		"redis_enterprise":                     &RedisEnterpriseGenerator{},
		"service_networking":                   &ServiceNetworkingGenerator{},
		"application_insights":                 &ApplicationInsightsGenerator{},
		"arc":                                  &ArcGenerator{},
		"arc_kubernetes":                       &ArcKubernetesGenerator{},
		"bastion":                              &BastionGenerator{},
		"batch":                                &BatchGenerator{},
		"data_share":                           &DataShareGenerator{},
		"dashboard":                            &DashboardGenerator{},
		"elastic_san":                          &ElasticSANGenerator{},
		"hdinsight":                            &HDInsightGenerator{},
		"healthcare":                           &HealthcareGenerator{},
		"hpc_cache":                            &HPCCacheGenerator{},
		"load_test":                            &LoadTestGenerator{},
		"private_dns_resolver":                 &PrivateDNSResolverGenerator{},
		"spring_cloud":                         &SpringCloudGenerator{},
		"digital_twins":                        &DigitalTwinsGenerator{},
		"maps":                                 &MapsGenerator{},
		"notification_hub":                     &NotificationHubGenerator{},
		"relay":                                &RelayGenerator{},
		"web_pubsub":                           &WebPubSubGenerator{},
		"attestation":                          &AttestationGenerator{},
		"automanage":                           &AutomanageGenerator{},
		"automation":                           &AutomationGenerator{},
		"cognitive":                            &CognitiveGenerator{},
		"data_protection":                      &DataProtectionGenerator{},
		"databox_edge":                         &DataBoxEdgeGenerator{},
		"database_migration":                   &DatabaseMigrationGenerator{},
		"dev_test_lab":                         &DevTestLabGenerator{},
		"device_update":                        &DeviceUpdateGenerator{},
		"lab_service":                          &LabServiceGenerator{},
		"kubernetes_fleet":                     &KubernetesFleetGenerator{},
		"virtual_desktop":                      &VirtualDesktopGenerator{},
		"datadog":                              &DatadogGenerator{},
		"dynatrace":                            &DynatraceGenerator{},
		"application_gateway":                  &ApplicationGatewayGenerator{},
		"cdn":                                  &CDNGenerator{},
		"chaos_studio":                         &ChaosStudioGenerator{},
		"communication":                        &CommunicationGenerator{},
		"confidential_ledger":                  &ConfidentialLedgerGenerator{},
		"cosmosdb":                             &CosmosDBGenerator{},
		"container":                            &ContainerGenerator{},
		"custom_provider":                      &CustomProviderGenerator{},
		"portal":                               &PortalGenerator{},
		"container_app":                        &ContainerAppGenerator{},
		"database":                             &DatabasesGenerator{},
		"databricks":                           &DatabricksGenerator{},
		"data_factory":                         &DataFactoryGenerator{},
		"ddos":                                 &DDoSGenerator{},
		"dev_center":                           &DevCenterGenerator{},
		"disk":                                 &DiskGenerator{},
		"dns":                                  &DNSGenerator{},
		"fluid_relay":                          &FluidRelayGenerator{},
		"eventgrid":                            &EventGridGenerator{},
		"eventhub":                             &EventHubGenerator{},
		"firewall":                             &FirewallGenerator{},
		"iothub":                               &IoTHubGenerator{},
		"keyvault":                             &KeyVaultGenerator{},
		"kubernetes":                           &KubernetesGenerator{},
		"kusto":                                &KustoGenerator{},
		"lighthouse":                           &LighthouseGenerator{},
		"load_balancer":                        &LoadBalancerGenerator{},
		"log_analytics":                        &LogAnalyticsGenerator{},
		"machine_learning":                     &MachineLearningGenerator{},
		"managed_identity":                     &ManagedIdentityGenerator{},
		"management_group":                     &ManagementGroupGenerator{},
		"management_lock":                      &ManagementLockGenerator{},
		"mobile_network":                       &MobileNetworkGenerator{},
		"monitor":                              &MonitorGenerator{},
		"mssql":                                &MSSQLGenerator{},
		"nat_gateway":                          &NatGatewayGenerator{},
		"netapp":                               &NetAppGenerator{},
		"new_relic":                            &NewRelicGenerator{},
		"nginx":                                &NginxGenerator{},
		"storage_mover":                        &StorageMoverGenerator{},
		"vmware":                               &VMwareGenerator{},
		"network_interface":                    &NetworkInterfaceGenerator{},
		"network_security_group":               &NetworkSecurityGroupGenerator{},
		"network_watcher":                      &NetworkWatcherGenerator{},
		"oracle":                               &OracleGenerator{},
		"orbital":                              &OrbitalGenerator{},
		"policy":                               &PolicyGenerator{},
		"private_dns":                          &PrivateDNSGenerator{},
		"private_endpoint":                     &PrivateEndpointGenerator{},
		"powerbi":                              &PowerBIGenerator{},
		"public_ip":                            &PublicIPGenerator{},
		"purview":                              &PurviewGenerator{},
		"recovery_services":                    &RecoveryServicesGenerator{},
		"redis":                                &RedisGenerator{},
		"resource_group":                       &ResourceGroupGenerator{},
		"role_assignment":                      &RoleAssignmentGenerator{},
		"route_table":                          &RouteTableGenerator{},
		"scaleset":                             &ScaleSetGenerator{},
		"search":                               &SearchGenerator{},
		"servicebus":                           &ServiceBusGenerator{},
		"signalr":                              &SignalRGenerator{},
		"security_center_contact":              &SecurityCenterContactGenerator{},
		"security_center_subscription_pricing": &SecurityCenterSubscriptionPricingGenerator{},
		"service_fabric":                       &ServiceFabricGenerator{},
		"spatial_anchors":                      &SpatialAnchorsGenerator{},
		"ssh_public_key":                       &SSHPublicKeyGenerator{},
		"storage_account":                      &StorageAccountGenerator{},
		"storage_blob":                         &StorageBlobGenerator{},
		"storage_container":                    &StorageContainerGenerator{},
		"stream_analytics":                     &StreamAnalyticsGenerator{},
		"synapse":                              &SynapseGenerator{},
		"subnet":                               &SubnetGenerator{},
		"traffic_manager":                      &TrafficManagerGenerator{},
		"virtual_machine":                      &VirtualMachineGenerator{},
		"voice_services":                       &VoiceServicesGenerator{},
		"virtual_network":                      &VirtualNetworkGenerator{},
		"virtual_wan":                          &VirtualWanGenerator{},
	}
}

func (p *AzureProvider) InitService(serviceName string, verbose bool) error {
	var isSupported bool
	if _, isSupported = p.GetSupportedService()[serviceName]; !isSupported {
		return errors.New("azurerm: " + serviceName + " not supported service")
	}
	p.Service = p.GetSupportedService()[serviceName]
	p.Service.SetName(serviceName)
	p.Service.SetVerbose(verbose)
	p.Service.SetProviderName(p.GetName())
	p.Service.SetArgs(map[string]interface{}{
		"config":           p.config,
		"token_credential": p.credential,
		"resource_group":   p.resourceGroup,
	})
	return nil
}
