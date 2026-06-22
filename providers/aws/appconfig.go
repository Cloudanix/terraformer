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
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appconfig"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type AppConfigGenerator struct {
	AWSService
}

// InitResources enumerates AppConfig applications. Import ID is the app id.
func (g *AppConfigGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := appconfig.NewFromConfig(config)

	ctx := context.TODO()
	add := func(id, name, tfType string) {
		if id != "" {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, name, tfType, "aws", defaultAllowEmptyValues))
		}
	}

	var appIDs []string
	p := appconfig.NewListApplicationsPaginator(svc, &appconfig.ListApplicationsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, app := range page.Items {
			id := StringValue(app.Id)
			if id == "" {
				continue
			}
			appIDs = append(appIDs, id)
			add(id, StringValue(app.Name), "aws_appconfig_application")
		}
	}

	for _, appID := range appIDs {
		app := appID
		var envIDs []string
		for ep := appconfig.NewListEnvironmentsPaginator(svc, &appconfig.ListEnvironmentsInput{ApplicationId: &app}); ep.HasMorePages(); {
			page, err := ep.NextPage(ctx)
			if err != nil {
				break
			}
			for _, env := range page.Items {
				id := StringValue(env.Id)
				if id == "" {
					continue
				}
				envIDs = append(envIDs, id)
				add(id+":"+app, StringValue(env.Name), "aws_appconfig_environment")
			}
		}
		for cp := appconfig.NewListConfigurationProfilesPaginator(svc, &appconfig.ListConfigurationProfilesInput{ApplicationId: &app}); cp.HasMorePages(); {
			page, err := cp.NextPage(ctx)
			if err != nil {
				break
			}
			for _, prof := range page.Items {
				pid := StringValue(prof.Id)
				if pid == "" {
					continue
				}
				add(pid+":"+app, StringValue(prof.Name), "aws_appconfig_configuration_profile")
				for hp := appconfig.NewListHostedConfigurationVersionsPaginator(svc, &appconfig.ListHostedConfigurationVersionsInput{ApplicationId: &app, ConfigurationProfileId: aws.String(pid)}); hp.HasMorePages(); {
					hpage, err := hp.NextPage(ctx)
					if err != nil {
						break
					}
					for _, v := range hpage.Items {
						ver := strconv.Itoa(int(v.VersionNumber))
						add(app+"/"+pid+"/"+ver, app+"_"+pid+"_"+ver, "aws_appconfig_hosted_configuration_version")
					}
				}
			}
		}
		for _, envID := range envIDs {
			env := envID
			for dp := appconfig.NewListDeploymentsPaginator(svc, &appconfig.ListDeploymentsInput{ApplicationId: &app, EnvironmentId: &env}); dp.HasMorePages(); {
				dpage, err := dp.NextPage(ctx)
				if err != nil {
					break
				}
				for _, d := range dpage.Items {
					num := strconv.Itoa(int(d.DeploymentNumber))
					add(app+"/"+env+"/"+num, app+"_"+env+"_"+num, "aws_appconfig_deployment")
				}
			}
		}
	}

	for sp := appconfig.NewListDeploymentStrategiesPaginator(svc, &appconfig.ListDeploymentStrategiesInput{}); sp.HasMorePages(); {
		page, err := sp.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, s := range page.Items {
			add(StringValue(s.Id), StringValue(s.Name), "aws_appconfig_deployment_strategy")
		}
	}
	for xp := appconfig.NewListExtensionsPaginator(svc, &appconfig.ListExtensionsInput{}); xp.HasMorePages(); {
		page, err := xp.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range page.Items {
			add(StringValue(x.Id), StringValue(x.Name), "aws_appconfig_extension")
		}
	}
	for xp := appconfig.NewListExtensionAssociationsPaginator(svc, &appconfig.ListExtensionAssociationsInput{}); xp.HasMorePages(); {
		page, err := xp.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range page.Items {
			add(StringValue(x.Id), StringValue(x.Id), "aws_appconfig_extension_association")
		}
	}
	return nil
}
