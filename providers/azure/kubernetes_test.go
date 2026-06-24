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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v6"
)

func agentPool(name string, mode *armcontainerservice.AgentPoolMode) *armcontainerservice.AgentPool {
	n := name
	p := &armcontainerservice.AgentPool{Name: &n}
	if mode != nil {
		p.Properties = &armcontainerservice.ManagedClusterAgentPoolProfileProperties{Mode: mode}
	}
	return p
}

func TestUserAgentPools(t *testing.T) {
	system := armcontainerservice.AgentPoolModeSystem
	user := armcontainerservice.AgentPoolModeUser
	pools := []*armcontainerservice.AgentPool{
		nil,                          // skipped
		agentPool("nomode", nil),     // no Properties -> skipped
		agentPool("system", &system), // System (inline default_node_pool) -> skipped
		agentPool("user1", &user),    // kept
		agentPool("user2", &user),    // kept
	}
	got := userAgentPools(pools)
	if len(got) != 2 {
		t.Fatalf("got %d user pools, want 2", len(got))
	}
	if *got[0].Name != "user1" || *got[1].Name != "user2" {
		t.Errorf("unexpected pools: %q, %q", *got[0].Name, *got[1].Name)
	}
}
