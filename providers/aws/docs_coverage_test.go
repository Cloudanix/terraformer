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
	"os"
	"strings"
	"testing"
)

// TestEveryServiceDocumented asserts every registered AWS service key appears in
// docs/aws.md (as a "`key`" bullet). This is the cross-cutting regression test
// for the coverage work: adding a generator without documenting it fails here,
// just as omitting its serviceScope entry fails TestServiceScope.
func TestEveryServiceDocumented(t *testing.T) {
	data, err := os.ReadFile("../../docs/aws.md")
	if err != nil {
		t.Fatalf("read docs/aws.md: %v", err)
	}
	docs := string(data)

	provider := &AWSProvider{}
	var missing []string
	for name := range provider.GetSupportedService() {
		if !strings.Contains(docs, "`"+name+"`") {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		t.Errorf("%d registered services missing from docs/aws.md: %v", len(missing), missing)
	}
}
