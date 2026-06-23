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

	return nil
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
	}
}
