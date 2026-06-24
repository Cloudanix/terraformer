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
	"github.com/aws/aws-sdk-go-v2/service/location"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type LocationGenerator struct {
	AWSService
}

// InitResources enumerates Location Service maps, place indexes, trackers,
// geofence collections, and route calculators. Import IDs are the resource name.
func (g *LocationGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := location.NewFromConfig(config)
	ctx := awsContext()

	maps := location.NewListMapsPaginator(svc, &location.ListMapsInput{})
	for maps.HasMorePages() {
		page, err := maps.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, m := range page.Entries {
			name := StringValue(m.MapName)
			if name != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					name, name, "aws_location_map", "aws", defaultAllowEmptyValues))
			}
		}
	}

	indexes := location.NewListPlaceIndexesPaginator(svc, &location.ListPlaceIndexesInput{})
	for indexes.HasMorePages() {
		page, err := indexes.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, idx := range page.Entries {
			name := StringValue(idx.IndexName)
			if name != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					name, name, "aws_location_place_index", "aws", defaultAllowEmptyValues))
			}
		}
	}

	trackers := location.NewListTrackersPaginator(svc, &location.ListTrackersInput{})
	for trackers.HasMorePages() {
		page, err := trackers.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, t := range page.Entries {
			name := StringValue(t.TrackerName)
			if name != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					name, name, "aws_location_tracker", "aws", defaultAllowEmptyValues))
				trackerName := name
				for cp := location.NewListTrackerConsumersPaginator(svc, &location.ListTrackerConsumersInput{TrackerName: &trackerName}); cp.HasMorePages(); {
					cpage, err := cp.NextPage(ctx)
					if err != nil {
						break
					}
					for _, consumerArn := range cpage.ConsumerArns {
						if consumerArn == "" {
							continue
						}
						g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
							trackerName+"|"+consumerArn, trackerName+"_consumer", "aws_location_tracker_association", "aws", defaultAllowEmptyValues))
					}
				}
			}
		}
	}

	collections := location.NewListGeofenceCollectionsPaginator(svc, &location.ListGeofenceCollectionsInput{})
	for collections.HasMorePages() {
		page, err := collections.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, c := range page.Entries {
			name := StringValue(c.CollectionName)
			if name != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					name, name, "aws_location_geofence_collection", "aws", defaultAllowEmptyValues))
			}
		}
	}

	calculators := location.NewListRouteCalculatorsPaginator(svc, &location.ListRouteCalculatorsInput{})
	for calculators.HasMorePages() {
		page, err := calculators.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, rc := range page.Entries {
			name := StringValue(rc.CalculatorName)
			if name != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					name, name, "aws_location_route_calculator", "aws", defaultAllowEmptyValues))
			}
		}
	}
	return nil
}
