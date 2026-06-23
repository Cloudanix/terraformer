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

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/datastream/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var datastreamAllowEmptyValues = []string{""}

var datastreamAdditionalFields = map[string]interface{}{}

type DatastreamGenerator struct {
	GCPService
}

// Run on streamsList and create for each TerraformResource
func (g DatastreamGenerator) createResources(ctx context.Context, streamsList *datastream.ProjectsLocationsStreamsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := streamsList.Pages(ctx, func(page *datastream.ListStreamsResponse) error {
		for _, obj := range page.Streams {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_datastream_stream",
				g.ProviderName,
				map[string]string{
					"name":      name,
					"stream_id": name,
					"project":   g.GetArgs()["project"].(string),
					"location":  location,
				},
				datastreamAllowEmptyValues,
				datastreamAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *DatastreamGenerator) InitResources() error {
	ctx := context.Background()
	datastreamService, err := datastream.NewService(ctx)
	if err != nil {
		return err
	}

	streamsList := datastreamService.Projects.Locations.Streams.List(
		"projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name)
	g.Resources = g.createResources(ctx, streamsList)
	return nil
}
