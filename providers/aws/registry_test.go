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

import "testing"

// TestAllServicesInstantiable exercises every registered service generator: the
// facade and its wrapped generator must be non-nil and accept the same
// name/args/provider wiring that cmd/import.go applies before InitResources.
// This is the per-service smoke test that complements TestServiceScope
// (region binding) and TestEveryServiceDocumented (docs) — every service added
// in the coverage work passes through all three. The live List/Describe path
// itself cannot be unit-tested without mocking the AWS SDK (which the codebase
// deliberately avoids) or launching the provider plugin (blocked: go-plugin
// needs a unix-socket bind); that path is covered by the integration round-trip.
func TestAllServicesInstantiable(t *testing.T) {
	provider := &AWSProvider{}
	services := provider.GetSupportedService()
	if len(services) == 0 {
		t.Fatal("no services registered")
	}
	for name, svc := range services {
		if svc == nil {
			t.Errorf("service %q registered as nil", name)
			continue
		}
		// Mirror cmd/import.go's per-service setup; must not panic for any service.
		svc.SetName(name)
		svc.SetProviderName("aws")
		svc.SetVerbose(false)
		svc.SetArgs(map[string]interface{}{
			"region":                 "us-east-1",
			"profile":                "",
			"skip_region_validation": true,
		})
		if got := svc.GetArgs()["region"]; got != "us-east-1" {
			t.Errorf("service %q: args not wired, region=%v", name, got)
		}
	}
}
