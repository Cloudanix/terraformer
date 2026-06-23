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

package gcp

import (
	"context"
	"log"
	"strings"

	"google.golang.org/api/identitytoolkit/v2"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var identityPlatformAllowEmptyValues = []string{""}

var identityPlatformAdditionalFields = map[string]interface{}{}

type IdentityPlatformGenerator struct {
	GCPService
}

func last(s string) string {
	t := strings.Split(s, "/")
	return t[len(t)-1]
}

// walkIdpConfigs enumerates the three idp-config collections at the given parent (project or tenant scope).
func (g *IdentityPlatformGenerator) walkIdpConfigs(ctx context.Context, svc *identitytoolkit.Service, parent, project, tenant string) {
	defAttrs := func(id string) map[string]string {
		m := map[string]string{"client_id": id, "project": project}
		if tenant != "" {
			m["tenant"] = tenant
		}
		return m
	}
	if tenant == "" {
		if err := svc.Projects.DefaultSupportedIdpConfigs.List(parent).Pages(ctx, func(p *identitytoolkit.GoogleCloudIdentitytoolkitAdminV2ListDefaultSupportedIdpConfigsResponse) error {
			for _, o := range p.DefaultSupportedIdpConfigs {
				g.Resources = append(g.Resources, terraformutils.NewResource(o.Name, last(o.Name), "google_identity_platform_default_supported_idp_config", g.ProviderName, defAttrs(last(o.Name)), identityPlatformAllowEmptyValues, identityPlatformAdditionalFields))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
		if err := svc.Projects.InboundSamlConfigs.List(parent).Pages(ctx, func(p *identitytoolkit.GoogleCloudIdentitytoolkitAdminV2ListInboundSamlConfigsResponse) error {
			for _, o := range p.InboundSamlConfigs {
				g.Resources = append(g.Resources, terraformutils.NewResource(o.Name, last(o.Name), "google_identity_platform_inbound_saml_config", g.ProviderName, defAttrs(last(o.Name)), identityPlatformAllowEmptyValues, identityPlatformAdditionalFields))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
		if err := svc.Projects.OauthIdpConfigs.List(parent).Pages(ctx, func(p *identitytoolkit.GoogleCloudIdentitytoolkitAdminV2ListOAuthIdpConfigsResponse) error {
			for _, o := range p.OauthIdpConfigs {
				g.Resources = append(g.Resources, terraformutils.NewResource(o.Name, last(o.Name), "google_identity_platform_oauth_idp_config", g.ProviderName, defAttrs(last(o.Name)), identityPlatformAllowEmptyValues, identityPlatformAdditionalFields))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
		return
	}
	if err := svc.Projects.Tenants.DefaultSupportedIdpConfigs.List(parent).Pages(ctx, func(p *identitytoolkit.GoogleCloudIdentitytoolkitAdminV2ListDefaultSupportedIdpConfigsResponse) error {
		for _, o := range p.DefaultSupportedIdpConfigs {
			g.Resources = append(g.Resources, terraformutils.NewResource(o.Name, last(o.Name), "google_identity_platform_tenant_default_supported_idp_config", g.ProviderName, defAttrs(last(o.Name)), identityPlatformAllowEmptyValues, identityPlatformAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := svc.Projects.Tenants.InboundSamlConfigs.List(parent).Pages(ctx, func(p *identitytoolkit.GoogleCloudIdentitytoolkitAdminV2ListInboundSamlConfigsResponse) error {
		for _, o := range p.InboundSamlConfigs {
			g.Resources = append(g.Resources, terraformutils.NewResource(o.Name, last(o.Name), "google_identity_platform_tenant_inbound_saml_config", g.ProviderName, defAttrs(last(o.Name)), identityPlatformAllowEmptyValues, identityPlatformAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := svc.Projects.Tenants.OauthIdpConfigs.List(parent).Pages(ctx, func(p *identitytoolkit.GoogleCloudIdentitytoolkitAdminV2ListOAuthIdpConfigsResponse) error {
		for _, o := range p.OauthIdpConfigs {
			g.Resources = append(g.Resources, terraformutils.NewResource(o.Name, last(o.Name), "google_identity_platform_tenant_oauth_idp_config", g.ProviderName, defAttrs(last(o.Name)), identityPlatformAllowEmptyValues, identityPlatformAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
}

// Generate TerraformResources from GCP API,
func (g *IdentityPlatformGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := identitytoolkit.NewService(ctx)
	if err != nil {
		return err
	}
	project := g.GetArgs()["project"].(string)
	projParent := "projects/" + project

	// Project-level config singleton.
	g.Resources = append(g.Resources, terraformutils.NewResource(
		projParent+"/config", project, "google_identity_platform_config", g.ProviderName,
		map[string]string{"project": project}, identityPlatformAllowEmptyValues, identityPlatformAdditionalFields))

	g.walkIdpConfigs(ctx, svc, projParent, project, "")

	if err := svc.Projects.Tenants.List(projParent).Pages(ctx, func(p *identitytoolkit.GoogleCloudIdentitytoolkitAdminV2ListTenantsResponse) error {
		for _, t := range p.Tenants {
			tenant := last(t.Name)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				t.Name, tenant, "google_identity_platform_tenant", g.ProviderName,
				map[string]string{"name": tenant, "project": project}, identityPlatformAllowEmptyValues, identityPlatformAdditionalFields))
			g.walkIdpConfigs(ctx, svc, t.Name, project, tenant)
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}
