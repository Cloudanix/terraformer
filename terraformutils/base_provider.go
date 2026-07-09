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

package terraformutils

import (
	"github.com/zclconf/go-cty/cty"
)

type ProviderGenerator interface {
	Init(args []string) error
	InitService(serviceName string, verbose bool) error
	GetName() string
	GetService() ServiceGenerator
	GetConfig() cty.Value
	GetBasicConfig() cty.Value
	GetSupportedService() map[string]ServiceGenerator
	GenerateFiles()
	GetProviderData(arg ...string) map[string]interface{}
	GenerateOutputPath() error
	GetResourceConnections() map[string]map[string][]string
}

type ProviderWithSource interface {
	GetSource() string
}

// CredentialRefresher is an optional interface a ProviderGenerator may
// implement to re-resolve (and re-bridge to the provider plugin) fresh
// credentials mid-run. The engine calls RefreshCredentials before the long
// refresh phase and again whenever a plugin call fails with an expired-token
// error, so a run that outlives a temporary session token (SSO / assumed role)
// can keep going instead of drowning in stale-token errors. Best-effort:
// implementations return an error only so the caller can log it.
type CredentialRefresher interface {
	RefreshCredentials() error
}

// CallerIdentityLogger is an optional interface a ProviderGenerator may
// implement to report which principal / credentials its API calls run as. The
// engine calls LogCallerIdentity once, right after Init, so every import records
// the identity (and thus the privilege) in effect — useful for understanding
// what an import could and couldn't read.
//
// Implementations MUST be best-effort: log what they can and never fail the
// import. Never log secret material (secret keys, client secrets, raw tokens) —
// only identifiers (ARNs, principals, subscription/project/tenant ids) and
// credential metadata (temporary vs long-lived, expiry).
type CallerIdentityLogger interface {
	LogCallerIdentity()
}

type Provider struct {
	Service ServiceGenerator
	Config  cty.Value
}

func (p *Provider) Init(args []string) error {
	panic("implement me")
}

func (p *Provider) GetConfig() cty.Value {
	return p.Config
}

func (p *Provider) GetName() string {
	panic("implement me")
}

func (p *Provider) InitService(serviceName string) error {
	panic("implement me")
}

func (p *Provider) GenerateOutputPath() error {
	panic("implement me")
}

func (p *Provider) GenerateFiles() {
	panic("implement me")
}

func (p *Provider) GetService() ServiceGenerator {
	return p.Service
}

func (p *Provider) GetSupportedService() map[string]ServiceGenerator {
	panic("implement me")
}

func (p *Provider) GetBasicConfig() cty.Value {
	return cty.ObjectVal(map[string]cty.Value{})
}
