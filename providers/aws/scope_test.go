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
	"sort"
	"testing"
)

// Every registered service must have exactly one serviceScope entry, and the
// scopeGlobal / scopeEastOnly sets must match the lists the import passes
// actually consult. A new service added to GetSupportedService without a scope
// (or misclassified) fails here instead of importing into the wrong region path.
func TestServiceScopeMatchesRegistry(t *testing.T) {
	registered := (&AWSProvider{}).GetSupportedService()

	// 1. completeness: every registered service is classified.
	for name := range registered {
		if _, ok := serviceScope[name]; !ok {
			t.Errorf("service %q is registered but missing from serviceScope", name)
		}
	}
	// 2. no stale entries: every serviceScope key is a real service.
	for name := range serviceScope {
		if _, ok := registered[name]; !ok {
			t.Errorf("serviceScope has %q but it is not a registered service", name)
		}
	}

	// 3. scopeGlobal set == SupportedGlobalResources.
	assertSetsEqual(t, "global", scopeKeys(scopeGlobal), SupportedGlobalResources)
	// 4. scopeEastOnly set == SupportedEastOnlyResources.
	assertSetsEqual(t, "eastOnly", scopeKeys(scopeEastOnly), SupportedEastOnlyResources)
}

func scopeKeys(want regionScope) []string {
	var keys []string
	for name, s := range serviceScope {
		if s == want {
			keys = append(keys, name)
		}
	}
	return keys
}

func assertSetsEqual(t *testing.T, label string, got, want []string) {
	t.Helper()
	gotCopy := append([]string(nil), got...)
	wantCopy := append([]string(nil), want...)
	sort.Strings(gotCopy)
	sort.Strings(wantCopy)
	if len(gotCopy) != len(wantCopy) {
		t.Errorf("%s scope mismatch: serviceScope has %v, list has %v", label, gotCopy, wantCopy)
		return
	}
	for i := range gotCopy {
		if gotCopy[i] != wantCopy[i] {
			t.Errorf("%s scope mismatch: serviceScope has %v, list has %v", label, gotCopy, wantCopy)
			return
		}
	}
}
