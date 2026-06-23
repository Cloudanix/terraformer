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

	"google.golang.org/api/aiplatform/v1"
	"google.golang.org/api/compute/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var vertexAIAllowEmptyValues = []string{""}

var vertexAIAdditionalFields = map[string]interface{}{}

type VertexAIGenerator struct {
	GCPService
}

// Run on endpointsList and create for each TerraformResource
func (g VertexAIGenerator) createResources(ctx context.Context, endpointsList *aiplatform.ProjectsLocationsEndpointsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := endpointsList.Pages(ctx, func(page *aiplatform.GoogleCloudAiplatformV1ListEndpointsResponse) error {
		for _, obj := range page.Endpoints {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_vertex_ai_endpoint",
				g.ProviderName,
				map[string]string{
					"name":     name,
					"project":  g.GetArgs()["project"].(string),
					"location": location,
				},
				vertexAIAllowEmptyValues,
				vertexAIAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Run on datasetsList and create for each TerraformResource
func (g VertexAIGenerator) createDatasetsResources(ctx context.Context, datasetsList *aiplatform.ProjectsLocationsDatasetsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := datasetsList.Pages(ctx, func(page *aiplatform.GoogleCloudAiplatformV1ListDatasetsResponse) error {
		for _, obj := range page.Datasets {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_vertex_ai_dataset",
				g.ProviderName,
				map[string]string{
					"name":     name,
					"project":  g.GetArgs()["project"].(string),
					"location": location,
				},
				vertexAIAllowEmptyValues,
				vertexAIAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *VertexAIGenerator) InitResources() error {
	ctx := context.Background()
	vertexAIService, err := aiplatform.NewService(ctx)
	if err != nil {
		return err
	}

	parent := "projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name
	endpointsList := vertexAIService.Projects.Locations.Endpoints.List(parent)
	g.Resources = append(g.Resources, g.createResources(ctx, endpointsList)...)

	datasetsList := vertexAIService.Projects.Locations.Datasets.List(parent)
	g.Resources = append(g.Resources, g.createDatasetsResources(ctx, datasetsList)...)

	fosList := vertexAIService.Projects.Locations.FeatureOnlineStores.List(parent)
	g.Resources = append(g.Resources, g.createFeatureOnlineStoresResources(ctx, fosList)...)

	loc := g.GetArgs()["region"].(compute.Region).Name
	proj := g.GetArgs()["project"].(string)
	mk := func(name, tfType string) terraformutils.Resource {
		return terraformutils.NewResource(
			"projects/"+proj+"/locations/"+loc+"/"+name, name, tfType, g.ProviderName,
			map[string]string{"name": name, "project": proj, "region": loc},
			vertexAIAllowEmptyValues, vertexAIAdditionalFields,
		)
	}
	if err := vertexAIService.Projects.Locations.Featurestores.List(parent).Pages(ctx, func(p *aiplatform.GoogleCloudAiplatformV1ListFeaturestoresResponse) error {
		for _, o := range p.Featurestores {
			t := strings.Split(o.Name, "/")
			fsName := t[len(t)-1]
			g.Resources = append(g.Resources, mk(fsName, "google_vertex_ai_featurestore"))
			if eerr := vertexAIService.Projects.Locations.Featurestores.EntityTypes.List(o.Name).Pages(ctx, func(ep *aiplatform.GoogleCloudAiplatformV1ListEntityTypesResponse) error {
				for _, et := range ep.EntityTypes {
					ett := strings.Split(et.Name, "/")
					etName := ett[len(ett)-1]
					g.Resources = append(g.Resources, terraformutils.NewResource(
						et.Name, fsName+"_"+etName, "google_vertex_ai_featurestore_entitytype", g.ProviderName,
						map[string]string{"name": etName, "featurestore": o.Name, "region": loc, "project": proj},
						vertexAIAllowEmptyValues, vertexAIAdditionalFields))
					if ferr := vertexAIService.Projects.Locations.Featurestores.EntityTypes.Features.List(et.Name).Pages(ctx, func(fp *aiplatform.GoogleCloudAiplatformV1ListFeaturesResponse) error {
						for _, f := range fp.Features {
							ft := strings.Split(f.Name, "/")
							g.Resources = append(g.Resources, terraformutils.NewResource(
								f.Name, etName+"_"+ft[len(ft)-1], "google_vertex_ai_featurestore_entitytype_feature", g.ProviderName,
								map[string]string{"name": ft[len(ft)-1], "entitytype": et.Name, "region": loc, "project": proj},
								vertexAIAllowEmptyValues, vertexAIAdditionalFields))
						}
						return nil
					}); ferr != nil {
						log.Println(ferr)
					}
				}
				return nil
			}); eerr != nil {
				log.Println(eerr)
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := vertexAIService.Projects.Locations.Tensorboards.List(parent).Pages(ctx, func(p *aiplatform.GoogleCloudAiplatformV1ListTensorboardsResponse) error {
		for _, o := range p.Tensorboards {
			t := strings.Split(o.Name, "/")
			g.Resources = append(g.Resources, mk(t[len(t)-1], "google_vertex_ai_tensorboard"))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := vertexAIService.Projects.Locations.FeatureGroups.List(parent).Pages(ctx, func(p *aiplatform.GoogleCloudAiplatformV1ListFeatureGroupsResponse) error {
		for _, o := range p.FeatureGroups {
			t := strings.Split(o.Name, "/")
			g.Resources = append(g.Resources, mk(t[len(t)-1], "google_vertex_ai_feature_group"))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := vertexAIService.Projects.Locations.Indexes.List(parent).Pages(ctx, func(p *aiplatform.GoogleCloudAiplatformV1ListIndexesResponse) error {
		for _, o := range p.Indexes {
			t := strings.Split(o.Name, "/")
			g.Resources = append(g.Resources, mk(t[len(t)-1], "google_vertex_ai_index"))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := vertexAIService.Projects.Locations.IndexEndpoints.List(parent).Pages(ctx, func(p *aiplatform.GoogleCloudAiplatformV1ListIndexEndpointsResponse) error {
		for _, o := range p.IndexEndpoints {
			t := strings.Split(o.Name, "/")
			ieName := t[len(t)-1]
			g.Resources = append(g.Resources, mk(ieName, "google_vertex_ai_index_endpoint"))
			for _, di := range o.DeployedIndexes {
				g.Resources = append(g.Resources, terraformutils.NewResource(
					o.Name+"/deployedIndex/"+di.Id, ieName+"_"+di.Id,
					"google_vertex_ai_index_endpoint_deployed_index", g.ProviderName,
					map[string]string{"index_endpoint": ieName, "deployed_index_id": di.Id, "project": proj, "region": loc},
					vertexAIAllowEmptyValues, vertexAIAdditionalFields))
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := vertexAIService.Projects.Locations.ReasoningEngines.List(parent).Pages(ctx, func(p *aiplatform.GoogleCloudAiplatformV1ListReasoningEnginesResponse) error {
		for _, o := range p.ReasoningEngines {
			t := strings.Split(o.Name, "/")
			g.Resources = append(g.Resources, mk(t[len(t)-1], "google_vertex_ai_reasoning_engine"))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := vertexAIService.Projects.Locations.DeploymentResourcePools.List(parent).Pages(ctx, func(p *aiplatform.GoogleCloudAiplatformV1ListDeploymentResourcePoolsResponse) error {
		for _, o := range p.DeploymentResourcePools {
			t := strings.Split(o.Name, "/")
			g.Resources = append(g.Resources, mk(t[len(t)-1], "google_vertex_ai_deployment_resource_pool"))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}

// Run on featureOnlineStoresList and create for each TerraformResource
func (g VertexAIGenerator) createFeatureOnlineStoresResources(ctx context.Context, list *aiplatform.ProjectsLocationsFeatureOnlineStoresListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := list.Pages(ctx, func(page *aiplatform.GoogleCloudAiplatformV1ListFeatureOnlineStoresResponse) error {
		for _, obj := range page.FeatureOnlineStores {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_vertex_ai_feature_online_store",
				g.ProviderName,
				map[string]string{
					"name":    name,
					"project": g.GetArgs()["project"].(string),
					"region":  location,
				},
				vertexAIAllowEmptyValues,
				vertexAIAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}
