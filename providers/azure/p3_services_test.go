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
	"reflect"
	"testing"
)

// assertRegistered asserts a service key resolves to the expected generator type.
func assertRegistered(t *testing.T, key string, want interface{}) {
	t.Helper()
	p := &AzureProvider{}
	svc, ok := p.GetSupportedService()[key]
	if !ok {
		t.Fatalf("service %q not registered in GetSupportedService", key)
	}
	if got, wantT := reflect.TypeOf(svc), reflect.TypeOf(want); got != wantT {
		t.Errorf("service %q registered as %s, want %s", key, got, wantT)
	}
}

func TestP3ServicesRegistered(t *testing.T) {
	assertRegistered(t, "automanage", &AutomanageGenerator{})
	assertRegistered(t, "spatial_anchors", &SpatialAnchorsGenerator{})
	assertRegistered(t, "lighthouse", &LighthouseGenerator{})
	assertRegistered(t, "hdinsight", &HDInsightGenerator{})
	assertRegistered(t, "service_fabric", &ServiceFabricGenerator{})
	assertRegistered(t, "arc", &ArcGenerator{})
	assertRegistered(t, "databox_edge", &DataBoxEdgeGenerator{})
	assertRegistered(t, "orbital", &OrbitalGenerator{})
	assertRegistered(t, "oracle", &OracleGenerator{})
	assertRegistered(t, "mobile_network", &MobileNetworkGenerator{})
	assertRegistered(t, "voice_services", &VoiceServicesGenerator{})
	assertRegistered(t, "new_relic", &NewRelicGenerator{})
	assertRegistered(t, "hpc_cache", &HPCCacheGenerator{})
	assertRegistered(t, "app_configuration", &AppConfigurationGenerator{})
	assertRegistered(t, "iotcentral", &IoTCentralGenerator{})
	assertRegistered(t, "elastic", &ElasticGenerator{})
	assertRegistered(t, "redis_enterprise", &RedisEnterpriseGenerator{})
	assertRegistered(t, "maintenance", &MaintenanceGenerator{})
	assertRegistered(t, "service_networking", &ServiceNetworkingGenerator{})
}
