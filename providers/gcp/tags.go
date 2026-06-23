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
	"context"
	"log"
	"strings"

	"google.golang.org/api/cloudresourcemanager/v3"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var tagsAllowEmptyValues = []string{""}

var tagsAdditionalFields = map[string]interface{}{}

type TagsGenerator struct {
	GCPService
}

// Generate TerraformResources from GCP API,
func (g *TagsGenerator) InitResources() error {
	ctx := context.Background()
	crmService, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return err
	}
	project := g.GetArgs()["project"].(string)

	tagKeysList := crmService.TagKeys.List().Parent("projects/" + project)
	if err := tagKeysList.Pages(ctx, func(page *cloudresourcemanager.ListTagKeysResponse) error {
		for _, obj := range page.TagKeys {
			t := strings.Split(obj.Name, "/")
			id := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name,
				id,
				"google_tags_tag_key",
				g.ProviderName,
				map[string]string{
					"parent":     "projects/" + project,
					"short_name": obj.ShortName,
				},
				tagsAllowEmptyValues,
				tagsAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}
