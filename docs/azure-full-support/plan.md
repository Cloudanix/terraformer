# Plan: Full Azure Service & Resource Coverage for Terraformer

Goal: import every Azure resource that is **importable** — i.e. every resource the
`terraform-provider-azurerm` plugin supports **and** that Azure exposes a List/Get
API for. Three work streams (one more than AWS — Azure carries an SDK debt AWS doesn't):

0. **SDK upgrade** — migrate the Azure provider off the **deprecated Track 1**
   monolithic SDK (`azure-sdk-for-go v63.4.0+incompatible`) onto the modular
   **Track 2** (`sdk/resourcemanager/*` + `azidentity`/`azcore`). This is a
   prerequisite, not optional: Track 1 is frozen and new services only ship Track 2.
1. **Missing services** — add a generator for each service not in the registry.
2. **Missing resources** — fill gaps inside services already registered.

> This mirrors `docs/aws-full-support/plan.md`. The structural difference: AWS was
> already on `aws-sdk-go-v2`, so AWS had no SDK stream. Azure's current SDK is the
> legacy track and the user explicitly asked to "update the azure sdk … upgrade it
> to latest version." That upgrade is **Phase 0** and everything fans out behind it.

---

## 0. Current state snapshot (facts, measured 2026-06-23)

Provider lives in **`providers/azure/`** (registry key prefix `azure`, not `azurerm`).

- **35** service keys registered in `AzureProvider.GetSupportedService()`
  (`providers/azure/azure_provider.go:350-388`).
- **141** distinct `azurerm_*` resource types emitted across `providers/azure/*.go`.
- **SDK: Track 1**, `github.com/Azure/azure-sdk-for-go v63.4.0+incompatible` —
  imports of the form `services/<svc>/mgmt/<api-version>/<svc>` (e.g.
  `services/network/mgmt/2019-08-01/network`). **Deprecated & frozen by Microsoft.**
- **Auth: go-autorest**, `github.com/Azure/go-autorest/autorest` +
  `github.com/hashicorp/go-azure-helpers/authentication` +
  `github.com/manicminer/hamilton/environments`. Returns an `autorest.Authorizer`
  injected via `Service.Args["authorizer"]`.
- **go.mod**: `go 1.24.0`.
- Cobra command `cmd/provider_cmd_azure.go` wires `--resource-group`/`-R` +
  base flags. Scope is **subscription + resource-group**, not region.
- `docs/azure.md` exists (230 lines) and lists all 141 resources by service.

### Registered services (35)

`resource_group`, `analysis`, `app_service`, `application_gateway`, `cosmosdb`,
`container`, `database`, `databricks`, `data_factory`, `disk`, `dns`, `eventhub`,
`keyvault`, `load_balancer`, `management_lock`, `network_interface`,
`network_security_group`, `network_watcher`, `private_dns`, `private_endpoint`,
`public_ip`, `purview`, `redis`, `route_table`, `scaleset`,
`security_center_contact`, `security_center_subscription_pricing`,
`ssh_public_key`, `storage_account`, `storage_blob`, `storage_container`,
`synapse`, `subnet`, `virtual_machine`, `virtual_network`.

The authoritative current set is reproduced by the script in §3 — keep it in
`docs/azure-full-support/current-coverage.txt` so diffs are reproducible.

---

## 1. Hard constraints (read first — they bound "full coverage")

Terraformer is **not** an Azure-API dumper. It is bounded on three sides:

- **Terraform provider bound.** Terraformer emits HCL + tfstate, then refreshes
  each resource through the real `terraform-provider-azurerm` plugin
  (`terraformutils/providerwrapper`). If the provider has **no resource** for a
  thing, terraformer **cannot** import it. The provider's resource set
  (~1100 `azurerm_*` resources in v4.x), **not** an Azure REST catalog, is the
  source of truth.
- **List-API bound.** Each generator's `InitResources()` must *enumerate* existing
  resources via a `List*`/`ListByResourceGroup*` pager. A TF resource with no
  list API (some singleton/sub-resource configs) cannot be auto-discovered.
- **Resource-group bound.** Azure discovery is scoped to the resource group(s)
  passed via `-R`. Subscription-wide resources (policy, management groups, some
  AAD) need an explicit subscription-scoped list call, not the RG-scoped one.

**Excluded from scope** (no importable infrastructure — do not build generators):

- Data-plane / runtime APIs (blob data ops, queue messages, event data).
- Read-only / advisory services with no `azurerm_*` resource (Advisor
  recommendations, Resource Graph queries, Cost Management *reads*).
- Azure AD / Entra ID objects — those live in a **separate** provider
  (`azuread`, not `azurerm`); out of scope for this provider unless we add an
  `azuread` provider separately (decide in §11).

> Any service marked "verify" below must be confirmed against the provider schema
> (§3) before building — best-effort from domain knowledge.

---

## 2. Phase 0 — SDK upgrade (the prerequisite, do this FIRST)

Track 1 (`azure-sdk-for-go/services/.../mgmt/...`) is in maintenance mode and
receives no new services. Track 2 is the modular set of
`github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/<svc>/arm<svc>` modules with
`azcore` (pipeline/pager) + `azidentity` (auth). New generators MUST be Track 2;
existing ones get migrated incrementally.

### 2a. Foundations to build FIRST (before any generator)

These shared pieces every Track 2 generator depends on — build + test them before
fanning out, or the boilerplate gets copy-pasted 35×+ and a bug lands everywhere.

- **`azidentity` credential in `azure_provider.go`** — replace the go-autorest
  `authentication.Builder`/`autorest.Authorizer` flow with
  `azidentity.NewDefaultAzureCredential` (honors `AZURE_*` env vars, MSI, CLI,
  workload identity, OIDC). Inject the `azcore.TokenCredential` + `*arm.ClientOptions`
  via `Service.Args` exactly where `authorizer` is injected today. **Keep both
  paths alive during migration** so half-migrated builds compile and run.
- **`AzureService.getClientOptions()` helper** (`azure_service.go`) — returns
  `(subscriptionID string, cred azcore.TokenCredential, opts *arm.ClientOptions)`,
  the Track 2 analogue of today's `getClientArgs()`. Sovereign-cloud endpoint
  handled via `arm.ClientOptions.Cloud` (replaces `CustomResourceManagerEndpoint`).
- **`appendSimpleResources[T]` generic helper** — factors the
  pager→item→`NewSimpleResource` loop so simple services are ~5 lines. Nil-safe
  (Track 2 still returns `*string` IDs; use a `valueOrEmpty` helper, never `*p.ID`),
  skips empty import IDs. Nested/composite services stay hand-written.
- **`defaultAllowEmptyValues = []string{"tags."}`** package var — Azure tags map
  the same way; replaces per-file empty slices.
- **Resource-group iteration helper** — `-R name1:name2` already splits multiple
  RGs upstream; centralize the "list across the requested RGs" loop so every
  generator doesn't re-implement it.

### 2b. Migration order (incremental, compile-green at every step)

Do NOT big-bang all 35 services. Order:

1. Land §2a foundations + the dual auth path (both Track 1 authorizer and Track 2
   credential available in `Args`). Build stays green; nothing migrated yet.
2. Migrate the **leaf / simple** services first (`resource_group`, `disk`,
   `public_ip`, `ssh_public_key`, `redis`, `management_lock`) — smallest blast
   radius, proves the pattern.
3. Migrate the **networking cluster** (`virtual_network`, `subnet`, `nsg`,
   `network_interface`, `route_table`, `load_balancer`, `public_ip`,
   `private_endpoint`, `private_dns`, `dns`) — shares the `armnetwork` module, so
   one module bump covers many files.
4. Migrate **compute/storage/db** (`virtual_machine`, `scaleset`, `storage_*`,
   `database`, `cosmosdb`, `eventhub`, `keyvault`).
5. Migrate the **long tail** (`data_factory` is the big one — ~40 resource types,
   `analysis`, `synapse`, `databricks`, `purview`, `app_service`,
   `application_gateway`, `container`, `security_center_*`, `network_watcher`).
6. Remove the Track 1 authorizer path + drop `azure-sdk-for-go v63.4.0` and
   `go-autorest`/`hamilton` deps from go.mod once the last file is migrated. Run
   `go mod tidy`. **Only now is Phase 0 done.**

Each migrated service = one PR. After each, `go build` + `terraform plan` round-trip
(§8) must show no diff vs the pre-migration output — the SDK swap must be behavior-
preserving. Pin `terraform-provider-azurerm` (§3) so the refresh target is stable
across the whole migration.

> Map for which Track 2 module replaces each Track 1 import: Microsoft's
> `azure-sdk-for-go/sdk/resourcemanager/<area>/arm<area>` — e.g.
> `services/network/mgmt/.../network` → `sdk/resourcemanager/network/armnetwork/v5`,
> `services/compute/.../compute` → `sdk/resourcemanager/compute/armcompute/v6`,
> `services/storage/.../storage` → `sdk/resourcemanager/storage/armstorage`. Record
> the exact module+major-version per service in a migration table as you go.

---

## 3. Establish the authoritative gap list (do this once, after Phase 0 foundations)

Hand-maintaining a 1100-row list is error-prone. Generate it.

**a. Dump every resource terraform-provider-azurerm supports** (source of truth):

```bash
# requires terraform + the azurerm provider installed in a scratch dir
# PIN the version — do NOT use -upgrade. Terraformer refreshes against whatever
# terraform-provider-azurerm the USER installed at runtime. If the gap list is
# computed against a newer provider than the user runs, you build generators for
# resources that fail at refresh. Pin to the declared supported floor and document it.
AZURERM_PROVIDER_VERSION="4.0.0"   # <- the supported floor; bump deliberately
cd $(mktemp -d)
cat > main.tf <<EOF
terraform {
  required_providers { azurerm = { source = "hashicorp/azurerm", version = "= ${AZURERM_PROVIDER_VERSION}" } }
}
EOF
terraform init >/dev/null
terraform providers schema -json \
  | jq -r '.provider_schemas[] | .resource_schemas | keys[]' \
  | sort > /tmp/tf-azurerm-all-resources.txt
wc -l /tmp/tf-azurerm-all-resources.txt   # ~1100
```

Record the pinned version in `docs/azure.md` as the supported provider floor so the
gap list and the user's runtime provider stay aligned.

**b. Dump what terraformer currently emits:**

```bash
cd providers/azure
grep -rhoE '"azurerm_[a-z0-9_]+"' *.go | tr -d '"' | sort -u \
  > docs/azure-full-support/current-coverage.txt   # 141 lines today
```

**c. The gap = (a) − (b):**

```bash
comm -23 /tmp/tf-azurerm-all-resources.txt \
         docs/azure-full-support/current-coverage.txt \
  > docs/azure-full-support/missing-resources.txt
```

Then bucket `missing-resources.txt` by service prefix → that bucketing drives both
work streams. A prefix entirely absent from the registry = **new service**; a prefix
partially present = **resource gap** in an existing service.

> Note: TF resource prefix ≠ terraformer registry key. Azure prefixes are usually
> the service name (`azurerm_kubernetes_cluster` → new service `kubernetes`/`aks`;
> `azurerm_monitor_*` → new service `monitor`). Map prefixes to registry keys when
> bucketing.

---

## 4. Gap inventory (best-effort; confirm via §3)

### 4a. Missing services that HAVE terraform-provider-azurerm resources (build these)

Grouped by priority. Each becomes one `providers/azure/<name>.go` + a registry entry.
The current 35 services are heavily networking/storage/db weighted; the biggest
absences are container/compute orchestration, monitoring, identity, and the entire
PaaS/AI surface.

**P1 — common infra, high demand**

| Service (key) | TF resource prefix | Notes |
|---|---|---|
| `kubernetes` (AKS) | `azurerm_kubernetes_cluster*` | cluster, node_pool — top ask, currently absent |
| `monitor` | `azurerm_monitor_*` | diagnostic_setting, action_group, metric/activity_log alerts, autoscale |
| `log_analytics` | `azurerm_log_analytics_*` | workspace, solution, saved_search, data sources |
| `application_insights` | `azurerm_application_insights*` | + api_key, smart_detection |
| `role_assignment` (RBAC) | `azurerm_role_assignment`, `azurerm_role_definition` | subscription-scoped |
| `policy` | `azurerm_policy_definition`, `_set_definition`, `_assignment` | governance |
| `managed_identity` | `azurerm_user_assigned_identity` | referenced everywhere |
| `function_app` | `azurerm_function_app*`, `azurerm_*_function_app` | linux/windows function apps |
| `app_service` (gaps) | `azurerm_linux_web_app`, `_windows_web_app`, `_service_plan` | existing covers legacy `azurerm_app_service` only |
| `recovery_services` | `azurerm_recovery_services_vault`, backup_* policies/protected_* | backup |
| `automation` | `azurerm_automation_*` | account, runbook, schedule |
| `cdn` / `frontdoor` | `azurerm_cdn_*`, `azurerm_cdn_frontdoor_*` | |
| `traffic_manager` | `azurerm_traffic_manager_*` | profile, endpoints |
| `nat_gateway` | `azurerm_nat_gateway` | |
| `virtual_wan` | `azurerm_virtual_wan`, `_hub`, `_gateway` | |
| `firewall` | `azurerm_firewall`, `_policy`, `_application/network/nat_rule_collection` | |

**P2 — security / governance / observability**

| Service | TF prefix | Notes |
|---|---|---|
| `key_vault` (gaps) | `azurerm_key_vault_key/secret/certificate/access_policy` | existing covers vault only |
| `security_center` (gaps) | `azurerm_security_center_*` (assessment, automation, setting, workspace) | existing covers contact + pricing only |
| `sentinel` | `azurerm_sentinel_*` | alert rules, data connectors |
| `ddos` | `azurerm_network_ddos_protection_plan` | |
| `bastion` | `azurerm_bastion_host` | |
| `private_link` (gaps) | `azurerm_private_dns_resolver_*` | newer DNS resolver |
| `lighthouse` | `azurerm_lighthouse_*` | delegations |
| `management_group` | `azurerm_management_group*` | subscription-scoped |
| `advisor` | `azurerm_advisor_*` | (verify — mostly read) |

**P3 — data / analytics / AI / app platform**

| Service | TF prefix | Notes |
|---|---|---|
| `servicebus` | `azurerm_servicebus_*` | namespace, queue, topic, subscription |
| `signalr` | `azurerm_signalr_service*` | |
| `apim` | `azurerm_api_management*` | large — API Management surface |
| `search` | `azurerm_search_service` | |
| `cognitive` | `azurerm_cognitive_account`, `_deployment` | Azure OpenAI / AI services |
| `machine_learning` | `azurerm_machine_learning_*` | workspace, compute |
| `synapse` (gaps) | `azurerm_synapse_*` (more pool/linked types) | existing partial |
| `hdinsight` | `azurerm_hdinsight_*_cluster` | |
| `data_lake` | `azurerm_data_lake_*` | store, analytics |
| `kusto` (Data Explorer) | `azurerm_kusto_*` | cluster, database |
| `stream_analytics` | `azurerm_stream_analytics_*` | |
| `iothub` | `azurerm_iothub*` | hub, dps, routes |
| `digital_twins` | `azurerm_digital_twins_*` | |
| `eventgrid` | `azurerm_eventgrid_*` | topic, subscription, domain |
| `mssql` (gaps) | `azurerm_mssql_managed_instance`, `_elasticpool`, `_job_agent` | existing covers server/db/fw only |
| `postgresql_flexible` | `azurerm_postgresql_flexible_server*` | newer flexible tier |
| `mysql_flexible` | `azurerm_mysql_flexible_server*` | newer flexible tier |
| `container_app` | `azurerm_container_app*` | environment, app — modern container PaaS |

**P4 — long tail (build if needed)**

`spring_cloud`, `batch`, `media_services`, `maps`, `relay`, `notification_hub`,
`spatial_anchors`, `mixed_reality`, `lab_services` (`azurerm_dev_test_*`),
`dashboard` (Grafana — `azurerm_dashboard_grafana`), `managed_lustre`,
`elastic_san`, `netapp`, `storage_sync`, `data_share`, `data_protection`,
`fluid_relay`, `web_pubsub`, `load_test`, `chaos_studio`, `dev_center`,
`orbital`, `confidential_ledger`, `automanage`, `workloads` (SAP),
`communication` (`azurerm_communication_service`), `healthcare`, `vmware`
(AVS — `azurerm_vmware_*`), `dedicated_host`, `capacity_reservation`,
`proximity_placement_group`, `availability_set`, `image`, `shared_image_*`
(`azurerm_shared_image_gallery`/`_version`), `dev_test_lab`.

### 4b. Partial services — known resource gaps (fill these)

| Registry key | Has | Missing (examples) |
|---|---|---|
| `key_vault` | vault | `azurerm_key_vault_key`, `_secret`, `_certificate`, `_access_policy`, `_managed_storage_account` |
| `app_service` | `azurerm_app_service` (legacy) | `azurerm_service_plan`, `_linux_web_app`, `_windows_web_app`, `_linux_function_app`, `_windows_function_app`, slots, custom_hostname_binding |
| `security_center_*` | contact, subscription_pricing | `azurerm_security_center_setting`, `_workspace`, `_assessment`, `_automation`, `_auto_provisioning` |
| `cosmosdb` | account, sql_database/container, table | `azurerm_cosmosdb_mongo_database/collection`, `_cassandra_*`, `_gremlin_*`, `_sql_function/trigger/stored_procedure` |
| `eventhub` | namespace, hub, consumer_group, ns_auth_rule | `azurerm_eventhub_authorization_rule`, `_namespace_disaster_recovery_config`, `_cluster` |
| `database` (sql/mysql/pg/mariadb) | server, db, fw, config, vnet_rule | flexible-server tiers (`azurerm_mysql_flexible_server`, `azurerm_postgresql_flexible_server*`), `azurerm_mssql_managed_instance`, `_job_agent`, `_server_extended_auditing_policy` |
| `storage_account` | account, blob, container | `azurerm_storage_queue`, `_table`, `_share`, `_share_file/directory`, `_management_policy`, `_account_network_rules`, `_data_lake_gen2_filesystem`, `_object_replication`, `_encryption_scope` |
| `network_security_group` | nsg, rule | (mostly complete — verify `azurerm_network_interface_security_group_association`) |
| `virtual_network` | vnet | `azurerm_virtual_network_peering`, `_virtual_network_gateway`, `_virtual_network_gateway_connection`, `_local_network_gateway`, `_vpn_gateway*` |
| `load_balancer` | lb, backend_pool, nat_rule, probe | `azurerm_lb_rule`, `_outbound_rule`, `_nat_pool`, `_backend_address_pool_address` |
| `application_gateway` | app_gateway | (verify WAF policy split `azurerm_web_application_firewall_policy`) |
| `container` | container_group, registry(+webhook) | `azurerm_container_registry_token`, `_scope_map`, `_task`, `_replication`; AKS → new `kubernetes` service |
| `synapse` | workspace, pools, fw, pep, plh | `azurerm_synapse_role_assignment`, `_linked_service`, `_integration_runtime_*`, `_managed_private_endpoint` (verify) |
| `data_factory` | ~40 dataset/linked/trigger/pipeline types | `azurerm_data_factory_managed_private_endpoint`, `_credential_*`, newer linked-service types added in provider v4 |
| `redis` | cache | `azurerm_redis_firewall_rule`, `_linked_server`, `_enterprise_cluster/database` |
| `network_watcher` | watcher, flow_log, packet_capture | `azurerm_network_connection_monitor`, `_network_watcher_flow_log` (verify v4 split) |

> The §3 diff produces the *complete* per-service gap; the table above is the
> high-value starter set, not exhaustive.

---

## 5. Recipe — add a NEW service generator (Track 2)

Reference: a migrated simple service (e.g. `disk.go` post-migration) for the simple
case, `data_factory.go` for multi-resource. Steps:

1. **Create `providers/azure/<service>.go`** using the §2a helper — do NOT hand-roll
   the pager loop:

```go
package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/<area>/arm<area>"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type <Svc>Generator struct {
	AzureService
}

func (g *<Svc>Generator) InitResources() error {
	subscriptionID, cred, opts := g.getClientOptions()   // Track 2 (see §2a)
	client, err := arm<area>.New<Thing>sClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}

	pager := client.NewListByResourceGroupPager(g.resourceGroup(), nil)
	for pager.More() {
		page, err := pager.NextPage(context.TODO())   // ponytail: TODO ctx repo-wide
		if err != nil {
			return err
		}
		for _, item := range page.Value {
			id := valueOrEmpty(item.ID)        // nil-safe: never *item.ID
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				valueOrEmpty(item.Name),
				"azurerm_<tf_type>",
				"azurerm",
				defaultAllowEmptyValues))
		}
	}
	return nil
}
```

> The §2a `appendSimpleResources[T]` helper collapses the inner two loops to one
> call for the common case; use it. Keep the explicit loop only when items need
> per-item branching (extra API calls, conditional resource types).

2. **Register it** in `azure_provider.go` → `GetSupportedService()`:

```go
"<service>": &<Svc>Generator{},
```

   **Key naming convention:** new keys = the snake_case Azure service name users
   expect (`kubernetes`, `monitor`, `log_analytics`, `role_assignment`). Match the
   style of existing keys; legacy keys stay for back-compat — document, don't churn.

3. **Scope** — most Azure resources are resource-group-scoped (`ListByResourceGroup`).
   Subscription-scoped resources (`role_assignment`, `policy_*`, `management_group`,
   `subscription`-level diagnostic settings) use the subscription-scoped `List` and
   must ignore `-R`. Mark which in the generator and document it (Azure has no
   region-pass machinery like AWS, so this is the only scoping decision).

4. **`PostConvertHook()`** (optional): strip computed-only fields, fix circular
   references (see existing `subnet.go`/`virtual_network.go` which strip inlined
   subnets), or wrap policy/JSON attributes.

5. **Docs**: append the service key + resources to `docs/azure.md`.

6. **Tests** (per `sg_test.go` convention in AWS, mirror for Azure): if
   `InitResources` does non-trivial result processing, extract it into a pure
   function taking the already-fetched SDK structs and unit-test that. A trivial
   list→`NewSimpleResource` mapper needs no test. Do **not** mock the SDK client.

**Per-resource gotchas:**

- **Never `*item.X`** — Track 2 SDK fields are pointers; use a `valueOrEmpty`/
  `to.Value` helper. A nil optional field deref panics and kills the whole run.
- The **import ID** (1st arg) is the full ARM resource ID
  (`/subscriptions/.../resourceGroups/.../providers/Microsoft.X/...`) — check the
  provider docs' "Import" section per resource; some composite IDs differ.
- Sub-resources needing a parent (e.g. blob container needs storage account) require
  the parent loop to nest the child enumeration — see `storage_*` files.
- Composite/parent-scoped import IDs use `terraformutils.NewResource(...)` with
  explicit attributes instead of `NewSimpleResource`.

## 6. Recipe — expand an EXISTING service (resource gaps)

1. From §3's per-service gap, list missing `azurerm_*` types for that file.
2. Add an enumeration block per missing type inside the existing `InitResources()`
   (or split into a helper). Reuse the already-built Track 2 client when same arm
   module; create a new client otherwise.
3. Add to that service's `PostConvertHook()` if needed.
4. Update `docs/azure.md`.

---

## 7. Phased rollout

- **Phase 0a — SDK foundations (§2a):** `azidentity` credential + dual auth path,
  `getClientOptions()`, `appendSimpleResources`, `defaultAllowEmptyValues`,
  RG-iteration helper, `valueOrEmpty`. Build stays green; **nothing fans out before
  this lands** — it stops the Track 2 boilerplate (and bugs) being copied 35×+.
- **Phase 0b — SDK migration (§2b):** migrate the 35 existing services to Track 2,
  leaf-first then by shared arm module, each a behavior-preserving PR with a
  no-diff `terraform plan` check. Drop Track 1 + go-autorest from go.mod at the end.
- **Phase 0c — tooling (1 day):** run §3 (pinned azurerm version), commit
  `current-coverage.txt`, `tf-azurerm-all-resources.txt`, `missing-resources.txt`;
  bucket by service. Produces the *real* backlog.
- **Phase 1 — partial-service gaps (§4b):** highest ROI, services already wired.
  Start with `key_vault`, `app_service`, `storage_account`, `cosmosdb`,
  `virtual_network`, `database` (flexible servers), `security_center_*`.
- **Phase 2 — P1 missing services (§4a):** kubernetes, monitor, log_analytics,
  application_insights, role_assignment, policy, managed_identity, function_app,
  recovery_services, automation, cdn/frontdoor, traffic_manager, firewall,
  nat_gateway, virtual_wan.
- **Phase 3 — P2 security/governance:** key_vault secrets, sentinel, ddos, bastion,
  management_group, lighthouse, private_dns_resolver.
- **Phase 4 — P3 data/AI/app:** servicebus, apim, signalr, search, cognitive,
  machine_learning, kusto, stream_analytics, iothub, eventgrid, container_app,
  hdinsight, flexible-server DB tiers.
- **Phase 5 — P4 long tail:** build on demand.

Each service = one focused PR (generator + registry + docs). SDK migration PRs stay
separate from new-feature PRs so reviews are small and bisectable.

---

## 8. Validation per service

- **Unit (required):** the §2a shared helpers — `appendSimpleResources` (nil ID
  skipped, empty page, mapping), `getClientOptions` wiring, `valueOrEmpty`. These
  are high-fan-in; a bug here is a bug in all generators.
- **Unit (per generator, conditional):** only if `InitResources` has non-trivial
  result processing — extract to a pure func and test it. Trivial mappers need none.
  No SDK mocking (matches codebase).
- **Integration (the correctness bar):**
  `go build -o terraformer && ./terraformer import azure -R <rg> --resources=<svc>`
  against a test subscription, then `terraform plan` on the output shows **no diff**
  (refresh round-trips cleanly). Use the provider version pinned in §3.
  - For **migration** PRs specifically: diff the generated `.tf`/`.tfstate` before
    vs after the Track 2 swap — must be identical (behavior-preserving).
- `go test ./providers/azure/...` (full suite pulls the provider plugin via
  `providerwrapper`; needs module cache + network).
- `golangci-lint run` (gocritic/revive/unconvert/unparam enabled).

---

## 9. Effort estimate

- SDK migration (§2b): **35 existing services** migrated to Track 2 — the bulk of
  Phase 0 effort; `data_factory` (~40 resource types) is the single largest file.
- Partial-service gaps (§4b): ~15 services touched, ~100 resource blocks.
- New services with TF resources (§4a): ~60 services (P1–P4).
- Total new `azurerm_*` resource types reachable: the §3 diff gives the exact count
  (expect the gap to be **700–900** of the provider's ~1100, since terraformer
  currently covers 141 and some provider resources have no list API).

Track progress by re-running §3 after each PR and shrinking
`missing-resources.txt`. Done = diff is empty modulo the documented "no list API"
exclusions, and zero Track 1 imports remain (`grep -r "azure-sdk-for-go/services"
providers/azure` is empty).

---

## 10. Performance & scale notes

- Service init runs **sequentially** (same engine as AWS). A full
  `--resources=*` import across many resource groups = Σ(every service's serial
  paginated calls) × every RG. Recommend users scope `--resources` and `-R`.
- Track 2 `azcore` pipeline supports built-in **retry policy** — configure
  `arm.ClientOptions.Retry` (max attempts + backoff) in `getClientOptions()` so
  throttled large imports degrade instead of aborting (the Track 2 analogue of the
  AWS SDK retryer).
- Azure has **no region-pass complexity** (resources are RG-scoped, region is an
  attribute, not a discovery axis) — simpler than the AWS 3-pass model. The only
  scoping fork is RG-scoped vs subscription-scoped lists (§5 step 3).
- Deferred (file in a `TODOS.md` if pursued): parallel service execution,
  per-service memory streaming, `context.TODO()` timeout sweep — same deferrals as
  the AWS plan.

## 11. Open decisions

- **Track 2 major versions:** `armnetwork`, `armcompute`, etc. are independently
  versioned (`/v5`, `/v6`). Pin each in go.mod and record in the migration table;
  bump deliberately, not via `go get -u`.
- **Azure AD / Entra objects:** users sometimes want `azuread_*` (users, groups,
  app registrations, service principals). That's a **separate Terraform provider**
  and would be a **new terraformer provider** (`providers/azuread/`), not part of
  `azurerm`. Decide whether it's in scope before promising "full Azure coverage."
- **Subscription-scoped resources:** `role_assignment`, `policy_*`,
  `management_group` ignore `-R`. Confirm the CLI UX — do we list across the whole
  subscription, or require an explicit flag? (Default: list subscription-wide,
  documented.)
- **Un-listable resources:** maintain `docs/azure-full-support/no-list-api.md` so
  "missing" never silently means "impossible".
- **Sovereign clouds:** Track 2 handles Gov/China via `arm.ClientOptions.Cloud`;
  make sure the migration preserves the existing `CustomResourceManagerEndpoint`
  behavior for users on non-public clouds.
