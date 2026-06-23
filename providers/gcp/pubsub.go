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

	"google.golang.org/api/pubsub/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var pubsubAllowEmptyValues = []string{""}

var pubsubAdditionalFields = map[string]interface{}{}

type PubsubGenerator struct {
	GCPService
}

// Run on subscriptionsList and create for each TerraformResource
func (g PubsubGenerator) createSubscriptionsResources(ctx context.Context, subscriptionsList *pubsub.ProjectsSubscriptionsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	if err := subscriptionsList.Pages(ctx, func(page *pubsub.ListSubscriptionsResponse) error {
		for _, obj := range page.Subscriptions {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				name,
				obj.Name,
				"google_pubsub_subscription",
				g.ProviderName,
				map[string]string{
					"name":    name,
					"project": g.GetArgs()["project"].(string),
				},
				pubsubAllowEmptyValues,
				pubsubAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Run on topicsList and create for each TerraformResource
func (g PubsubGenerator) createTopicsListResources(ctx context.Context, topicsList *pubsub.ProjectsTopicsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	if err := topicsList.Pages(ctx, func(page *pubsub.ListTopicsResponse) error {
		for _, obj := range page.Topics {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				g.GetArgs()["project"].(string)+"/"+name,
				obj.Name,
				"google_pubsub_topic",
				g.ProviderName,
				map[string]string{
					"name":    name,
					"project": g.GetArgs()["project"].(string),
				},
				pubsubAllowEmptyValues,
				pubsubAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Run on schemasList and create for each TerraformResource
func (g PubsubGenerator) createSchemasResources(ctx context.Context, schemasList *pubsub.ProjectsSchemasListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	if err := schemasList.Pages(ctx, func(page *pubsub.ListSchemasResponse) error {
		for _, obj := range page.Schemas {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_pubsub_schema",
				g.ProviderName,
				map[string]string{
					"name":    name,
					"project": g.GetArgs()["project"].(string),
				},
				pubsubAllowEmptyValues,
				pubsubAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *PubsubGenerator) InitResources() error {
	ctx := context.Background()
	pubsubService, err := pubsub.NewService(ctx)
	if err != nil {
		return err
	}

	subscriptionsList := pubsubService.Projects.Subscriptions.List("projects/" + g.GetArgs()["project"].(string))
	subscriptionsResources := g.createSubscriptionsResources(ctx, subscriptionsList)

	topicsList := pubsubService.Projects.Topics.List("projects/" + g.GetArgs()["project"].(string))
	topicsResources := g.createTopicsListResources(ctx, topicsList)

	schemasList := pubsubService.Projects.Schemas.List("projects/" + g.GetArgs()["project"].(string))
	schemasResources := g.createSchemasResources(ctx, schemasList)

	g.Resources = append(g.Resources, subscriptionsResources...)
	g.Resources = append(g.Resources, topicsResources...)
	g.Resources = append(g.Resources, schemasResources...)

	project := g.GetArgs()["project"].(string)
	for _, r := range topicsResources {
		name := r.Item["name"].(string)
		full := "projects/" + project + "/topics/" + name
		if policy, perr := pubsubService.Projects.Topics.GetIamPolicy(full).Do(); perr == nil {
			for _, b := range policy.Bindings {
				for _, m := range b.Members {
					g.Resources = append(g.Resources, terraformutils.NewResource(
						full+" "+b.Role+" "+m, name+"_"+b.Role+"_"+m,
						"google_pubsub_topic_iam_member", g.ProviderName,
						map[string]string{"topic": name, "role": b.Role, "member": m, "project": project},
						pubsubAllowEmptyValues, pubsubAdditionalFields))
				}
			}
		}
	}
	for _, r := range subscriptionsResources {
		name := r.Item["name"].(string)
		full := "projects/" + project + "/subscriptions/" + name
		if policy, perr := pubsubService.Projects.Subscriptions.GetIamPolicy(full).Do(); perr == nil {
			for _, b := range policy.Bindings {
				for _, m := range b.Members {
					g.Resources = append(g.Resources, terraformutils.NewResource(
						full+" "+b.Role+" "+m, name+"_"+b.Role+"_"+m,
						"google_pubsub_subscription_iam_member", g.ProviderName,
						map[string]string{"subscription": name, "role": b.Role, "member": m, "project": project},
						pubsubAllowEmptyValues, pubsubAdditionalFields))
				}
			}
		}
	}
	return nil
}

func (g *PubsubGenerator) PostConvertHook() error {
	for i, r := range g.Resources {
		for _, topic := range g.Resources {
			if r.InstanceState.Attributes["topic"] == "projects/"+g.GetArgs()["project"].(string)+"/topics/"+topic.InstanceState.Attributes["name"] {
				g.Resources[i].Item["topic"] = "${google_pubsub_topic." + topic.ResourceName + ".name}"
			}
		}
	}
	return nil
}
