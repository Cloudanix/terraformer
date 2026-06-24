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
	"github.com/aws/aws-sdk-go-v2/service/acmpca"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type ACMPCAGenerator struct {
	AWSService
}

// InitResources enumerates ACM Private CA certificate authorities. Import ID is
// the certificate authority ARN.
func (g *ACMPCAGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := acmpca.NewFromConfig(config)

	p := acmpca.NewListCertificateAuthoritiesPaginator(svc, &acmpca.ListCertificateAuthoritiesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, ca := range page.CertificateAuthorities {
			arn := StringValue(ca.Arn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_acmpca_certificate_authority", "aws", defaultAllowEmptyValues))
			// The installed CA certificate is a singleton imported by the CA ARN.
			if cert, err := svc.GetCertificateAuthorityCertificate(awsContext(), &acmpca.GetCertificateAuthorityCertificateInput{CertificateAuthorityArn: ca.Arn}); err == nil && StringValue(cert.Certificate) != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					arn, arn, "aws_acmpca_certificate_authority_certificate", "aws", defaultAllowEmptyValues))
			}
			// A resource-based policy on the CA (cross-account sharing).
			if pol, err := svc.GetPolicy(awsContext(), &acmpca.GetPolicyInput{ResourceArn: ca.Arn}); err == nil && StringValue(pol.Policy) != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					arn, arn, "aws_acmpca_policy", "aws", defaultAllowEmptyValues))
			}
			// ACM service principal permissions delegated on the CA.
			if perms, err := svc.ListPermissions(awsContext(), &acmpca.ListPermissionsInput{CertificateAuthorityArn: ca.Arn}); err == nil {
				for _, p := range perms.Permissions {
					principal := StringValue(p.Principal)
					if principal == "" {
						continue
					}
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						arn+","+principal, principal, "aws_acmpca_permission", "aws", defaultAllowEmptyValues))
				}
			}
		}
	}
	return nil
}
