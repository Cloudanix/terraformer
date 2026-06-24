# AGENTS.md

This file provides guidance to coding agents when working with code in this repository.

## What this is

Terraformer — a CLI that generates `tf`/`json` + `tfstate` files from **existing** cloud infrastructure (reverse Terraform). Reads live resources via each provider's SDK, then drives the real Terraform provider plugin to refresh state and emit HCL.

> Upstream (GoogleCloudPlatform/terraformer) is archived/read-only as of 2026-03-16. This is the Cloudanix fork.

## Build / test / lint

```bash
go build -v                      # build binary (./terraformer)
go test ./...                    # full test suite (what CI runs)
go test ./terraformutils/...     # one package
go test ./terraformutils/ -run TestFlatMap   # single test by name
golangci-lint run                # lint (config .golangci.json, v2 schema)
go mod tidy                      # CI runs this; keep go.mod/go.sum clean
```

A `Makefile` wraps these so local == CI (`make help` lists targets). `make ci`
reproduces the test workflow (tidy + build + test); `make check` adds the linter
(`make lint`); `make release` / `make build-providers` mirror the release
workflow. CI (`.github/workflows/test.yml`, `release.yaml`) invokes these targets.

CI matrix: ubuntu + macos, Go version pinned by `go.mod`. CI test job = `make ci`.

Lint enables `gocritic`, `revive`, `unconvert`, `unparam` + `gofmt`/`goimports`. Run `golangci-lint run` (or `make lint`) before committing.

### Running a generated import

```bash
terraformer import aws --resources=ec2_instance,ebs --regions=us-east-1
terraformer import google --resources=gcs,forwardingRules --projects=my-proj --zone=us-central1-a
terraformer plan aws --resources=...     # write plan.json without executing
terraformer import plan path/to/plan.json
```

`--resources=*` imports all supported services for that provider. Provider auth comes from the normal SDK env vars / shared credential files.

## Architecture

Three layers: **cmd** (Cobra CLI + orchestration) → **terraformutils** (provider-agnostic engine) → **providers/<name>** (per-cloud resource discovery).

### Control flow (`cmd/import.go`)

`Import()` is the entry point per provider:
1. `provider.Init(args)` — parse provider flags (region, profile, project…).
2. For each requested service: `provider.InitService(name, verbose)` then `service.InitResources()` discovers resources via the cloud SDK.
3. A `providerwrapper.ProviderWrapper` launches the **real Terraform provider plugin** (go-plugin/gRPC) to `Refresh` each resource's full state — this is why the matching `terraform-provider-*` must be resolvable.
4. `PostConvertHook` / cleanup / filters run, then `terraformoutput` writes `*.tf` + `terraform.tfstate` under the path pattern.

`ProvidersMapping` (`terraformutils/providers_mapping.go`) deep-copies the base provider once per service so services can run/refresh concurrently without sharing mutable state.

### Adding to / modifying a provider

Each provider lives in `providers/<name>/`. Anatomy (AWS as the canonical example):

- `<name>_provider.go` — implements `terraformutils.ProviderGenerator` (`terraformutils/base_provider.go`). Key method `GetSupportedService()` returns `map[serviceName]ServiceGenerator` — **this is the registry**; add a service by adding a map entry here.
- `<service>.go` — one file per cloud service. Defines a `XxxGenerator` struct embedding the provider's base service, and implements `InitResources()` to enumerate live resources and append `terraformutils.NewSimpleResource(id, name, tfType, providerName, allowEmptyValues)` to `g.Resources`. Optional `PostConvertHook()` rewrites attributes after refresh (e.g. wrapping IAM policy JSON in heredocs — see `providers/aws/ecr.go`).
- AWS wraps every generator in `AwsFacade` (`aws_facade.go`); other providers register generators directly.

To register a new provider end-to-end: add the importer constructor to `providerImporterSubcommands()` in `cmd/root.go`, a `cmd/provider_cmd_<name>.go` (Cobra command + flag wiring), and a `docs/<name>.md`.

### terraformutils engine (provider-agnostic)

- `service.go` / `resource.go` — `ServiceGenerator` / `Service` base + `Resource` model. `Service` provides filter parsing (`--filter`), cleanup, ignore-keys.
- `providerwrapper/` — spawns and speaks gRPC to the real terraform provider plugin (schema, refresh, ignore-keys).
- `hcl.go`, `flatmap.go`, `json.go`, `walk.go` — convert refreshed flatmap state ↔ HCL/JSON, walk nested attributes (used by filters and output).
- `terraformoutput/` — writes the final `.tf`/`.tfstate` files.

### AWS region handling (subtle, don't break)

AWS import splits resources into **global** (region `aws-global` / `GlobalRegion`), **us-east-1-only**, and **regional** groups and imports them in separate passes within one process (`cmd/provider_cmd_aws.go` → `parseAndGroupResources`). Because passes share a process, `aws_service.go` caches the SDK `aws.Config` **per region** (`configCache` keyed by region) — a single shared config would pin every later region to the first pass's endpoint, causing wrong-region signing / `aws-global` DNS failures. Keep config caching region-scoped.

## Conventions

- Every source file carries the Apache 2.0 header — copy it into new files.
- Resource creation goes through `terraformutils.NewSimpleResource` / `NewResource`; don't hand-build `Resource` structs.
- `AllowEmptyValues` (e.g. `"tags."`) lists attribute prefixes that may be empty without being dropped — set per service.

## Token Efficiency (MANDATORY)

Token optimization is not optional. Use every tool below on every session.

### Setup — Install All Tools

Check and install once per machine:

```bash
# 1. RTK — CLI token proxy
brew install rtk
rtk --version   # verify: rtk X.Y.Z

# 2. caveman — terse communication style plugin
claude plugin install caveman@caveman

# 3. context-mode — context window protection plugin
claude plugin install context-mode@context-mode

# 4. code-review-graph — structural code knowledge graph
pip install code-review-graph
code-review-graph --version   # verify: code-review-graph X.Y.Z
# then build the graph for this repo:
code-review-graph build .
```

After install, restart Claude Code to activate plugins.

### Communication Style — Caveman Mode

Respond terse. Drop: articles (a/an/the), filler (just/really/basically/actually), pleasantries (sure/certainly/happy to), hedging. Fragments OK. Short synonyms (fix not "implement a solution for"). Technical terms exact. Code blocks unchanged.

Auto-expand ONLY for: security warnings, irreversible action confirmations, user confusion.

### RTK — CLI Token Proxy (60-90% savings on dev ops)

All shell commands route through RTK automatically (hook-based, transparent). Use `rtk` directly only for meta commands:
```bash
rtk gain              # token savings analytics
rtk gain --history    # command history with savings
rtk discover          # missed optimization opportunities
rtk proxy <cmd>       # raw command without filtering (debug)
```

Verify install: `rtk --version`. If `rtk gain` fails, may have wrong `rtk` binary (Rust Type Kit collision) — check `which rtk`.

### context-mode — Context Window Protection

Raw tool output floods context. **MUST** use context-mode MCP tools. Only printed summary enters context.

**Tool hierarchy (strict order):**

| Priority | Tool | Use for |
|----------|------|---------|
| 1st | `ctx_batch_execute(commands, queries)` | Research — runs commands + indexes + searches in one call |
| 2nd | `ctx_search(queries: [...])` | All follow-up questions — one call, many queries |
| 3rd | `ctx_execute(language, code)` / `ctx_execute_file(path, language, code)` | API calls, log analysis, data processing |

**Forbidden patterns:**
- NO Bash for commands producing >20 lines output → use `ctx_execute`
- NO Read for analysis → use `ctx_execute_file` (Read is correct only when you need to Edit the file)
- NO WebFetch → use `ctx_fetch_and_index`
- Bash ONLY for: `git`, `mkdir`, `rm`, `mv`, navigation

### MCP Tools: code-review-graph

**ALWAYS use code-review-graph tools BEFORE Grep/Glob/Read.** Faster, cheaper, gives structural context (callers, dependents, test coverage).

| Tool | Use when |
|------|----------|
| `detect_changes` | Code review — risk-scored analysis |
| `get_review_context` | Source snippets — token-efficient |
| `get_impact_radius` | Blast radius of a change |
| `get_affected_flows` | Impacted execution paths |
| `query_graph` | Trace callers/callees/imports/tests |
| `semantic_search_nodes` | Find functions/classes by name/keyword |
| `get_architecture_overview` | High-level structure |
| `refactor_tool` | Renames, dead code |

Fallback to Grep/Glob/Read only when graph insufficient. Graph auto-updates on file changes.

### Grep alternative: fff

For any file search or grep in the current git-indexed directory, use fff tools.

## Individual Settings
Refer to `AGENTS.LOCAL.MD` if it exists.
