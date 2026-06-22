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
	"testing"

	securitylaketypes "github.com/aws/aws-sdk-go-v2/service/securitylake/types"
)

// Per plan §8: generators whose InitResources does non-trivial result
// processing extract that logic into pure functions tested here (the sg_test.go
// convention — pure transforms, no SDK client mocking).

func TestNetworkmanagerAttachmentResourceType(t *testing.T) {
	cases := map[string]string{
		"CONNECT":                     "aws_networkmanager_connect_attachment",
		"SITE_TO_SITE_VPN":            "aws_networkmanager_site_to_site_vpn_attachment",
		"TRANSIT_GATEWAY_ROUTE_TABLE": "aws_networkmanager_transit_gateway_route_table_attachment",
		"VPC":                         "aws_networkmanager_vpc_attachment",
		"UNKNOWN":                     "", // no dedicated resource → skipped
		"":                            "",
	}
	for in, want := range cases {
		if got := networkmanagerAttachmentResourceType(in); got != want {
			t.Errorf("networkmanagerAttachmentResourceType(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNaclRuleImportID(t *testing.T) {
	ptr := func(i int32) *int32 { return &i }
	bptr := func(b bool) *bool { return &b }

	if _, ok := naclRuleImportID("acl-1", nil, "6", bptr(false)); ok {
		t.Error("nil rule number must be skipped")
	}
	if _, ok := naclRuleImportID("acl-1", ptr(32767), "-1", bptr(false)); ok {
		t.Error("implicit 32767 default-deny rule must be skipped")
	}
	id, ok := naclRuleImportID("acl-1", ptr(100), "6", bptr(false))
	if !ok || id != "acl-1:100:6:false" {
		t.Errorf("ingress rule = %q (ok=%v), want acl-1:100:6:false", id, ok)
	}
	id, ok = naclRuleImportID("acl-2", ptr(200), "17", bptr(true))
	if !ok || id != "acl-2:200:17:true" {
		t.Errorf("egress rule = %q (ok=%v), want acl-2:200:17:true", id, ok)
	}
	// nil egress pointer defaults to false.
	id, _ = naclRuleImportID("acl-3", ptr(1), "6", nil)
	if id != "acl-3:1:6:false" {
		t.Errorf("nil egress = %q, want acl-3:1:6:false", id)
	}
}

func TestIsServiceLinkedRolePath(t *testing.T) {
	if !isServiceLinkedRolePath("/aws-service-role/elasticache.amazonaws.com/AWSServiceRoleForElastiCache") {
		t.Error("service-role path should be detected")
	}
	if isServiceLinkedRolePath("/") {
		t.Error("plain root path is not service-linked")
	}
	if isServiceLinkedRolePath("/my/custom/path/") {
		t.Error("custom path is not service-linked")
	}
}

func TestSecurityLakeLogSource(t *testing.T) {
	seen := map[string]bool{}

	aws1 := &securitylaketypes.LogSourceResourceMemberAwsLogSource{
		Value: securitylaketypes.AwsLogSourceResource{SourceName: securitylaketypes.AwsLogSourceName("ROUTE53")},
	}
	tfType, name, ok := securityLakeLogSource(aws1, seen)
	if !ok || tfType != "aws_securitylake_aws_log_source" || name != "ROUTE53" {
		t.Fatalf("aws log source = (%q,%q,%v), want (aws_securitylake_aws_log_source,ROUTE53,true)", tfType, name, ok)
	}
	// Same source seen again (different account/region) → deduped.
	if _, _, ok := securityLakeLogSource(aws1, seen); ok {
		t.Error("duplicate aws log source must be deduped")
	}

	custom := &securitylaketypes.LogSourceResourceMemberCustomLogSource{
		Value: securitylaketypes.CustomLogSourceResource{SourceName: stringPtr("my-custom")},
	}
	tfType, name, ok = securityLakeLogSource(custom, seen)
	if !ok || tfType != "aws_securitylake_custom_log_source" || name != "my-custom" {
		t.Fatalf("custom log source = (%q,%q,%v), want (aws_securitylake_custom_log_source,my-custom,true)", tfType, name, ok)
	}

	// Empty name → skipped.
	emptyName := &securitylaketypes.LogSourceResourceMemberAwsLogSource{}
	if _, _, ok := securityLakeLogSource(emptyName, seen); ok {
		t.Error("empty source name must be skipped")
	}
}

func stringPtr(s string) *string { return &s }
