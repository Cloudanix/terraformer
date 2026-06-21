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
	"sync"
	"testing"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
)

type fakeItem struct {
	id   string
	name string
}

func idOf(i fakeItem) string   { return i.id }
func nameOf(i fakeItem) string { return i.name }

func TestAppendSimpleResources(t *testing.T) {
	items := []fakeItem{
		{id: "id-1", name: "first"},
		{id: "", name: "skip-me"}, // empty import ID → skipped
		{id: "id-2", name: "second"},
	}

	got := appendSimpleResources(nil, items, "aws_thing", defaultAllowEmptyValues, idOf, nameOf)

	if len(got) != 2 {
		t.Fatalf("expected 2 resources (empty-ID item skipped), got %d", len(got))
	}
	if got[0].InstanceState.ID != "id-1" || got[1].InstanceState.ID != "id-2" {
		t.Errorf("unexpected import IDs: %q, %q", got[0].InstanceState.ID, got[1].InstanceState.ID)
	}
	if got[0].InstanceInfo.Type != "aws_thing" {
		t.Errorf("unexpected resource type: %q", got[0].InstanceInfo.Type)
	}
	// ResourceName is TfSanitize(name) — terraformer prefixes "tfer--".
	if got[0].ResourceName != "tfer--first" || got[1].ResourceName != "tfer--second" {
		t.Errorf("unexpected resource names: %q, %q", got[0].ResourceName, got[1].ResourceName)
	}
}

func TestAppendSimpleResourcesAppendsToExisting(t *testing.T) {
	seed := appendSimpleResources(nil, []fakeItem{{id: "a", name: "a"}}, "aws_thing", defaultAllowEmptyValues, idOf, nameOf)
	got := appendSimpleResources(seed, []fakeItem{{id: "b", name: "b"}}, "aws_thing", defaultAllowEmptyValues, idOf, nameOf)
	if len(got) != 2 {
		t.Fatalf("expected append to existing slice (2), got %d", len(got))
	}
}

func TestAppendSimpleResourcesEmpty(t *testing.T) {
	got := appendSimpleResources([]terraformutils.Resource{}, []fakeItem{}, "aws_thing", defaultAllowEmptyValues, idOf, nameOf)
	if len(got) != 0 {
		t.Fatalf("expected 0 resources for empty input, got %d", len(got))
	}
}

func policyResource(resourceType, policy string) terraformutils.Resource {
	r := terraformutils.NewSimpleResource("id", "name", resourceType, "aws", defaultAllowEmptyValues)
	r.Item = map[string]any{"policy": policy}
	return r
}

func TestWrapPolicyAttribute(t *testing.T) {
	s := &AWSService{}
	resources := []terraformutils.Resource{
		policyResource("aws_ecr_repository_policy", `{"Statement":"x"}`),
		policyResource("aws_ecr_lifecycle_policy", `{"rules":1}`),
		policyResource("aws_other_type", `{"untouched":true}`), // type not listed → untouched
	}
	// non-string policy attr → skipped, no panic
	nonString := terraformutils.NewSimpleResource("id", "name", "aws_ecr_repository_policy", "aws", defaultAllowEmptyValues)
	nonString.Item = map[string]any{"policy": 42}
	resources = append(resources, nonString)
	// missing policy attr → skipped, no panic
	missing := terraformutils.NewSimpleResource("id", "name", "aws_ecr_repository_policy", "aws", defaultAllowEmptyValues)
	missing.Item = map[string]any{}
	resources = append(resources, missing)

	s.wrapPolicyAttribute(resources, "policy", "aws_ecr_repository_policy", "aws_ecr_lifecycle_policy")

	want0 := "<<POLICY\n{\"Statement\":\"x\"}\nPOLICY"
	if resources[0].Item["policy"] != want0 {
		t.Errorf("repository_policy not wrapped:\n got %q\nwant %q", resources[0].Item["policy"], want0)
	}
	if resources[1].Item["policy"] != "<<POLICY\n{\"rules\":1}\nPOLICY" {
		t.Errorf("lifecycle_policy not wrapped: %q", resources[1].Item["policy"])
	}
	if resources[2].Item["policy"] != `{"untouched":true}` {
		t.Errorf("non-listed type should be untouched, got %q", resources[2].Item["policy"])
	}
	if resources[3].Item["policy"] != 42 {
		t.Errorf("non-string policy should be untouched, got %v", resources[3].Item["policy"])
	}
}

func TestWrapPolicyAttributeEscapesInterpolation(t *testing.T) {
	s := &AWSService{}
	// ${aws:username} must be escaped to $${aws:username} so terraform does not
	// treat it as an interpolation.
	resources := []terraformutils.Resource{policyResource("aws_ecr_repository_policy", "${aws:username}")}
	s.wrapPolicyAttribute(resources, "policy", "aws_ecr_repository_policy")
	want := "<<POLICY\n$${aws:username}\nPOLICY"
	if resources[0].Item["policy"] != want {
		t.Errorf("interpolation not escaped:\n got %q\nwant %q", resources[0].Item["policy"], want)
	}
}

// resetConfigCache clears the package-level per-region cache and restores it
// after the test so cache tests don't pollute each other.
func resetConfigCache(t *testing.T) {
	t.Helper()
	configCacheMu.Lock()
	prev := configCache
	configCache = map[string]*aws.Config{}
	configCacheMu.Unlock()
	t.Cleanup(func() {
		configCacheMu.Lock()
		configCache = prev
		configCacheMu.Unlock()
	})
}

func newServiceForRegion(region string) *AWSService {
	s := &AWSService{}
	s.SetArgs(map[string]any{"region": region, "profile": ""})
	return s
}

// Exercises the cache-HIT path only (seeded), so no live AWS credentials/network
// are touched — generateConfig returns the cached config before any SDK call.
func TestGenerateConfigCacheHitPerRegion(t *testing.T) {
	resetConfigCache(t)

	configCache["us-east-1"] = &aws.Config{Region: "us-east-1"}
	configCache["eu-west-1"] = &aws.Config{Region: "eu-west-1"}

	east := newServiceForRegion("us-east-1")
	cfg, err := east.generateConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Region != "us-east-1" {
		t.Errorf("same region should return its cached config, got region %q", cfg.Region)
	}

	euCfg, err := newServiceForRegion("eu-west-1").generateConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if euCfg.Region != "eu-west-1" {
		t.Errorf("different region must return a distinct config, got region %q", euCfg.Region)
	}
}

// Run with -race: many concurrent generateConfig calls for a seeded region must
// be safe (the cache is mutex-guarded).
func TestGenerateConfigCacheConcurrent(t *testing.T) {
	resetConfigCache(t)
	configCache["ap-south-1"] = &aws.Config{Region: "ap-south-1"}

	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s := newServiceForRegion("ap-south-1")
			if _, err := s.generateConfig(); err != nil {
				t.Errorf("concurrent generateConfig error: %v", err)
			}
		}()
	}
	wg.Wait()
}
