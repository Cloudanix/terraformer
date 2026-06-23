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

// AUTO-GENERATED CODE. DO NOT EDIT.
package gcp

import (
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// Map of supported GCP compute service with code generate
var ComputeServices = map[string]terraformutils.ServiceGenerator{

	"addresses":                     &GCPFacade{service: &AddressesGenerator{}},
	"autoscalers":                   &GCPFacade{service: &AutoscalersGenerator{}},
	"backendBuckets":                &GCPFacade{service: &BackendBucketsGenerator{}},
	"backendServices":               &GCPFacade{service: &BackendServicesGenerator{}},
	"crossSiteNetworks":             &GCPFacade{service: &CrossSiteNetworksGenerator{}},
	"disks":                         &GCPFacade{service: &DisksGenerator{}},
	"externalVpnGateways":           &GCPFacade{service: &ExternalVpnGatewaysGenerator{}},
	"firewall":                      &GCPFacade{service: &FirewallGenerator{}},
	"forwardingRules":               &GCPFacade{service: &ForwardingRulesGenerator{}},
	"globalAddresses":               &GCPFacade{service: &GlobalAddressesGenerator{}},
	"globalForwardingRules":         &GCPFacade{service: &GlobalForwardingRulesGenerator{}},
	"globalNetworkEndpointGroups":   &GCPFacade{service: &GlobalNetworkEndpointGroupsGenerator{}},
	"healthChecks":                  &GCPFacade{service: &HealthChecksGenerator{}},
	"httpHealthChecks":              &GCPFacade{service: &HttpHealthChecksGenerator{}},
	"httpsHealthChecks":             &GCPFacade{service: &HttpsHealthChecksGenerator{}},
	"images":                        &GCPFacade{service: &ImagesGenerator{}},
	"instanceGroupManagers":         &GCPFacade{service: &InstanceGroupManagersGenerator{}},
	"instanceGroups":                &GCPFacade{service: &InstanceGroupsGenerator{}},
	"instanceTemplates":             &GCPFacade{service: &InstanceTemplatesGenerator{}},
	"instantSnapshots":              &GCPFacade{service: &InstantSnapshotsGenerator{}},
	"interconnectAttachments":       &GCPFacade{service: &InterconnectAttachmentsGenerator{}},
	"interconnectGroups":            &GCPFacade{service: &InterconnectGroupsGenerator{}},
	"interconnects":                 &GCPFacade{service: &InterconnectsGenerator{}},
	"machineImages":                 &GCPFacade{service: &MachineImagesGenerator{}},
	"networkAttachments":            &GCPFacade{service: &NetworkAttachmentsGenerator{}},
	"networkEndpointGroups":         &GCPFacade{service: &NetworkEndpointGroupsGenerator{}},
	"networkFirewallPolicies":       &GCPFacade{service: &NetworkFirewallPoliciesGenerator{}},
	"networks":                      &GCPFacade{service: &NetworksGenerator{}},
	"nodeGroups":                    &GCPFacade{service: &NodeGroupsGenerator{}},
	"nodeTemplates":                 &GCPFacade{service: &NodeTemplatesGenerator{}},
	"packetMirrorings":              &GCPFacade{service: &PacketMirroringsGenerator{}},
	"publicAdvertisedPrefixes":      &GCPFacade{service: &PublicAdvertisedPrefixesGenerator{}},
	"publicDelegatedPrefixes":       &GCPFacade{service: &PublicDelegatedPrefixesGenerator{}},
	"regionAutoscalers":             &GCPFacade{service: &RegionAutoscalersGenerator{}},
	"regionBackendServices":         &GCPFacade{service: &RegionBackendServicesGenerator{}},
	"regionCommitments":             &GCPFacade{service: &RegionCommitmentsGenerator{}},
	"regionDisks":                   &GCPFacade{service: &RegionDisksGenerator{}},
	"regionHealthChecks":            &GCPFacade{service: &RegionHealthChecksGenerator{}},
	"regionInstanceGroupManagers":   &GCPFacade{service: &RegionInstanceGroupManagersGenerator{}},
	"regionInstanceGroups":          &GCPFacade{service: &RegionInstanceGroupsGenerator{}},
	"regionInstanceTemplates":       &GCPFacade{service: &RegionInstanceTemplatesGenerator{}},
	"regionInstantSnapshots":        &GCPFacade{service: &RegionInstantSnapshotsGenerator{}},
	"regionNetworkEndpointGroups":   &GCPFacade{service: &RegionNetworkEndpointGroupsGenerator{}},
	"regionNetworkFirewallPolicies": &GCPFacade{service: &RegionNetworkFirewallPoliciesGenerator{}},
	"regionSecurityPolicies":        &GCPFacade{service: &RegionSecurityPoliciesGenerator{}},
	"regionSslCertificates":         &GCPFacade{service: &RegionSslCertificatesGenerator{}},
	"regionSslPolicies":             &GCPFacade{service: &RegionSslPoliciesGenerator{}},
	"regionTargetHttpProxies":       &GCPFacade{service: &RegionTargetHttpProxiesGenerator{}},
	"regionTargetHttpsProxies":      &GCPFacade{service: &RegionTargetHttpsProxiesGenerator{}},
	"regionTargetTcpProxies":        &GCPFacade{service: &RegionTargetTcpProxiesGenerator{}},
	"regionUrlMaps":                 &GCPFacade{service: &RegionUrlMapsGenerator{}},
	"reservations":                  &GCPFacade{service: &ReservationsGenerator{}},
	"resourcePolicies":              &GCPFacade{service: &ResourcePoliciesGenerator{}},
	"routers":                       &GCPFacade{service: &RoutersGenerator{}},
	"routes":                        &GCPFacade{service: &RoutesGenerator{}},
	"securityPolicies":              &GCPFacade{service: &SecurityPoliciesGenerator{}},
	"serviceAttachments":            &GCPFacade{service: &ServiceAttachmentsGenerator{}},
	"snapshots":                     &GCPFacade{service: &SnapshotsGenerator{}},
	"sslCertificates":               &GCPFacade{service: &SslCertificatesGenerator{}},
	"sslPolicies":                   &GCPFacade{service: &SslPoliciesGenerator{}},
	"storagePools":                  &GCPFacade{service: &StoragePoolsGenerator{}},
	"subnetworks":                   &GCPFacade{service: &SubnetworksGenerator{}},
	"targetGrpcProxies":             &GCPFacade{service: &TargetGrpcProxiesGenerator{}},
	"targetHttpProxies":             &GCPFacade{service: &TargetHttpProxiesGenerator{}},
	"targetHttpsProxies":            &GCPFacade{service: &TargetHttpsProxiesGenerator{}},
	"targetInstances":               &GCPFacade{service: &TargetInstancesGenerator{}},
	"targetPools":                   &GCPFacade{service: &TargetPoolsGenerator{}},
	"targetSslProxies":              &GCPFacade{service: &TargetSslProxiesGenerator{}},
	"targetTcpProxies":              &GCPFacade{service: &TargetTcpProxiesGenerator{}},
	"targetVpnGateways":             &GCPFacade{service: &TargetVpnGatewaysGenerator{}},
	"urlMaps":                       &GCPFacade{service: &UrlMapsGenerator{}},
	"vpnGateways":                   &GCPFacade{service: &VpnGatewaysGenerator{}},
	"vpnTunnels":                    &GCPFacade{service: &VpnTunnelsGenerator{}},
}
