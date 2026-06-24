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

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armpolicy"
)

func TestIsCustomPolicy(t *testing.T) {
	custom := armpolicy.PolicyTypeCustom
	builtIn := armpolicy.PolicyTypeBuiltIn
	cases := []struct {
		name string
		in   *armpolicy.PolicyType
		want bool
	}{
		{"nil", nil, false},
		{"custom", &custom, true},
		{"builtin", &builtIn, false},
	}
	for _, c := range cases {
		if got := isCustomPolicy(c.in); got != c.want {
			t.Errorf("isCustomPolicy(%s) = %v, want %v", c.name, got, c.want)
		}
	}
}
