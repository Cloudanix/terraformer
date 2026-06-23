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

func TestAppEngineVersionType(t *testing.T) {
	cases := map[string]string{
		"flexible": "google_app_engine_flexible_app_version",
		"flex":     "google_app_engine_flexible_app_version",
		"standard": "google_app_engine_standard_app_version",
		"":         "google_app_engine_standard_app_version",
		"unknown":  "google_app_engine_standard_app_version",
	}
	for in, want := range cases {
		if got := appEngineVersionType(in); got != want {
			t.Errorf("appEngineVersionType(%q) = %q, want %q", in, got, want)
		}
	}
}
