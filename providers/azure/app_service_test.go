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

import "testing"

func TestSiteResourceType(t *testing.T) {
	cases := map[string]string{
		"app":               "azurerm_windows_web_app",
		"":                  "azurerm_windows_web_app",
		"app,linux":         "azurerm_linux_web_app",
		"linux":             "azurerm_linux_web_app",
		"functionapp":       "azurerm_windows_function_app",
		"functionapp,linux": "azurerm_linux_function_app",
		"FunctionApp,Linux": "azurerm_linux_function_app", // case-insensitive
	}
	for kind, want := range cases {
		if got := siteResourceType(kind); got != want {
			t.Errorf("siteResourceType(%q) = %q, want %q", kind, got, want)
		}
	}
}
