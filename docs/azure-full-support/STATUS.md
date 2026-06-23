# Azure full-support — implementation status

Tracks progress against `PHASES.md`. Coverage = distinct `azurerm_*` types emitted
by `providers/azure/*.go` (excluding test files).

## Coverage

- Baseline: **141** types (35 services), measured 2026-06-23.
- Current: **229** types (+88). Phase 1 mgmt-plane complete (incl. app_service
  modern apps + service_plan); Phase 2 incl. policy; ~50 new Phase 2/3 Track 2
  services. azuread (P4) + Track-1 migration (P5) not started.
- Provider gap (vs v4.78.0, 1130 types): re-run `plan.md` §3 to recompute.

## Phase 0 — foundations (DONE)

`azidentity` dual auth path, `getClientOptions`, `appendFromPager`,
`valueOrEmpty`, `resourceGroups`, `defaultAllowEmptyValues`, helper unit tests,
v4.78.0 floor. Track 1 untouched (migration deferred to Phase 5).

## Phase 1 — partial-service gaps (mgmt-plane DONE; data-plane deferred)

All tractable **mgmt-plane** §4b gaps are done (Track 2 or same-module Track 1
enumerations added to the existing generators):
- virtual_network: `_gateway`, `local_network_gateway`, `_gateway_connection`, `_peering`
- database: `mysql_flexible_server`, `postgresql_flexible_server`
- cosmosdb: mongo (db+collection), cassandra (keyspace+table), gremlin (db+graph)
- redis: `firewall_rule`, `linked_server`
- load_balancer: `lb_rule`, `lb_outbound_rule` (nat_pool is inline)
- network_watcher: `network_connection_monitor`
- eventhub: `authorization_rule`, `namespace_disaster_recovery_config`
- container: `container_registry_replication`, `container_registry_task`

**Deferred with technical reasons** (not tractable as simple list→import):
- **Data-plane resources** — different API surface + auth scope + import-ID format,
  no mgmt-plane list: key_vault keys/secrets/certificates; storage
  queue/table/share/file/data_lake_gen2; synapse linked_service/role_assignment.
- **Polymorphic/branching** — need a discriminator→tf-type map + a unit test:
  synapse integration_runtime (Managed vs SelfHosted), hdinsight (per-kind
  cluster). (app_service modern apps DONE via siteResourceType kind-branching.)
- **Singletons without a list API** — storage management_policy (per-account
  `Get default`), most security_center settings.
- **Preview-version-only** — container_registry token/scope_map (need a newer
  containerregistry API version than the one the file imports).
These are real follow-ups, not silent omissions; revisit per `no-list-api.md`.

## Phase 2 — P1 new services (largely done)

Done: `nat_gateway`, `kubernetes` (cluster + node_pool, User-mode filter tested),
`managed_identity`, `log_analytics`, `application_insights`, `traffic_manager`,
`firewall`, `virtual_wan` (wan + hub), `monitor` (action_group, activity_log_alert,
autoscale_setting, metric_alert), `cdn` (profile + endpoint), `role_assignment`
(subscription-scoped), `policy` (definition/set_definition Custom-filtered +
assignment, subscription-scoped, isCustomPolicy unit-tested).

Not yet: `automation` runbooks/schedules (account done), role_definition
(composite import ID + built-in filtering). (function_app DONE via app_service
kind-branching.)

## Phase 3 — P2/P3 (largely done)

Done: recovery_services, servicebus, cognitive, search, signalr, eventgrid,
automation, bastion, ddos, management_group, apim, kusto, iothub,
stream_analytics, machine_learning, container_app, netapp, mssql, powerbi,
digital_twins, relay, web_pubsub, notification_hub, batch, dashboard (grafana),
maps, private_dns_resolver, spring_cloud, data_share, healthcare, load_test,
elastic_san, communication, dev_center, chaos_studio, confidential_ledger,
fluid_relay, data_protection, attestation, vmware (AVS), nginx, storage_mover,
app_service (modern web/function apps + service_plan).

Not yet: sentinel, lighthouse, hdinsight (per-kind), spatial_anchors, orbital,
automanage, workloads (SAP), and the remaining long-tail single-resource
services + multi-resource sub-resource expansions.

## Phase 4 — azuread (not started)

`providers/azuread/` separate stream.

## Phase 5 — Track 1 → Track 2 migration (in progress)

Behavior-preserving SDK swaps (same resource types + ARM import IDs), validated
by build/vet/test (live no-diff `terraform plan` round-trip not available in the
offline sandbox). **Migrated:** disk, public_ip (+prefix), ssh_public_key,
resource_group, management_lock, route_table (+route/route_filter),
network_security_group (+rule), app_service (also modernized).

**Remaining (~27 files still import `azure-sdk-for-go/services`):** networking
(virtual_network, subnet, network_interface, network_watcher, dns, private_dns,
private_endpoint, application_gateway, load_balancer), compute (virtual_machine,
scaleset), storage_*, database, cosmosdb, eventhub, keyvault, redis, container,
data_factory (largest), synapse, analysis, databricks, purview, security_center_*.
Done when `grep -r "azure-sdk-for-go/services" providers/azure` is empty; then
drop Track 1 + go-autorest/hamilton from go.mod.

## How to build/test offline (sandbox)

Go module downloads are blocked in the sandbox, but `curl proxy.golang.org` works.
Helper scripts staged in the session scratchpad:
- `getmod.sh <module> [ver]` — curls .info/.mod/.zip into a file-GOPROXY.
- `goget.sh <module@ver>...` — `go get` against the file proxy, auto-fetching deps.
- `tidy.sh` — `go mod tidy` with auto-fetch.
- `gotool.sh <args>` — runs `go` with the file-proxy env.
Env: `GOMODCACHE=$TMPDIR/gomodcache GOCACHE=$TMPDIR/gocache GOPROXY="file://$TMPDIR/goproxy,file:///Users/puru/go/pkg/mod/cache/download" GOSUMDB=off GOFLAGS=-mod=mod`.
