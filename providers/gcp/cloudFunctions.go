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

	"google.golang.org/api/cloudfunctions/v2"
	"google.golang.org/api/compute/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var cloudFunctionsAllowEmptyValues = []string{""}

var cloudFunctionsAdditionalFields = map[string]interface{}{}

type CloudFunctionsGenerator struct {
	GCPService
}

// Run on CloudFunctionsList and create for each TerraformResource
func (g CloudFunctionsGenerator) createCloudFunctionsResources(ctx context.Context, functionsList *cloudfunctions.ProjectsLocationsFunctionsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	if err := functionsList.Pages(ctx, func(page *cloudfunctions.ListFunctionsResponse) error {
		for _, functions := range page.Functions {
			t := strings.Split(functions.Name, "/")
			if functions.Environment == "GEN_1" {
				name := t[len(t)-1]
				resources = append(resources, terraformutils.NewResource(
					g.GetArgs()["project"].(string)+"/"+g.GetArgs()["region"].(compute.Region).Name+"/"+name,
					g.GetArgs()["region"].(compute.Region).Name+"_"+name,
					"google_cloudfunctions_function",
					g.ProviderName,
					map[string]string{
						"name":     name,
						"project":  g.GetArgs()["project"].(string),
						"location": g.GetArgs()["region"].(compute.Region).Name,
					},
					cloudFunctionsAllowEmptyValues,
					cloudFunctionsAdditionalFields,
				))
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

func (g CloudFunctionsGenerator) createCloudFunctions2ndGenResources(ctx context.Context, functionsList *cloudfunctions.ProjectsLocationsFunctionsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	if err := functionsList.Pages(ctx, func(page *cloudfunctions.ListFunctionsResponse) error {
		for _, functions := range page.Functions {
			t := strings.Split(functions.Name, "/")
			if functions.Environment == "GEN_2" {
				name := t[len(t)-1]
				resources = append(resources, terraformutils.NewResource(
					g.GetArgs()["project"].(string)+"/"+g.GetArgs()["region"].(compute.Region).Name+"/"+name,
					g.GetArgs()["region"].(compute.Region).Name+"_"+name,
					"google_cloudfunctions2_function",
					g.ProviderName,
					map[string]string{
						"name":     name,
						"project":  g.GetArgs()["project"].(string),
						"location": g.GetArgs()["region"].(compute.Region).Name,
					},
					cloudFunctionsAllowEmptyValues,
					cloudFunctionsAdditionalFields,
				))
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
// from each CloudFunctions create 1 TerraformResource
// Need CloudFunctions name as ID for terraform resource
func (g *CloudFunctionsGenerator) InitResources() error {
	ctx := context.Background()
	cloudfunctionsService, err := cloudfunctions.NewService(ctx)
	if err != nil {
		return err
	}

	functionsList := cloudfunctionsService.Projects.Locations.Functions.List("projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name)

	g.Resources = append(g.Resources, g.createCloudFunctionsResources(ctx, functionsList)...)
	g.Resources = append(g.Resources, g.createCloudFunctions2ndGenResources(ctx, functionsList)...)

	// Per-function IAM (member form).
	parent := "projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name
	if err := cloudfunctionsService.Projects.Locations.Functions.List(parent).Pages(ctx, func(page *cloudfunctions.ListFunctionsResponse) error {
		for _, fn := range page.Functions {
			policy, perr := cloudfunctionsService.Projects.Locations.Functions.GetIamPolicy(fn.Name).Do()
			if perr != nil {
				continue
			}
			short := strings.Split(fn.Name, "/")[len(strings.Split(fn.Name, "/"))-1]
			for _, b := range policy.Bindings {
				for _, m := range b.Members {
					g.Resources = append(g.Resources, terraformutils.NewResource(
						fn.Name+" "+b.Role+" "+m, short+"_"+b.Role+"_"+m,
						"google_cloudfunctions2_function_iam_member", g.ProviderName,
						map[string]string{"cloud_function": short, "role": b.Role, "member": m, "project": g.GetArgs()["project"].(string), "location": g.GetArgs()["region"].(compute.Region).Name},
						[]string{""}, map[string]interface{}{}))
				}
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	return nil
}
