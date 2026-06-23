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

func TestWafv2APIKeyID(t *testing.T) {
	if got := wafv2APIKeyID("abc123==", "REGIONAL"); got != "abc123==,REGIONAL" {
		t.Errorf("got %q", got)
	}
	if got := wafv2APIKeyID("k", "CLOUDFRONT"); got != "k,CLOUDFRONT" {
		t.Errorf("got %q", got)
	}
}
