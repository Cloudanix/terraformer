# Azure Full-Support — Execution Phases

Execution-order source of truth. Supersedes `plan.md` §7 (which is migrate-first).
`plan.md` stays the reference for *how* (recipes §5/§6, constraints §1, gap method §3).

## Decisions (locked 2026-06-23)

1. **Value-first sequencing.** Build foundations + dual auth path, add NEW Track-2
   coverage first. Migrating the 35 working Track-1 services is debt — done LAST,
   not before coverage.
2. **Depth: P1+P2, then reassess.** §4b partial gaps + P1/P2 new services. Stop
   before the P3/P4 long tail; re-evaluate with real demand.
3. **azuread IN scope.** Separate `providers/azuread/` stream (own SDK + auth),
   runs parallel to azurerm work after Phase 0.
4. **Pin azurerm v4.78.0.** Gap measured against it. Document as floor in `docs/azure.md`.

Dual-SDK is fine mid-flight. Track-1 stays alive until Phase 5. A file may mix
Track-1 (existing) + Track-2 (new enumeration) temporarily — acceptable, tracked.

Re-run `plan.md` §3 after each PR; shrink `missing-resources.txt`.

---

## Phase 0 — Foundations (§2a) + dual auth — BLOCKING, nothing fans out before this

No migration of existing services. Build shared Track-2 pieces once.

- `azidentity` credential in `azure_provider.go`, injected via `Service.Args`
  **alongside** the existing autorest `authorizer` (keep both).
- `getClientOptions() (subID string, cred azcore.TokenCredential, opts *arm.ClientOptions)`.
- `appendSimpleResources[T]` generic (pager→`NewSimpleResource`, nil-safe, skips empty IDs).
- `valueOrEmpty` helper; `defaultAllowEmptyValues = []string{"tags."}` package var.
- RG-iteration helper (loop the `-R name1:name2` set).
- Sovereign cloud via `arm.ClientOptions.Cloud`; retry via `arm.ClientOptions.Retry`.
- Unit tests for the helpers (high fan-in — bug here = bug everywhere).
- Pin v4.78.0; record floor in `docs/azure.md`.

**Gate:** `go build` green; existing 35 services import unchanged (behavior preserved).

## Phase 1 — Partial-service gaps (§4b) — highest ROI, services already wired

New resource types into existing generators. Use a new Track-2 `arm` client for the
new enumeration (don't migrate the whole file). Start order:

`key_vault` (key/secret/certificate/access_policy) · `app_service`
(service_plan, linux/windows web_app + function_app, slots) · `storage_account`
(queue/table/share/management_policy/network_rules/data_lake_gen2) · `cosmosdb`
(mongo/cassandra/gremlin/sql_function) · `virtual_network`
(peering, gateways, connections) · `database` (mysql/postgresql flexible servers,
mssql managed_instance) · `security_center` (setting/workspace/assessment/automation).

Each = one PR (gap blocks + `PostConvertHook` if needed + `docs/azure.md`).

## Phase 2 — P1 new services (§4a) — greenfield Track-2, zero migration debt

`kubernetes` (AKS) · `monitor` · `log_analytics` · `application_insights` ·
`role_assignment` (RBAC, subscription-scoped) · `policy` · `managed_identity` ·
`function_app` · `recovery_services` · `automation` · `cdn`/`frontdoor` ·
`traffic_manager` · `firewall` · `nat_gateway` · `virtual_wan`.

Recipe = `plan.md` §5. One service per PR. Mark subscription-scoped services
(role_assignment, policy, management_group) — they ignore `-R`, list subscription-wide.

## Phase 3 — P2 security/governance (§4a P2)

`sentinel` · `ddos` · `bastion` · `management_group` · `lighthouse` ·
`private_dns_resolver`. One service per PR.

**→ Reassess here.** P3/P4 (servicebus, apim, cognitive, ML, kusto, iothub,
eventgrid, container_app, flexible-tier DBs, long tail) built on demand only.

## Phase 4 — azuread provider (parallel stream, starts after Phase 0)

Independent of azurerm work — different SDK (`microsoft-graph`/`azidentity`),
different files, no conflict. Can run concurrently with Phase 2/3.

- Scaffold `providers/azuread/`: `azuread_provider.go` (ProviderGenerator),
  `azuread_service.go` (base + auth), register in `cmd/root.go` +
  `cmd/provider_cmd_azuread.go`, `docs/azuread.md`.
- Services: `users`, `groups`, `applications`, `service_principals` (+ app roles,
  directory roles as gaps).
- Gap method mirrors §3 against `terraform-provider-azuread`.

## Phase 5 — Migrate 35 Track-1 services to Track-2 — debt paydown, LAST

Behavior-preserving, no new coverage. Each PR = no-diff `terraform plan` round-trip
(§8). Order: leaf (resource_group, disk, public_ip, ssh_public_key, redis,
management_lock) → networking cluster (shared `armnetwork`) → compute/storage/db →
long tail (`data_factory` biggest). Drop `azure-sdk-for-go/services/*`, go-autorest,
hamilton from go.mod; `go mod tidy`.

**Done when:** `grep -r "azure-sdk-for-go/services" providers/azure` is empty AND
P1+P2 gap closed (modulo `no-list-api.md` exclusions).

---

## Dependency graph

```
Phase 0 (blocking)
 ├─ Phase 1 ─ Phase 2 ─ Phase 3 ─[reassess]─ P3/P4 on demand
 ├─ Phase 4 (azuread, parallel)
 └─ Phase 5 (migration, last / lowest priority)
```

PR discipline: one service (or one partial-gap service) per PR. Migration PRs never
mix with feature PRs (small, bisectable reviews).
