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

package gcp

import "testing"

func TestDiscoveryEngineEngineType(t *testing.T) {
	cases := map[string]string{
		"SOLUTION_TYPE_CHAT":            "google_discovery_engine_chat_engine",
		"SOLUTION_TYPE_RECOMMENDATION":  "google_discovery_engine_recommendation_engine",
		"SOLUTION_TYPE_SEARCH":          "google_discovery_engine_search_engine",
		"":                              "google_discovery_engine_search_engine",
		"SOLUTION_TYPE_GENERATIVE_CHAT": "google_discovery_engine_search_engine",
	}
	for in, want := range cases {
		if got := discoveryEngineEngineType(in); got != want {
			t.Errorf("discoveryEngineEngineType(%q) = %q, want %q", in, got, want)
		}
	}
}
