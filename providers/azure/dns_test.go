// Copyright 2020 The Terraformer Authors.
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

func TestDNSRecordResourceType(t *testing.T) {
	cases := map[string]string{
		"Microsoft.Network/dnszones/A":     "azurerm_dns_a_record",
		"Microsoft.Network/dnszones/AAAA":  "azurerm_dns_aaaa_record",
		"Microsoft.Network/dnszones/CNAME": "azurerm_dns_cname_record",
		"Microsoft.Network/dnszones/TXT":   "azurerm_dns_txt_record",
		"Microsoft.Network/dnszones/SOA":   "", // not imported
		"weird":                            "",
	}
	for in, want := range cases {
		if got := dnsRecordResourceType(in); got != want {
			t.Errorf("dnsRecordResourceType(%q) = %q, want %q", in, got, want)
		}
	}
}
