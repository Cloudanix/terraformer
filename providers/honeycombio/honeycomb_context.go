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

package honeycombio

import "context"

// importRunCtx holds the current import run's context (deadline + Ctrl-C), set by
// the orchestrator via SetContext before each service's sequential discovery.
// ponytail: a package var beats threading ctx through every existing call site;
// safe because InitResources runs sequentially (cmd/import.go).
var importRunCtx context.Context

// runContext returns the import-run context, or context.Background() when unset.
func runContext() context.Context {
	if importRunCtx != nil {
		return importRunCtx
	}
	return context.Background()
}

// SetContext captures the import-run context for runContext().
func (s *HoneycombService) SetContext(ctx context.Context) {
	importRunCtx = ctx
	s.Service.SetContext(ctx)
}
