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
	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas/types"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type ServiceQuotasGenerator struct {
	AWSService
}

// quotaRef is a deduplicated, importable service-quota reference.
type quotaRef struct {
	id   string // "<service-code>/<quota-code>" — the aws_servicequotas_service_quota import ID
	name string // terraform resource name
}

// quotasFromChangeHistory turns the account's quota change-request history into
// a deduplicated set of importable quotas. Only quotas the account has actually
// requested a change for are emitted — not the thousands of untouched AWS
// defaults (see TODOS.md T5). Multiple change requests for the same quota
// collapse to one resource. Items missing a service or quota code are skipped
// (un-importable). Order is preserved (first occurrence wins) for stable output.
func quotasFromChangeHistory(changes []types.RequestedServiceQuotaChange) []quotaRef {
	seen := map[string]bool{}
	var refs []quotaRef
	for _, c := range changes {
		serviceCode := StringValue(c.ServiceCode)
		quotaCode := StringValue(c.QuotaCode)
		if serviceCode == "" || quotaCode == "" {
			continue
		}
		id := serviceCode + "/" + quotaCode
		if seen[id] {
			continue
		}
		seen[id] = true
		name := StringValue(c.QuotaName)
		if name == "" {
			name = id
		}
		refs = append(refs, quotaRef{id: id, name: name})
	}
	return refs
}

// InitResources emits aws_servicequotas_service_quota only for quotas the
// account has requested a change for (via the change-request history) — the
// "user-managed quotas" signal — instead of dumping every default quota for
// every service. Import ID is "<service-code>/<quota-code>".
func (g *ServiceQuotasGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := servicequotas.NewFromConfig(config)

	var changes []types.RequestedServiceQuotaChange
	p := servicequotas.NewListRequestedServiceQuotaChangeHistoryPaginator(svc, &servicequotas.ListRequestedServiceQuotaChangeHistoryInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		changes = append(changes, page.RequestedQuotas...)
	}

	for _, ref := range quotasFromChangeHistory(changes) {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			ref.id, ref.name, "aws_servicequotas_service_quota", "aws", defaultAllowEmptyValues))
	}

	for tp := servicequotas.NewListServiceQuotaIncreaseRequestsInTemplatePaginator(svc, &servicequotas.ListServiceQuotaIncreaseRequestsInTemplateInput{}); tp.HasMorePages(); {
		page, err := tp.NextPage(awsContext())
		if err != nil {
			break
		}
		for _, t := range page.ServiceQuotaIncreaseRequestInTemplateList {
			quota := StringValue(t.QuotaCode)
			service := StringValue(t.ServiceCode)
			region := StringValue(t.AwsRegion)
			if quota == "" || service == "" || region == "" {
				continue
			}
			id := quota + "/" + service + "/" + region
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_servicequotas_template", "aws", defaultAllowEmptyValues))
		}
	}

	// Template association is an account singleton (ASSOCIATED when the quota
	// template is applied org-wide); imported by account id.
	if assoc, err := svc.GetAssociationForServiceQuotaTemplate(awsContext(), &servicequotas.GetAssociationForServiceQuotaTemplateInput{}); err == nil &&
		assoc.ServiceQuotaTemplateAssociationStatus == types.ServiceQuotaTemplateAssociationStatusAssociated {
		if account, err := g.getAccountNumber(config); err == nil {
			id := StringValue(account)
			if id != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, id, "aws_servicequotas_template_association", "aws", defaultAllowEmptyValues))
			}
		}
	}
	return nil
}
