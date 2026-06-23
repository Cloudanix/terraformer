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
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"google.golang.org/api/iterator"
	loggingv2 "google.golang.org/api/logging/v2"

	"cloud.google.com/go/logging/logadmin"
)

var loggingAllowEmptyValues = []string{}

var loggingAdditionalFields = map[string]interface{}{}

type LoggingGenerator struct {
	GCPService
}

func (g *LoggingGenerator) loadLoggingMetrics(ctx context.Context, client *logadmin.Client) error {
	metricIterator := client.Metrics(ctx)

	for {
		metric, err := metricIterator.Next()

		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		g.Resources = append(g.Resources, terraformutils.NewResource(
			metric.ID,
			metric.ID,
			"google_logging_metric",
			g.ProviderName,
			map[string]string{
				"name":    metric.ID,
				"project": g.GetArgs()["project"].(string),
			},
			loggingAllowEmptyValues,
			loggingAdditionalFields,
		))
	}
	return nil
}

func (g *LoggingGenerator) loadLoggingSinks(ctx context.Context, client *logadmin.Client) error {
	sinkIterator := client.Sinks(ctx)
	for {
		sink, err := sinkIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		g.Resources = append(g.Resources, terraformutils.NewResource(
			sink.ID,
			sink.ID,
			"google_logging_project_sink",
			g.ProviderName,
			map[string]string{
				"name":    sink.ID,
				"project": g.GetArgs()["project"].(string),
			},
			loggingAllowEmptyValues,
			loggingAdditionalFields,
		))
	}
	return nil
}

// Generate TerraformResources from GCP API
func (g *LoggingGenerator) InitResources() error {
	project := g.GetArgs()["project"].(string)
	ctx := context.Background()
	client, err := logadmin.NewClient(ctx, project)
	if err != nil {
		return err
	}

	if err := g.loadLoggingMetrics(ctx, client); err != nil {
		return err
	}
	if err := g.loadLoggingSinks(ctx, client); err != nil {
		return err
	}
	g.loadScopedSinks(ctx)
	g.loadProjectLogging(ctx, project)

	return nil
}

// loadProjectLogging enumerates project-scoped logging resources via the loggingv2 REST API.
func (g *LoggingGenerator) loadProjectLogging(ctx context.Context, project string) {
	svc, err := loggingv2.NewService(ctx)
	if err != nil {
		return
	}
	projParent := "projects/" + project
	locParent := projParent + "/locations/-"

	if err := svc.Projects.Exclusions.List(projParent).Pages(ctx, func(p *loggingv2.ListExclusionsResponse) error {
		for _, o := range p.Exclusions {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				projParent+"/exclusions/"+o.Name, o.Name, "google_logging_project_exclusion", g.ProviderName,
				map[string]string{"name": o.Name, "project": project}, loggingAllowEmptyValues, loggingAdditionalFields))
		}
		return nil
	}); err != nil {
		_ = err
	}
	if err := svc.Projects.Locations.SavedQueries.List(locParent).Pages(ctx, func(p *loggingv2.ListSavedQueriesResponse) error {
		for _, o := range p.SavedQueries {
			t := strings.Split(o.Name, "/")
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, t[len(t)-1], "google_logging_saved_query", g.ProviderName,
				map[string]string{"name": t[len(t)-1], "project": project}, loggingAllowEmptyValues, loggingAdditionalFields))
		}
		return nil
	}); err != nil {
		_ = err
	}
	if err := svc.Projects.Locations.LogScopes.List(locParent).Pages(ctx, func(p *loggingv2.ListLogScopesResponse) error {
		for _, o := range p.LogScopes {
			t := strings.Split(o.Name, "/")
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, t[len(t)-1], "google_logging_log_scope", g.ProviderName,
				map[string]string{"name": t[len(t)-1], "project": project}, loggingAllowEmptyValues, loggingAdditionalFields))
		}
		return nil
	}); err != nil {
		_ = err
	}
	if err := svc.Projects.Locations.Buckets.List(locParent).Pages(ctx, func(p *loggingv2.ListBucketsResponse) error {
		for _, b := range p.Buckets {
			bt := strings.Split(b.Name, "/")
			bucketID := bt[len(bt)-1]
			loc := bt[len(bt)-3]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				b.Name, bucketID, "google_logging_project_bucket_config", g.ProviderName,
				map[string]string{"bucket_id": bucketID, "location": loc, "project": project},
				loggingAllowEmptyValues, loggingAdditionalFields))
			if verr := svc.Projects.Locations.Buckets.Views.List(b.Name).Pages(ctx, func(vp *loggingv2.ListViewsResponse) error {
				for _, v := range vp.Views {
					vt := strings.Split(v.Name, "/")
					g.Resources = append(g.Resources, terraformutils.NewResource(
						v.Name, vt[len(vt)-1], "google_logging_log_view", g.ProviderName,
						map[string]string{"name": vt[len(vt)-1], "bucket": b.Name, "location": loc, "project": project},
						loggingAllowEmptyValues, loggingAdditionalFields))
				}
				return nil
			}); verr != nil {
				_ = verr
			}
			if lerr := svc.Projects.Locations.Buckets.Links.List(b.Name).Pages(ctx, func(lp *loggingv2.ListLinksResponse) error {
				for _, l := range lp.Links {
					lt := strings.Split(l.Name, "/")
					g.Resources = append(g.Resources, terraformutils.NewResource(
						l.Name, lt[len(lt)-1], "google_logging_linked_dataset", g.ProviderName,
						map[string]string{"link_id": lt[len(lt)-1], "bucket": b.Name, "location": loc, "project": project},
						loggingAllowEmptyValues, loggingAdditionalFields))
				}
				return nil
			}); lerr != nil {
				_ = lerr
			}
		}
		return nil
	}); err != nil {
		_ = err
	}
}

// loadScopedSinks enumerates folder/organization log sinks when GOOGLE_FOLDER /
// GOOGLE_ORGANIZATION are set (org/folder-scoped; no-op otherwise).
func (g *LoggingGenerator) loadScopedSinks(ctx context.Context) {
	svc, err := loggingv2.NewService(ctx)
	if err != nil {
		return
	}
	if folder, _ := g.GetArgs()["folder"].(string); folder != "" {
		if err := svc.Folders.Sinks.List("folders/"+folder).Pages(ctx, func(page *loggingv2.ListSinksResponse) error {
			for _, s := range page.Sinks {
				g.Resources = append(g.Resources, terraformutils.NewResource(
					"folders/"+folder+"/sinks/"+s.Name, s.Name, "google_logging_folder_sink", g.ProviderName,
					map[string]string{"name": s.Name, "folder": folder}, loggingAllowEmptyValues, loggingAdditionalFields,
				))
			}
			return nil
		}); err != nil {
			_ = err
		}
		if err := svc.Folders.Exclusions.List("folders/"+folder).Pages(ctx, func(page *loggingv2.ListExclusionsResponse) error {
			for _, o := range page.Exclusions {
				g.Resources = append(g.Resources, terraformutils.NewResource(
					"folders/"+folder+"/exclusions/"+o.Name, o.Name, "google_logging_folder_exclusion", g.ProviderName,
					map[string]string{"name": o.Name, "folder": folder}, loggingAllowEmptyValues, loggingAdditionalFields))
			}
			return nil
		}); err != nil {
			_ = err
		}
		if err := svc.Folders.Locations.Buckets.List("folders/"+folder+"/locations/-").Pages(ctx, func(page *loggingv2.ListBucketsResponse) error {
			for _, b := range page.Buckets {
				bt := strings.Split(b.Name, "/")
				g.Resources = append(g.Resources, terraformutils.NewResource(
					b.Name, bt[len(bt)-1], "google_logging_folder_bucket_config", g.ProviderName,
					map[string]string{"bucket_id": bt[len(bt)-1], "location": bt[len(bt)-3], "folder": "folders/" + folder},
					loggingAllowEmptyValues, loggingAdditionalFields))
			}
			return nil
		}); err != nil {
			_ = err
		}
		g.Resources = append(g.Resources, terraformutils.NewResource(
			"folders/"+folder+"/settings", folder, "google_logging_folder_settings", g.ProviderName,
			map[string]string{"folder": folder}, loggingAllowEmptyValues, loggingAdditionalFields))
	}
	if org, _ := g.GetArgs()["organization"].(string); org != "" {
		if err := svc.Organizations.Sinks.List("organizations/"+org).Pages(ctx, func(page *loggingv2.ListSinksResponse) error {
			for _, s := range page.Sinks {
				g.Resources = append(g.Resources, terraformutils.NewResource(
					"organizations/"+org+"/sinks/"+s.Name, s.Name, "google_logging_organization_sink", g.ProviderName,
					map[string]string{"name": s.Name, "org_id": org}, loggingAllowEmptyValues, loggingAdditionalFields,
				))
			}
			return nil
		}); err != nil {
			_ = err
		}
		if err := svc.Organizations.Exclusions.List("organizations/"+org).Pages(ctx, func(page *loggingv2.ListExclusionsResponse) error {
			for _, o := range page.Exclusions {
				g.Resources = append(g.Resources, terraformutils.NewResource(
					"organizations/"+org+"/exclusions/"+o.Name, o.Name, "google_logging_organization_exclusion", g.ProviderName,
					map[string]string{"name": o.Name, "org_id": org}, loggingAllowEmptyValues, loggingAdditionalFields))
			}
			return nil
		}); err != nil {
			_ = err
		}
		if err := svc.Organizations.Locations.Buckets.List("organizations/"+org+"/locations/-").Pages(ctx, func(page *loggingv2.ListBucketsResponse) error {
			for _, b := range page.Buckets {
				bt := strings.Split(b.Name, "/")
				g.Resources = append(g.Resources, terraformutils.NewResource(
					b.Name, bt[len(bt)-1], "google_logging_organization_bucket_config", g.ProviderName,
					map[string]string{"bucket_id": bt[len(bt)-1], "location": bt[len(bt)-3], "organization": org},
					loggingAllowEmptyValues, loggingAdditionalFields))
			}
			return nil
		}); err != nil {
			_ = err
		}
		g.Resources = append(g.Resources, terraformutils.NewResource(
			"organizations/"+org+"/settings", org, "google_logging_organization_settings", g.ProviderName,
			map[string]string{"organization": org}, loggingAllowEmptyValues, loggingAdditionalFields))
	}
}
