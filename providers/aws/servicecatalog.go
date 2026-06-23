// Copyright 2020 The Terraformer Authors.
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
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalog"
)

var servicecatalogAllowEmptyValues = []string{"tags."}

type ServiceCatalogGenerator struct {
	AWSService
}

func (g *ServiceCatalogGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := servicecatalog.NewFromConfig(config)
	p := servicecatalog.NewListPortfoliosPaginator(svc, &servicecatalog.ListPortfoliosInput{})
	var resources []terraformutils.Resource
	var portfolioIDs []string
	for p.HasMorePages() {
		page, err := p.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, portfolio := range page.PortfolioDetails {
			portfolioID := StringValue(portfolio.Id)
			portfolioName := StringValue(portfolio.DisplayName)
			portfolioIDs = append(portfolioIDs, portfolioID)
			resources = append(resources, terraformutils.NewSimpleResource(
				portfolioID,
				portfolioName,
				"aws_servicecatalog_portfolio",
				"aws",
				servicecatalogAllowEmptyValues))
		}
	}

	for _, portfolioID := range portfolioIDs {
		if portfolioID == "" {
			continue
		}
		cp := servicecatalog.NewListConstraintsForPortfolioPaginator(svc, &servicecatalog.ListConstraintsForPortfolioInput{PortfolioId: &portfolioID})
		for cp.HasMorePages() {
			page, err := cp.NextPage(awsContext())
			if err != nil {
				break
			}
			for _, c := range page.ConstraintDetails {
				id := StringValue(c.ConstraintId)
				if id == "" {
					continue
				}
				resources = append(resources, terraformutils.NewSimpleResource(
					id, id, "aws_servicecatalog_constraint", "aws", servicecatalogAllowEmptyValues))
			}
		}
		if budgets, err := svc.ListBudgetsForResource(awsContext(), &servicecatalog.ListBudgetsForResourceInput{ResourceId: &portfolioID}); err == nil {
			for _, b := range budgets.Budgets {
				bn := StringValue(b.BudgetName)
				if bn == "" {
					continue
				}
				resources = append(resources, terraformutils.NewSimpleResource(
					bn+":"+portfolioID, bn+"_"+portfolioID, "aws_servicecatalog_budget_resource_association", "aws", servicecatalogAllowEmptyValues))
			}
		}
		for ppp := servicecatalog.NewListPrincipalsForPortfolioPaginator(svc, &servicecatalog.ListPrincipalsForPortfolioInput{PortfolioId: &portfolioID}); ppp.HasMorePages(); {
			page, err := ppp.NextPage(awsContext())
			if err != nil {
				break
			}
			for _, pr := range page.Principals {
				arn := StringValue(pr.PrincipalARN)
				if arn == "" {
					continue
				}
				id := "en," + arn + "," + portfolioID + "," + string(pr.PrincipalType)
				resources = append(resources, terraformutils.NewSimpleResource(
					id, portfolioID+"_principal", "aws_servicecatalog_principal_portfolio_association", "aws", servicecatalogAllowEmptyValues))
			}
		}
		for prod := servicecatalog.NewSearchProductsAsAdminPaginator(svc, &servicecatalog.SearchProductsAsAdminInput{PortfolioId: &portfolioID}); prod.HasMorePages(); {
			page, err := prod.NextPage(awsContext())
			if err != nil {
				break
			}
			for _, pvd := range page.ProductViewDetails {
				if pvd.ProductViewSummary == nil {
					continue
				}
				productID := StringValue(pvd.ProductViewSummary.ProductId)
				if productID == "" {
					continue
				}
				resources = append(resources, terraformutils.NewSimpleResource(
					"en:"+productID+":"+portfolioID, portfolioID+"_"+productID, "aws_servicecatalog_product_portfolio_association", "aws", servicecatalogAllowEmptyValues))
			}
		}
		if access, err := svc.ListPortfolioAccess(awsContext(), &servicecatalog.ListPortfolioAccessInput{PortfolioId: &portfolioID}); err == nil {
			for _, acct := range access.AccountIds {
				if acct == "" {
					continue
				}
				id := portfolioID + ":ACCOUNT:" + acct
				resources = append(resources, terraformutils.NewSimpleResource(
					id, portfolioID+"_"+acct, "aws_servicecatalog_portfolio_share", "aws", servicecatalogAllowEmptyValues))
			}
		}
	}

	pp := servicecatalog.NewSearchProductsAsAdminPaginator(svc, &servicecatalog.SearchProductsAsAdminInput{})
	for pp.HasMorePages() {
		page, err := pp.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, pvd := range page.ProductViewDetails {
			if pvd.ProductViewSummary == nil {
				continue
			}
			productID := StringValue(pvd.ProductViewSummary.ProductId)
			if productID == "" {
				continue
			}
			resources = append(resources, terraformutils.NewSimpleResource(
				productID,
				StringValue(pvd.ProductViewSummary.Name),
				"aws_servicecatalog_product",
				"aws",
				servicecatalogAllowEmptyValues))
			if arts, err := svc.ListProvisioningArtifacts(awsContext(), &servicecatalog.ListProvisioningArtifactsInput{ProductId: &productID}); err == nil {
				for _, a := range arts.ProvisioningArtifactDetails {
					aid := StringValue(a.Id)
					if aid == "" {
						continue
					}
					resources = append(resources, terraformutils.NewSimpleResource(
						aid+":"+productID, aid, "aws_servicecatalog_provisioning_artifact", "aws", servicecatalogAllowEmptyValues))
				}
			}
		}
	}

	for sp := servicecatalog.NewSearchProvisionedProductsPaginator(svc, &servicecatalog.SearchProvisionedProductsInput{}); sp.HasMorePages(); {
		page, err := sp.NextPage(awsContext())
		if err != nil {
			break
		}
		for _, pp := range page.ProvisionedProducts {
			id := StringValue(pp.Id)
			if id == "" {
				continue
			}
			resources = append(resources, terraformutils.NewSimpleResource(
				id, StringValue(pp.Name), "aws_servicecatalog_provisioned_product", "aws", servicecatalogAllowEmptyValues))
		}
	}

	tagOptions := servicecatalog.NewListTagOptionsPaginator(svc, &servicecatalog.ListTagOptionsInput{})
	for tagOptions.HasMorePages() {
		page, err := tagOptions.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, t := range page.TagOptionDetails {
			id := StringValue(t.Id)
			if id == "" {
				continue
			}
			resources = append(resources, terraformutils.NewSimpleResource(
				id, id, "aws_servicecatalog_tag_option", "aws", servicecatalogAllowEmptyValues))
			tagOptionID := id
			for rp := servicecatalog.NewListResourcesForTagOptionPaginator(svc, &servicecatalog.ListResourcesForTagOptionInput{TagOptionId: &tagOptionID}); rp.HasMorePages(); {
				rpage, err := rp.NextPage(awsContext())
				if err != nil {
					break
				}
				for _, r := range rpage.ResourceDetails {
					resourceID := StringValue(r.Id)
					if resourceID == "" {
						continue
					}
					resources = append(resources, terraformutils.NewSimpleResource(
						tagOptionID+":"+resourceID, tagOptionID+"_"+resourceID, "aws_servicecatalog_tag_option_resource_association", "aws", servicecatalogAllowEmptyValues))
				}
			}
		}
	}

	serviceActions := servicecatalog.NewListServiceActionsPaginator(svc, &servicecatalog.ListServiceActionsInput{})
	for serviceActions.HasMorePages() {
		page, err := serviceActions.NextPage(awsContext())
		if err != nil {
			return err
		}
		for _, sa := range page.ServiceActionSummaries {
			id := StringValue(sa.Id)
			if id == "" {
				continue
			}
			resources = append(resources, terraformutils.NewSimpleResource(
				id, StringValue(sa.Name), "aws_servicecatalog_service_action", "aws", servicecatalogAllowEmptyValues))
		}
	}

	g.Resources = resources
	return nil
}
