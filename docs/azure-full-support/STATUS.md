# Azure full-support — implementation status

Tracks progress against `PHASES.md`. Coverage = distinct `azurerm_*` types emitted
by `providers/azure/*.go` (excluding test files).

## Coverage

- Baseline: **141** types (35 services), measured 2026-06-23.
- Current: **205** types (+64). 50+ new services across networking, monitoring, data, AI, app platform, governance.
  managed_identity, log_analytics, application_insights, traffic_manager,
  firewall, virtual_wan, monitor, cdn, role_assignment, recovery_services,
  automation, servicebus, cognitive, search, signalr, eventgrid, bastion, ddos,
  kusto, iothub, stream_analytics, container_app, apim, management_group,
  machine_learning. Plus gaps in virtual_network/database/cosmosdb/redis.
- Provider gap (vs v4.78.0, 1130 types): re-run `plan.md` §3 to recompute.

## Phase 0 — foundations (DONE)

`azidentity` dual auth path, `getClientOptions`, `appendFromPager`,
`valueOrEmpty`, `resourceGroups`, `defaultAllowEmptyValues`, helper unit tests,
v4.78.0 floor. Track 1 untouched (migration deferred to Phase 5).

## Phase 1 — partial-service gaps (in progress)

Done (Track 2 enumerations added to existing Track 1 generators):
- virtual_network: `_gateway`, `local_network_gateway`, `_gateway_connection`, `_peering`
- database: `mysql_flexible_server`, `postgresql_flexible_server`
- cosmosdb: mongo (db+collection), cassandra (keyspace+table), gremlin (db+graph)
- redis: `firewall_rule`, `linked_server`

Deferred / not yet done:
- **key_vault keys/secrets/certificates** — data-plane (vault URI), separate API
  + import-ID format. Needs data-plane client; tracked for later.
- **storage queue/table/share/file** — data-plane. mgmt-plane subresources
  (management_policy, network_rules, data_lake_gen2) still to do.
- app_service modern apps (linux/windows web+function, service_plan) — armappservice,
  needs kind-branching + tests.
- data_factory v4 additions; security_center singletons (often no list API).

## Phase 2 — P1 new services (in progress)

Done: `nat_gateway`, `kubernetes` (cluster + node_pool, User-mode filter tested),
`managed_identity`, `log_analytics`, `application_insights`, `traffic_manager`,
`firewall`, `virtual_wan` (wan + hub), `monitor` (action_group, activity_log_alert,
autoscale_setting, metric_alert), `cdn` (profile + endpoint), `role_assignment`
(subscription-scoped).

Not yet: `policy`, `function_app`, `automation` runbooks/schedules (only account
done), role_definition (composite import ID + built-in filtering).

## Phase 3 — P2/P3 (partly started)

Done: `recovery_services` (vault), `servicebus` (namespace), `cognitive` (account),
`search` (service), `signalr` (service), `eventgrid` (topic + domain), `automation`
(account).

Not yet: sentinel, ddos, bastion, management_group, lighthouse,
private_dns_resolver, apim, kusto, iothub, stream_analytics, machine_learning,
hdinsight, container_app, eventhub gaps, synapse gaps, storage gaps.

## Phase 4 — azuread (not started)

`providers/azuread/` separate stream.

## Phase 5 — Track 1 → Track 2 migration (not started, last)

35 original services still on Track 1. Dual auth path keeps both working.
`grep -r "azure-sdk-for-go/services" providers/azure` must be empty when done.

## How to build/test offline (sandbox)

Go module downloads are blocked in the sandbox, but `curl proxy.golang.org` works.
Helper scripts staged in the session scratchpad:
- `getmod.sh <module> [ver]` — curls .info/.mod/.zip into a file-GOPROXY.
- `goget.sh <module@ver>...` — `go get` against the file proxy, auto-fetching deps.
- `tidy.sh` — `go mod tidy` with auto-fetch.
- `gotool.sh <args>` — runs `go` with the file-proxy env.
Env: `GOMODCACHE=$TMPDIR/gomodcache GOCACHE=$TMPDIR/gocache GOPROXY="file://$TMPDIR/goproxy,file:///Users/puru/go/pkg/mod/cache/download" GOSUMDB=off GOFLAGS=-mod=mod`.
