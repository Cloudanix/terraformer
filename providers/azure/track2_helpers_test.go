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
	"context"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
)

func strptr(s string) *string { return &s }

func TestValueOrEmpty(t *testing.T) {
	if got := valueOrEmpty(nil); got != "" {
		t.Errorf("valueOrEmpty(nil) = %q, want empty", got)
	}
	if got := valueOrEmpty(strptr("x")); got != "x" {
		t.Errorf("valueOrEmpty(*x) = %q, want x", got)
	}
}

func TestCloudConfig(t *testing.T) {
	cases := map[string]cloud.Configuration{
		"":                      cloud.AzurePublic,
		"public":                cloud.AzurePublic,
		"usgovernment":          cloud.AzureGovernment,
		"USGovernmentCloud":     cloud.AzureGovernment,
		"china":                 cloud.AzureChina,
		"AzureChinaCloud":       cloud.AzureChina,
		"something-unknown-env": cloud.AzurePublic,
	}
	for env, want := range cases {
		if got := cloudConfig(env); got.ActiveDirectoryAuthorityHost != want.ActiveDirectoryAuthorityHost {
			t.Errorf("cloudConfig(%q) host = %q, want %q", env, got.ActiveDirectoryAuthorityHost, want.ActiveDirectoryAuthorityHost)
		}
	}
}

func TestResourceGroups(t *testing.T) {
	cases := []struct {
		raw  string
		want []string
	}{
		{"", nil},
		{"rg1", []string{"rg1"}},
		{"rg1:rg2:rg3", []string{"rg1", "rg2", "rg3"}},
		{" rg1 : : rg2 ", []string{"rg1", "rg2"}}, // trims + drops empties
	}
	for _, c := range cases {
		az := &AzureService{}
		az.SetArgs(map[string]interface{}{"resource_group": c.raw})
		got := az.resourceGroups()
		if len(got) != len(c.want) {
			t.Fatalf("resourceGroups(%q) = %v, want %v", c.raw, got, c.want)
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("resourceGroups(%q)[%d] = %q, want %q", c.raw, i, got[i], c.want[i])
			}
		}
	}
}

// --- appendFromPager ---

type fakeItem struct {
	ID   *string
	Name *string
}
type fakePage struct {
	Value []*fakeItem
}

func pagerOf(pages []fakePage) *runtime.Pager[fakePage] {
	idx := -1
	return runtime.NewPager(runtime.PagingHandler[fakePage]{
		More: func(fakePage) bool { return idx < len(pages)-1 },
		Fetcher: func(context.Context, *fakePage) (fakePage, error) {
			idx++
			return pages[idx], nil
		},
	})
}

func drain(t *testing.T, pages []fakePage) *AzureService {
	t.Helper()
	az := &AzureService{}
	az.SetProviderName("azurerm")
	err := appendFromPager(az, pagerOf(pages),
		func(p fakePage) []*fakeItem { return p.Value },
		func(i *fakeItem) string { return valueOrEmpty(i.ID) },
		func(i *fakeItem) string { return valueOrEmpty(i.Name) },
		"azurerm_test")
	if err != nil {
		t.Fatalf("appendFromPager error: %v", err)
	}
	return az
}

func TestAppendFromPagerEmpty(t *testing.T) {
	az := drain(t, []fakePage{{Value: nil}})
	if len(az.Resources) != 0 {
		t.Errorf("empty page produced %d resources, want 0", len(az.Resources))
	}
}

func TestAppendFromPagerSkipsNilAndEmptyID(t *testing.T) {
	pages := []fakePage{
		{Value: []*fakeItem{
			nil,                                              // nil item skipped
			{ID: strptr(""), Name: strptr("noid")},          // empty id skipped
			{ID: strptr("/sub/x/a"), Name: strptr("alpha")}, // kept
		}},
		{Value: []*fakeItem{
			{ID: strptr("/sub/x/b"), Name: nil}, // kept, nil name -> ""
		}},
	}
	az := drain(t, pages)
	if len(az.Resources) != 2 {
		t.Fatalf("got %d resources, want 2", len(az.Resources))
	}
	if az.Resources[0].InstanceState.ID != "/sub/x/a" {
		t.Errorf("resource[0] id = %q, want /sub/x/a", az.Resources[0].InstanceState.ID)
	}
	if az.Resources[1].InstanceState.ID != "/sub/x/b" {
		t.Errorf("resource[1] id = %q, want /sub/x/b", az.Resources[1].InstanceState.ID)
	}
}
