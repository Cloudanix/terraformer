// Copyright 2019 The Terraformer Authors.
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

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"

	"github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk"
)

var beanstalkAllowEmptyValues = []string{"tags."}

type BeanstalkGenerator struct {
	AWSService
}

func (g *BeanstalkGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	client := elasticbeanstalk.NewFromConfig(config)

	err := g.addApplications(client)
	if err != nil {
		return err
	}
	err = g.addEnvironments(client)
	if err != nil {
		return err
	}
	return g.addApplicationVersions(client)
}

func (g *BeanstalkGenerator) addApplicationVersions(client *elasticbeanstalk.Client) error {
	response, err := client.DescribeApplicationVersions(context.TODO(), &elasticbeanstalk.DescribeApplicationVersionsInput{})
	if err != nil {
		return err
	}
	for _, v := range response.ApplicationVersions {
		app := StringValue(v.ApplicationName)
		label := StringValue(v.VersionLabel)
		if app == "" || label == "" {
			continue
		}
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			app+"/"+label, app+"_"+label, "aws_elastic_beanstalk_application_version", "aws", beanstalkAllowEmptyValues))
	}
	return nil
}

func (g *BeanstalkGenerator) addApplications(client *elasticbeanstalk.Client) error {
	response, err := client.DescribeApplications(context.TODO(), &elasticbeanstalk.DescribeApplicationsInput{})
	if err != nil {
		return err
	}
	for _, application := range response.Applications {
		appName := StringValue(application.ApplicationName)
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			appName,
			appName,
			"aws_elastic_beanstalk_application",
			"aws",
			beanstalkAllowEmptyValues,
		))
		for _, tmpl := range application.ConfigurationTemplates {
			if tmpl == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				appName+"/"+tmpl, appName+"_"+tmpl, "aws_elastic_beanstalk_configuration_template", "aws", beanstalkAllowEmptyValues))
		}
	}
	return nil
}

func (g *BeanstalkGenerator) addEnvironments(client *elasticbeanstalk.Client) error {
	response, err := client.DescribeEnvironments(context.TODO(), &elasticbeanstalk.DescribeEnvironmentsInput{})
	if err != nil {
		return err
	}
	for _, environment := range response.Environments {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			*environment.EnvironmentId,
			*environment.EnvironmentName,
			"aws_elastic_beanstalk_environment",
			"aws",
			beanstalkAllowEmptyValues,
		))
	}
	return nil
}
