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

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// valueOrEmpty dereferences a Track 2 *string safely. Track 2 SDK fields are
// pointers; a nil optional field deref panics and kills the whole import run.
// Always use this, never *p.Field.
func valueOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// appendFromPager drains a Track 2 pager and appends each item as a simple
// resource. It is the Track 2 analogue of the hand-rolled
// pager->item->NewSimpleResource loop and collapses a new simple service to a
// few lines. Nil items and items with an empty import ID are skipped.
//
//	P = the pager's page/response type (e.g. armcompute.DisksClientListResponse)
//	I = the item type    (e.g. armcompute.Disk)
//
// values extracts the page's slice (usually `func(p P) []*I { return p.Value }`),
// id/name extract the ARM ID and display name (use valueOrEmpty).
func appendFromPager[P any, I any](
	az *AzureService,
	pager *runtime.Pager[P],
	values func(P) []*I,
	id func(*I) string,
	name func(*I) string,
	tfType string,
) error {
	for pager.More() {
		page, err := pager.NextPage(context.TODO()) // ponytail: TODO ctx repo-wide
		if err != nil {
			return err
		}
		for _, item := range values(page) {
			if item == nil {
				continue
			}
			resourceID := id(item)
			if resourceID == "" {
				continue
			}
			az.Resources = append(az.Resources, terraformutils.NewSimpleResource(
				resourceID,
				name(item),
				tfType,
				az.ProviderName,
				defaultAllowEmptyValues))
		}
	}
	return nil
}
