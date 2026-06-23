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
	"google.golang.org/api/contactcenterinsights/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var contactCenterInsightsAllowEmptyValues = []string{""}

var contactCenterInsightsAdditionalFields = map[string]interface{}{}

type ContactCenterInsightsGenerator struct {
	GCPService
}

// Generate TerraformResources from GCP API,
func (g *ContactCenterInsightsGenerator) InitResources() error {
	ctx := context.Background()
	cciService, err := contactcenterinsights.NewService(ctx)
	if err != nil {
		return err
	}
	location := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + g.GetArgs()["project"].(string) + "/locations/" + location

	project := g.GetArgs()["project"].(string)
	viewsList := cciService.Projects.Locations.Views.List(parent)
	if err := viewsList.Pages(ctx, func(page *contactcenterinsights.GoogleCloudContactcenterinsightsV1ListViewsResponse) error {
		for _, obj := range page.Views {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name, name, "google_contact_center_insights_view", g.ProviderName,
				map[string]string{"name": name, "project": project, "location": location},
				contactCenterInsightsAllowEmptyValues, contactCenterInsightsAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := cciService.Projects.Locations.AnalysisRules.List(parent).Pages(ctx, func(p *contactcenterinsights.GoogleCloudContactcenterinsightsV1ListAnalysisRulesResponse) error {
		for _, o := range p.AnalysisRules {
			t := strings.Split(o.Name, "/")
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, t[len(t)-1], "google_contact_center_insights_analysis_rule", g.ProviderName,
				map[string]string{"name": t[len(t)-1], "location": location, "project": project},
				contactCenterInsightsAllowEmptyValues, contactCenterInsightsAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := cciService.Projects.Locations.AssessmentRules.List(parent).Pages(ctx, func(p *contactcenterinsights.GoogleCloudContactcenterinsightsV1ListAssessmentRulesResponse) error {
		for _, o := range p.AssessmentRules {
			t := strings.Split(o.Name, "/")
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, t[len(t)-1], "google_contact_center_insights_assessment_rule", g.ProviderName,
				map[string]string{"assessment_rule_id": t[len(t)-1], "location": location, "project": project},
				contactCenterInsightsAllowEmptyValues, contactCenterInsightsAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := cciService.Projects.Locations.AutoLabelingRules.List(parent).Pages(ctx, func(p *contactcenterinsights.GoogleCloudContactcenterinsightsV1ListAutoLabelingRulesResponse) error {
		for _, o := range p.AutoLabelingRules {
			t := strings.Split(o.Name, "/")
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, t[len(t)-1], "google_contact_center_insights_auto_labeling_rule", g.ProviderName,
				map[string]string{"name": t[len(t)-1], "location": location, "project": project},
				contactCenterInsightsAllowEmptyValues, contactCenterInsightsAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := cciService.Projects.Locations.QaScorecards.List(parent).Pages(ctx, func(p *contactcenterinsights.GoogleCloudContactcenterinsightsV1ListQaScorecardsResponse) error {
		for _, o := range p.QaScorecards {
			t := strings.Split(o.Name, "/")
			scorecardID := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, scorecardID, "google_contact_center_insights_qa_scorecard", g.ProviderName,
				map[string]string{"qa_scorecard_id": scorecardID, "location": location, "project": project},
				contactCenterInsightsAllowEmptyValues, contactCenterInsightsAdditionalFields))
			if rerr := cciService.Projects.Locations.QaScorecards.Revisions.List(o.Name).Pages(ctx, func(rp *contactcenterinsights.GoogleCloudContactcenterinsightsV1ListQaScorecardRevisionsResponse) error {
				for _, rev := range rp.QaScorecardRevisions {
					rt := strings.Split(rev.Name, "/")
					g.Resources = append(g.Resources, terraformutils.NewResource(
						rev.Name, scorecardID+"_"+rt[len(rt)-1], "google_contact_center_insights_qa_scorecard_revision", g.ProviderName,
						map[string]string{"qa_scorecard_revision_id": rt[len(rt)-1], "qa_scorecard": scorecardID, "location": location, "project": project},
						contactCenterInsightsAllowEmptyValues, contactCenterInsightsAdditionalFields))
					if qerr := cciService.Projects.Locations.QaScorecards.Revisions.QaQuestions.List(rev.Name).Pages(ctx, func(qp *contactcenterinsights.GoogleCloudContactcenterinsightsV1ListQaQuestionsResponse) error {
						for _, q := range qp.QaQuestions {
							qt := strings.Split(q.Name, "/")
							g.Resources = append(g.Resources, terraformutils.NewResource(
								q.Name, scorecardID+"_"+qt[len(qt)-1], "google_contact_center_insights_qa_question", g.ProviderName,
								map[string]string{"qa_question_id": qt[len(qt)-1], "location": location, "project": project},
								contactCenterInsightsAllowEmptyValues, contactCenterInsightsAdditionalFields))
						}
						return nil
					}); qerr != nil {
						log.Println(qerr)
					}
				}
				return nil
			}); rerr != nil {
				log.Println(rerr)
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}
