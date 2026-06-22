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

	"github.com/aws/aws-sdk-go-v2/service/signer"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type SignerGenerator struct {
	AWSService
}

// InitResources enumerates Signer signing profiles. Import ID is the profile name.
func (g *SignerGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := signer.NewFromConfig(config)

	p := signer.NewListSigningProfilesPaginator(svc, &signer.ListSigningProfilesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, profile := range page.Profiles {
			name := StringValue(profile.ProfileName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_signer_signing_profile", "aws", defaultAllowEmptyValues))
			profileName := name
			if perms, err := svc.ListProfilePermissions(context.TODO(), &signer.ListProfilePermissionsInput{ProfileName: &profileName}); err == nil {
				for _, perm := range perms.Permissions {
					sid := StringValue(perm.StatementId)
					if sid == "" {
						continue
					}
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						profileName+"/"+sid, profileName+"_"+sid, "aws_signer_signing_profile_permission", "aws", defaultAllowEmptyValues))
				}
			}
		}
	}

	jobInput := &signer.ListSigningJobsInput{}
	for {
		out, err := svc.ListSigningJobs(context.TODO(), jobInput)
		if err != nil {
			break
		}
		for _, job := range out.Jobs {
			jobID := StringValue(job.JobId)
			if jobID == "" || job.Status != "Succeeded" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				jobID, jobID, "aws_signer_signing_job", "aws", defaultAllowEmptyValues))
		}
		if out.NextToken == nil {
			break
		}
		jobInput.NextToken = out.NextToken
	}
	return nil
}
