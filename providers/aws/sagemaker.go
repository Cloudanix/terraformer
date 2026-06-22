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
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sagemaker"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type SageMakerGenerator struct {
	AWSService
}

// InitResources enumerates the common SageMaker resources: domains, notebook
// instances, models, endpoints, endpoint configs, and code repositories.
// Import IDs are the resource's name (domain id for domains).
func (g *SageMakerGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := sagemaker.NewFromConfig(config)
	ctx := context.TODO()

	domains := sagemaker.NewListDomainsPaginator(svc, &sagemaker.ListDomainsInput{})
	for domains.HasMorePages() {
		page, err := domains.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, d := range page.Domains {
			id := StringValue(d.DomainId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(d.DomainName), "aws_sagemaker_domain", "aws", defaultAllowEmptyValues))
		}
	}

	notebooks := sagemaker.NewListNotebookInstancesPaginator(svc, &sagemaker.ListNotebookInstancesInput{})
	for notebooks.HasMorePages() {
		page, err := notebooks.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, n := range page.NotebookInstances {
			name := StringValue(n.NotebookInstanceName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_sagemaker_notebook_instance", "aws", defaultAllowEmptyValues))
		}
	}

	models := sagemaker.NewListModelsPaginator(svc, &sagemaker.ListModelsInput{})
	for models.HasMorePages() {
		page, err := models.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, m := range page.Models {
			name := StringValue(m.ModelName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_sagemaker_model", "aws", defaultAllowEmptyValues))
		}
	}

	endpoints := sagemaker.NewListEndpointsPaginator(svc, &sagemaker.ListEndpointsInput{})
	for endpoints.HasMorePages() {
		page, err := endpoints.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, ep := range page.Endpoints {
			name := StringValue(ep.EndpointName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_sagemaker_endpoint", "aws", defaultAllowEmptyValues))
		}
	}

	endpointConfigs := sagemaker.NewListEndpointConfigsPaginator(svc, &sagemaker.ListEndpointConfigsInput{})
	for endpointConfigs.HasMorePages() {
		page, err := endpointConfigs.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, ec := range page.EndpointConfigs {
			name := StringValue(ec.EndpointConfigName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_sagemaker_endpoint_configuration", "aws", defaultAllowEmptyValues))
		}
	}

	repos := sagemaker.NewListCodeRepositoriesPaginator(svc, &sagemaker.ListCodeRepositoriesInput{})
	for repos.HasMorePages() {
		page, err := repos.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, repo := range page.CodeRepositorySummaryList {
			name := StringValue(repo.CodeRepositoryName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_sagemaker_code_repository", "aws", defaultAllowEmptyValues))
		}
	}

	g.addMoreSageMaker(ctx, svc)
	return nil
}

// addMoreSageMaker enumerates the additional top-level SageMaker resources that
// have a simple List* paginator returning a name. Import ID is the name. Errors
// on any single list are logged and skipped so one missing permission doesn't
// abort the whole SageMaker import.
func (g *SageMakerGenerator) addMoreSageMaker(ctx context.Context, svc *sagemaker.Client) {
	add := func(name, tfType string) {
		if name != "" {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, tfType, "aws", defaultAllowEmptyValues))
		}
	}
	if p := sagemaker.NewListAppImageConfigsPaginator(svc, &sagemaker.ListAppImageConfigsInput{}); true {
		for p.HasMorePages() {
			pg, e := p.NextPage(ctx)
			if e != nil {
				break
			}
			for _, x := range pg.AppImageConfigs {
				add(StringValue(x.AppImageConfigName), "aws_sagemaker_app_image_config")
			}
		}
	}
	for p := sagemaker.NewListDeviceFleetsPaginator(svc, &sagemaker.ListDeviceFleetsInput{}); p.HasMorePages(); {
		pg, e := p.NextPage(ctx)
		if e != nil {
			break
		}
		for _, x := range pg.DeviceFleetSummaries {
			add(StringValue(x.DeviceFleetName), "aws_sagemaker_device_fleet")
		}
	}
	for p := sagemaker.NewListFeatureGroupsPaginator(svc, &sagemaker.ListFeatureGroupsInput{}); p.HasMorePages(); {
		pg, e := p.NextPage(ctx)
		if e != nil {
			break
		}
		for _, x := range pg.FeatureGroupSummaries {
			add(StringValue(x.FeatureGroupName), "aws_sagemaker_feature_group")
		}
	}
	for p := sagemaker.NewListFlowDefinitionsPaginator(svc, &sagemaker.ListFlowDefinitionsInput{}); p.HasMorePages(); {
		pg, e := p.NextPage(ctx)
		if e != nil {
			break
		}
		for _, x := range pg.FlowDefinitionSummaries {
			add(StringValue(x.FlowDefinitionName), "aws_sagemaker_flow_definition")
		}
	}
	for p := sagemaker.NewListHumanTaskUisPaginator(svc, &sagemaker.ListHumanTaskUisInput{}); p.HasMorePages(); {
		pg, e := p.NextPage(ctx)
		if e != nil {
			break
		}
		for _, x := range pg.HumanTaskUiSummaries {
			add(StringValue(x.HumanTaskUiName), "aws_sagemaker_human_task_ui")
		}
	}
	for p := sagemaker.NewListImagesPaginator(svc, &sagemaker.ListImagesInput{}); p.HasMorePages(); {
		pg, e := p.NextPage(ctx)
		if e != nil {
			break
		}
		for _, x := range pg.Images {
			add(StringValue(x.ImageName), "aws_sagemaker_image")
		}
	}
	for p := sagemaker.NewListMlflowTrackingServersPaginator(svc, &sagemaker.ListMlflowTrackingServersInput{}); p.HasMorePages(); {
		pg, e := p.NextPage(ctx)
		if e != nil {
			break
		}
		for _, x := range pg.TrackingServerSummaries {
			add(StringValue(x.TrackingServerName), "aws_sagemaker_mlflow_tracking_server")
		}
	}
	for p := sagemaker.NewListModelPackageGroupsPaginator(svc, &sagemaker.ListModelPackageGroupsInput{}); p.HasMorePages(); {
		pg, e := p.NextPage(ctx)
		if e != nil {
			break
		}
		for _, x := range pg.ModelPackageGroupSummaryList {
			add(StringValue(x.ModelPackageGroupName), "aws_sagemaker_model_package_group")
		}
	}
	for p := sagemaker.NewListMonitoringSchedulesPaginator(svc, &sagemaker.ListMonitoringSchedulesInput{}); p.HasMorePages(); {
		pg, e := p.NextPage(ctx)
		if e != nil {
			break
		}
		for _, x := range pg.MonitoringScheduleSummaries {
			add(StringValue(x.MonitoringScheduleName), "aws_sagemaker_monitoring_schedule")
		}
	}
	for p := sagemaker.NewListNotebookInstanceLifecycleConfigsPaginator(svc, &sagemaker.ListNotebookInstanceLifecycleConfigsInput{}); p.HasMorePages(); {
		pg, e := p.NextPage(ctx)
		if e != nil {
			break
		}
		for _, x := range pg.NotebookInstanceLifecycleConfigs {
			add(StringValue(x.NotebookInstanceLifecycleConfigName), "aws_sagemaker_notebook_instance_lifecycle_configuration")
		}
	}
	for p := sagemaker.NewListPipelinesPaginator(svc, &sagemaker.ListPipelinesInput{}); p.HasMorePages(); {
		pg, e := p.NextPage(ctx)
		if e != nil {
			break
		}
		for _, x := range pg.PipelineSummaries {
			add(StringValue(x.PipelineName), "aws_sagemaker_pipeline")
		}
	}
	for p := sagemaker.NewListProjectsPaginator(svc, &sagemaker.ListProjectsInput{}); p.HasMorePages(); {
		pg, e := p.NextPage(ctx)
		if e != nil {
			break
		}
		for _, x := range pg.ProjectSummaryList {
			add(StringValue(x.ProjectName), "aws_sagemaker_project")
		}
	}
	for p := sagemaker.NewListStudioLifecycleConfigsPaginator(svc, &sagemaker.ListStudioLifecycleConfigsInput{}); p.HasMorePages(); {
		pg, e := p.NextPage(ctx)
		if e != nil {
			break
		}
		for _, x := range pg.StudioLifecycleConfigs {
			add(StringValue(x.StudioLifecycleConfigName), "aws_sagemaker_studio_lifecycle_config")
		}
	}
	for p := sagemaker.NewListWorkforcesPaginator(svc, &sagemaker.ListWorkforcesInput{}); p.HasMorePages(); {
		pg, e := p.NextPage(ctx)
		if e != nil {
			break
		}
		for _, x := range pg.Workforces {
			add(StringValue(x.WorkforceName), "aws_sagemaker_workforce")
		}
	}
	for p := sagemaker.NewListWorkteamsPaginator(svc, &sagemaker.ListWorkteamsInput{}); p.HasMorePages(); {
		pg, e := p.NextPage(ctx)
		if e != nil {
			break
		}
		for _, x := range pg.Workteams {
			add(StringValue(x.WorkteamName), "aws_sagemaker_workteam")
		}
	}
	for p := sagemaker.NewListDataQualityJobDefinitionsPaginator(svc, &sagemaker.ListDataQualityJobDefinitionsInput{}); p.HasMorePages(); {
		pg, e := p.NextPage(ctx)
		if e != nil {
			break
		}
		for _, x := range pg.JobDefinitionSummaries {
			add(StringValue(x.MonitoringJobDefinitionName), "aws_sagemaker_data_quality_job_definition")
		}
	}
	var hubToken *string
	for {
		pg, e := svc.ListHubs(ctx, &sagemaker.ListHubsInput{NextToken: hubToken})
		if e != nil {
			break
		}
		for _, x := range pg.HubSummaries {
			add(StringValue(x.HubName), "aws_sagemaker_hub")
		}
		if pg.NextToken == nil {
			break
		}
		hubToken = pg.NextToken
	}
	for p := sagemaker.NewListUserProfilesPaginator(svc, &sagemaker.ListUserProfilesInput{}); p.HasMorePages(); {
		pg, e := p.NextPage(ctx)
		if e != nil {
			break
		}
		for _, x := range pg.UserProfiles {
			domainID := StringValue(x.DomainId)
			name := StringValue(x.UserProfileName)
			if domainID == "" || name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				domainID+"/"+name, domainID+"_"+name, "aws_sagemaker_user_profile", "aws", defaultAllowEmptyValues))
		}
	}
	for p := sagemaker.NewListSpacesPaginator(svc, &sagemaker.ListSpacesInput{}); p.HasMorePages(); {
		pg, e := p.NextPage(ctx)
		if e != nil {
			break
		}
		for _, x := range pg.Spaces {
			domainID := StringValue(x.DomainId)
			name := StringValue(x.SpaceName)
			if domainID == "" || name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				domainID+"/"+name, domainID+"_"+name, "aws_sagemaker_space", "aws", defaultAllowEmptyValues))
		}
	}
	for p := sagemaker.NewListAppsPaginator(svc, &sagemaker.ListAppsInput{}); p.HasMorePages(); {
		pg, e := p.NextPage(ctx)
		if e != nil {
			break
		}
		for _, x := range pg.Apps {
			domainID := StringValue(x.DomainId)
			appName := StringValue(x.AppName)
			user := StringValue(x.UserProfileName)
			if domainID == "" || appName == "" || user == "" {
				continue
			}
			// import: <domain>/<user-profile>/<app-type>/<app-name>
			id := domainID + "/" + user + "/" + string(x.AppType) + "/" + appName
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, domainID+"_"+user+"_"+appName, "aws_sagemaker_app", "aws", defaultAllowEmptyValues))
		}
	}
}
