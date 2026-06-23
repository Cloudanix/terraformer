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

	"google.golang.org/api/bigqueryreservation/v1"
	"google.golang.org/api/compute/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var bigQueryReservationAllowEmptyValues = []string{""}

var bigQueryReservationAdditionalFields = map[string]interface{}{}

type BigQueryReservationGenerator struct {
	GCPService
}

// Generate TerraformResources from GCP API,
func (g *BigQueryReservationGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := bigqueryreservation.NewService(ctx)
	if err != nil {
		return err
	}
	project := g.GetArgs()["project"].(string)
	location := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + location
	tail := func(s string) string { p := strings.Split(s, "/"); return p[len(p)-1] }

	list := svc.Projects.Locations.Reservations.List(parent)
	if err := list.Pages(ctx, func(page *bigqueryreservation.ListReservationsResponse) error {
		for _, obj := range page.Reservations {
			name := tail(obj.Name)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name, name, "google_bigquery_reservation", g.ProviderName,
				map[string]string{"name": name, "project": project, "location": location},
				bigQueryReservationAllowEmptyValues, bigQueryReservationAdditionalFields,
			))
			if ae := svc.Projects.Locations.Reservations.Assignments.List(obj.Name).Pages(ctx, func(ap *bigqueryreservation.ListAssignmentsResponse) error {
				for _, a := range ap.Assignments {
					g.Resources = append(g.Resources, terraformutils.NewResource(
						a.Name, name+"_"+tail(a.Name), "google_bigquery_reservation_assignment", g.ProviderName,
						map[string]string{"reservation": name, "location": location, "project": project},
						bigQueryReservationAllowEmptyValues, bigQueryReservationAdditionalFields))
				}
				return nil
			}); ae != nil {
				log.Println(ae)
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := svc.Projects.Locations.CapacityCommitments.List(parent).Pages(ctx, func(p *bigqueryreservation.ListCapacityCommitmentsResponse) error {
		for _, o := range p.CapacityCommitments {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, tail(o.Name), "google_bigquery_capacity_commitment", g.ProviderName,
				map[string]string{"capacity_commitment_id": tail(o.Name), "location": location, "project": project},
				bigQueryReservationAllowEmptyValues, bigQueryReservationAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := svc.Projects.Locations.ReservationGroups.List(parent).Pages(ctx, func(p *bigqueryreservation.ListReservationGroupsResponse) error {
		for _, o := range p.ReservationGroups {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, tail(o.Name), "google_bigquery_reservation_group", g.ProviderName,
				map[string]string{"reservation_group_id": tail(o.Name), "location": location, "project": project},
				bigQueryReservationAllowEmptyValues, bigQueryReservationAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	// BI reservation is a per-location singleton.
	if bi, e := svc.Projects.Locations.GetBiReservation(parent + "/biReservation").Do(); e == nil && bi != nil {
		g.Resources = append(g.Resources, terraformutils.NewResource(
			bi.Name, project+"_"+location+"_bi_reservation", "google_bigquery_bi_reservation", g.ProviderName,
			map[string]string{"location": location, "project": project},
			bigQueryReservationAllowEmptyValues, bigQueryReservationAdditionalFields))
	}
	return nil
}
