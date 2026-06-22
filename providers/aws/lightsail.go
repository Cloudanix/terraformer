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

	"github.com/aws/aws-sdk-go-v2/service/lightsail"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type LightsailGenerator struct {
	AWSService
}

// InitResources enumerates Lightsail instances. Import ID is the instance name.
func (g *LightsailGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := lightsail.NewFromConfig(config)

	var token *string
	for {
		out, err := svc.GetInstances(context.TODO(), &lightsail.GetInstancesInput{PageToken: token})
		if err != nil {
			return err
		}
		for _, instance := range out.Instances {
			name := StringValue(instance.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_lightsail_instance", "aws", defaultAllowEmptyValues))
		}
		if out.NextPageToken == nil {
			break
		}
		token = out.NextPageToken
	}

	g.addLightsailExtras(svc)
	return nil
}

// addLightsailExtras enumerates the other top-level Lightsail resources, each a
// Get* returning a named list. Import ID is the resource name. Errors are
// logged via a skip so one missing permission doesn't abort the whole import.
func (g *LightsailGenerator) addLightsailExtras(svc *lightsail.Client) {
	ctx := context.TODO()
	add := func(name, tfType string) {
		if name != "" {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, tfType, "aws", defaultAllowEmptyValues))
		}
	}

	if out, err := svc.GetBuckets(ctx, &lightsail.GetBucketsInput{}); err == nil {
		for _, x := range out.Buckets {
			add(StringValue(x.Name), "aws_lightsail_bucket")
		}
	}
	if out, err := svc.GetCertificates(ctx, &lightsail.GetCertificatesInput{}); err == nil {
		for _, x := range out.Certificates {
			add(StringValue(x.CertificateName), "aws_lightsail_certificate")
		}
	}
	if out, err := svc.GetContainerServices(ctx, &lightsail.GetContainerServicesInput{}); err == nil {
		for _, x := range out.ContainerServices {
			add(StringValue(x.ContainerServiceName), "aws_lightsail_container_service")
		}
	}
	if out, err := svc.GetRelationalDatabases(ctx, &lightsail.GetRelationalDatabasesInput{}); err == nil {
		for _, x := range out.RelationalDatabases {
			add(StringValue(x.Name), "aws_lightsail_database")
		}
	}
	if out, err := svc.GetDisks(ctx, &lightsail.GetDisksInput{}); err == nil {
		for _, x := range out.Disks {
			add(StringValue(x.Name), "aws_lightsail_disk")
		}
	}
	if out, err := svc.GetDistributions(ctx, &lightsail.GetDistributionsInput{}); err == nil {
		for _, x := range out.Distributions {
			add(StringValue(x.Name), "aws_lightsail_distribution")
		}
	}
	if out, err := svc.GetDomains(ctx, &lightsail.GetDomainsInput{}); err == nil {
		for _, x := range out.Domains {
			add(StringValue(x.Name), "aws_lightsail_domain")
		}
	}
	if out, err := svc.GetKeyPairs(ctx, &lightsail.GetKeyPairsInput{}); err == nil {
		for _, x := range out.KeyPairs {
			add(StringValue(x.Name), "aws_lightsail_key_pair")
		}
	}
	if out, err := svc.GetLoadBalancers(ctx, &lightsail.GetLoadBalancersInput{}); err == nil {
		for _, x := range out.LoadBalancers {
			add(StringValue(x.Name), "aws_lightsail_lb")
		}
	}
	if out, err := svc.GetStaticIps(ctx, &lightsail.GetStaticIpsInput{}); err == nil {
		for _, x := range out.StaticIps {
			add(StringValue(x.Name), "aws_lightsail_static_ip")
		}
	}
}
