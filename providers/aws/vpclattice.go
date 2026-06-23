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
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type VPCLatticeGenerator struct {
	AWSService
}

// InitResources enumerates VPC Lattice services and service networks. Import IDs
// are the resource id.
func (g *VPCLatticeGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := vpclattice.NewFromConfig(config)
	ctx := context.TODO()

	// authPolicy emits aws_vpclattice_auth_policy (import resource id) and
	// resourcePolicy emits aws_vpclattice_resource_policy (import resource ARN)
	// when the resource carries one.
	authPolicy := func(id string) {
		if id == "" {
			return
		}
		if out, err := svc.GetAuthPolicy(ctx, &vpclattice.GetAuthPolicyInput{ResourceIdentifier: &id}); err == nil && StringValue(out.Policy) != "" {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_vpclattice_auth_policy", "aws", defaultAllowEmptyValues))
		}
	}
	resourcePolicy := func(arn string) {
		if arn == "" {
			return
		}
		if out, err := svc.GetResourcePolicy(ctx, &vpclattice.GetResourcePolicyInput{ResourceArn: &arn}); err == nil && StringValue(out.Policy) != "" {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_vpclattice_resource_policy", "aws", defaultAllowEmptyValues))
		}
	}

	var serviceIDs []string
	services := vpclattice.NewListServicesPaginator(svc, &vpclattice.ListServicesInput{})
	for services.HasMorePages() {
		page, err := services.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, s := range page.Items {
			id := StringValue(s.Id)
			if id == "" {
				continue
			}
			serviceIDs = append(serviceIDs, id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(s.Name), "aws_vpclattice_service", "aws", defaultAllowEmptyValues))
			authPolicy(id)
			resourcePolicy(StringValue(s.Arn))
		}
	}

	for _, serviceID := range serviceIDs {
		sid := serviceID
		for lp := vpclattice.NewListAccessLogSubscriptionsPaginator(svc, &vpclattice.ListAccessLogSubscriptionsInput{ResourceIdentifier: &sid}); lp.HasMorePages(); {
			page, err := lp.NextPage(ctx)
			if err != nil {
				break
			}
			for _, a := range page.Items {
				id := StringValue(a.Id)
				if id == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, id, "aws_vpclattice_access_log_subscription", "aws", defaultAllowEmptyValues))
			}
		}
		for lp := vpclattice.NewListListenersPaginator(svc, &vpclattice.ListListenersInput{ServiceIdentifier: &sid}); lp.HasMorePages(); {
			page, err := lp.NextPage(ctx)
			if err != nil {
				break
			}
			for _, l := range page.Items {
				lid := StringValue(l.Id)
				if lid == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					sid+"/"+lid, sid+"_"+lid, "aws_vpclattice_listener", "aws", defaultAllowEmptyValues))

				for rp := vpclattice.NewListRulesPaginator(svc, &vpclattice.ListRulesInput{ServiceIdentifier: &sid, ListenerIdentifier: aws.String(lid)}); rp.HasMorePages(); {
					rpage, err := rp.NextPage(ctx)
					if err != nil {
						break
					}
					for _, r := range rpage.Items {
						rid := StringValue(r.Id)
						if rid == "" {
							continue
						}
						g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
							sid+"/"+lid+"/"+rid, sid+"_"+lid+"_"+rid, "aws_vpclattice_listener_rule", "aws", defaultAllowEmptyValues))
					}
				}
			}
		}
	}

	networks := vpclattice.NewListServiceNetworksPaginator(svc, &vpclattice.ListServiceNetworksInput{})
	for networks.HasMorePages() {
		page, err := networks.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, n := range page.Items {
			id := StringValue(n.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(n.Name), "aws_vpclattice_service_network", "aws", defaultAllowEmptyValues))
			authPolicy(id)
			resourcePolicy(StringValue(n.Arn))
		}
	}

	snServiceAssocs := vpclattice.NewListServiceNetworkServiceAssociationsPaginator(svc, &vpclattice.ListServiceNetworkServiceAssociationsInput{})
	for snServiceAssocs.HasMorePages() {
		page, err := snServiceAssocs.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, a := range page.Items {
			id := StringValue(a.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_vpclattice_service_network_service_association", "aws", defaultAllowEmptyValues))
		}
	}

	targetGroups := vpclattice.NewListTargetGroupsPaginator(svc, &vpclattice.ListTargetGroupsInput{})
	for targetGroups.HasMorePages() {
		page, err := targetGroups.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, t := range page.Items {
			id := StringValue(t.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(t.Name), "aws_vpclattice_target_group", "aws", defaultAllowEmptyValues))
			tgID := id
			for tp := vpclattice.NewListTargetsPaginator(svc, &vpclattice.ListTargetsInput{TargetGroupIdentifier: &tgID}); tp.HasMorePages(); {
				tpage, err := tp.NextPage(ctx)
				if err != nil {
					break
				}
				for _, tgt := range tpage.Items {
					targetID := StringValue(tgt.Id)
					if targetID == "" {
						continue
					}
					attID := tgID + "/" + targetID
					if tgt.Port != nil {
						attID = tgID + "/" + targetID + "/" + strconv.Itoa(int(*tgt.Port))
					}
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						attID, attID, "aws_vpclattice_target_group_attachment", "aws", defaultAllowEmptyValues))
				}
			}
		}
	}

	snVpcAssocs := vpclattice.NewListServiceNetworkVpcAssociationsPaginator(svc, &vpclattice.ListServiceNetworkVpcAssociationsInput{})
	for snVpcAssocs.HasMorePages() {
		page, err := snVpcAssocs.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, a := range page.Items {
			id := StringValue(a.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_vpclattice_service_network_vpc_association", "aws", defaultAllowEmptyValues))
		}
	}

	for rg := vpclattice.NewListResourceGatewaysPaginator(svc, &vpclattice.ListResourceGatewaysInput{}); rg.HasMorePages(); {
		page, err := rg.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, r := range page.Items {
			id := StringValue(r.Id)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(r.Name), "aws_vpclattice_resource_gateway", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
