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

import (
	"testing"

	"google.golang.org/api/compute/v1"
)

func TestExpandComputeBindings(t *testing.T) {
	cases := []struct {
		name     string
		bindings []*compute.Binding
		want     []computeIamRoleMember
	}{
		{
			name:     "nil bindings",
			bindings: nil,
			want:     []computeIamRoleMember{},
		},
		{
			name: "single role multiple members flattens per member",
			bindings: []*compute.Binding{
				{Role: "roles/compute.admin", Members: []string{"user:a@x.com", "user:b@x.com"}},
			},
			want: []computeIamRoleMember{
				{Role: "roles/compute.admin", Member: "user:a@x.com"},
				{Role: "roles/compute.admin", Member: "user:b@x.com"},
			},
		},
		{
			name: "multiple roles preserve order and skip nil binding",
			bindings: []*compute.Binding{
				{Role: "roles/viewer", Members: []string{"user:c@x.com"}},
				nil,
				{Role: "roles/editor", Members: []string{"serviceAccount:s@x.iam.gserviceaccount.com"}},
			},
			want: []computeIamRoleMember{
				{Role: "roles/viewer", Member: "user:c@x.com"},
				{Role: "roles/editor", Member: "serviceAccount:s@x.iam.gserviceaccount.com"},
			},
		},
		{
			name: "role with no members yields nothing",
			bindings: []*compute.Binding{
				{Role: "roles/owner", Members: nil},
			},
			want: []computeIamRoleMember{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := expandComputeBindings(tc.bindings)
			if len(got) != len(tc.want) {
				t.Fatalf("len = %d, want %d (%+v)", len(got), len(tc.want), got)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Errorf("[%d] = %+v, want %+v", i, got[i], tc.want[i])
				}
			}
		})
	}
}
