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
	"strings"
	"testing"
)

func TestRDSCustomEngineVersionIDFormat(t *testing.T) {
	if got := rdsCustomEngineVersionID("custom-oracle-ee", "19.cdb_cev1"); got != "custom-oracle-ee:19.cdb_cev1" {
		t.Errorf("got %q", got)
	}
}

// All custom engine families must carry the custom- prefix; a stray non-custom
// entry would import AWS-managed engine versions as if they were customer CEVs.
func TestCustomRDSEnginesArePrefixed(t *testing.T) {
	for _, e := range customRDSEngines {
		if !strings.HasPrefix(e, "custom-") {
			t.Errorf("engine %q lacks custom- prefix", e)
		}
	}
}
