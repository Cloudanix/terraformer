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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	datazonetypes "github.com/aws/aws-sdk-go-v2/service/datazone/types"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type DataZoneGenerator struct {
	AWSService
}

// InitResources enumerates DataZone domains. Import ID is the domain id.
func (g *DataZoneGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := datazone.NewFromConfig(config)
	ctx := awsContext()

	var domainIDs []string
	p := datazone.NewListDomainsPaginator(svc, &datazone.ListDomainsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, domain := range page.Items {
			id := StringValue(domain.Id)
			if id == "" {
				continue
			}
			domainIDs = append(domainIDs, id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(domain.Name), "aws_datazone_domain", "aws", defaultAllowEmptyValues))
		}
	}

	for _, domainID := range domainIDs {
		dom := domainID
		var projectIDs []string
		for pp := datazone.NewListProjectsPaginator(svc, &datazone.ListProjectsInput{DomainIdentifier: &dom}); pp.HasMorePages(); {
			page, err := pp.NextPage(ctx)
			if err != nil {
				break
			}
			for _, proj := range page.Items {
				pid := StringValue(proj.Id)
				if pid == "" {
					continue
				}
				projectIDs = append(projectIDs, pid)
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					dom+":"+pid, dom+"_"+pid, "aws_datazone_project", "aws", defaultAllowEmptyValues))
			}
		}
		for _, projectID := range projectIDs {
			pid := projectID
			for ep := datazone.NewListEnvironmentsPaginator(svc, &datazone.ListEnvironmentsInput{DomainIdentifier: &dom, ProjectIdentifier: &pid}); ep.HasMorePages(); {
				page, err := ep.NextPage(ctx)
				if err != nil {
					break
				}
				for _, env := range page.Items {
					eid := StringValue(env.Id)
					if eid == "" {
						continue
					}
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						dom+":"+eid, dom+"_"+eid, "aws_datazone_environment", "aws", defaultAllowEmptyValues))
				}
			}
		}
		for ep := datazone.NewListEnvironmentProfilesPaginator(svc, &datazone.ListEnvironmentProfilesInput{DomainIdentifier: &dom}); ep.HasMorePages(); {
			page, err := ep.NextPage(ctx)
			if err != nil {
				break
			}
			for _, prof := range page.Items {
				pfid := StringValue(prof.Id)
				if pfid == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					dom+":"+pfid, dom+"_"+pfid, "aws_datazone_environment_profile", "aws", defaultAllowEmptyValues))
			}
		}
		// Custom (non-managed) asset & form types via SearchTypes.
		for _, scope := range []datazonetypes.TypesSearchScope{datazonetypes.TypesSearchScopeAssetType, datazonetypes.TypesSearchScopeFormType} {
			for sp := datazone.NewSearchTypesPaginator(svc, &datazone.SearchTypesInput{
				DomainIdentifier: &dom, SearchScope: scope, Managed: aws.Bool(false),
			}); sp.HasMorePages(); {
				page, err := sp.NextPage(ctx)
				if err != nil {
					break
				}
				for _, item := range page.Items {
					switch v := item.(type) {
					case *datazonetypes.SearchTypesResultItemMemberAssetTypeItem:
						name := StringValue(v.Value.Name)
						if name != "" {
							g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
								dom+","+name, dom+"_"+name, "aws_datazone_asset_type", "aws", defaultAllowEmptyValues))
						}
					case *datazonetypes.SearchTypesResultItemMemberFormTypeItem:
						name := StringValue(v.Value.Name)
						if name != "" {
							g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
								dom+","+name, dom+"_"+name, "aws_datazone_form_type", "aws", defaultAllowEmptyValues))
						}
					}
				}
			}
		}
		// User profiles.
		for up := datazone.NewSearchUserProfilesPaginator(svc, &datazone.SearchUserProfilesInput{
			DomainIdentifier: &dom, UserType: datazonetypes.UserSearchTypeDatazoneUser,
		}); up.HasMorePages(); {
			page, err := up.NextPage(ctx)
			if err != nil {
				break
			}
			for _, u := range page.Items {
				uid := StringValue(u.Id)
				if uid == "" {
					continue
				}
				utype := "IAM"
				if u.Type == datazonetypes.UserProfileTypeSso {
					utype = "SSO"
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					uid+","+dom+","+utype, dom+"_"+uid, "aws_datazone_user_profile", "aws", defaultAllowEmptyValues))
			}
		}
		for bp := datazone.NewListEnvironmentBlueprintConfigurationsPaginator(svc, &datazone.ListEnvironmentBlueprintConfigurationsInput{DomainIdentifier: &dom}); bp.HasMorePages(); {
			page, err := bp.NextPage(ctx)
			if err != nil {
				break
			}
			for _, bc := range page.Items {
				bid := StringValue(bc.EnvironmentBlueprintId)
				if bid == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					dom+"/"+bid, dom+"_"+bid, "aws_datazone_environment_blueprint_configuration", "aws", defaultAllowEmptyValues))
			}
		}
	}
	return nil
}
