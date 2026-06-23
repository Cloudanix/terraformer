# Plan: Full GCP Service & Resource Coverage for Terraformer

Goal: import every GCP resource that is **importable** — i.e. every resource the
`terraform-provider-google` plugin supports **and** that GCP exposes a
List/AggregatedList API for. Same shape as the AWS effort
(`docs/aws-full-support/plan.md`), but adapted to GCP's two-SDK + code-generator
reality. The user's explicit asks, mapped to sections:

1. **Bump the GCP SDK to latest** → §7 (SDK upgrade), done **first** because it
   moves the provider/API floor everything else is measured against.
2. **Enumerate everything terraform-provider-google supports** → §3a.
3. **Compute the delta vs. what terraformer emits today** → §3.
4. **Add the new ones** → §5/§6 recipes, §8 phased rollout.
5. **Tests + validation + one-service-per-commit** → §8/§9.

> Not yet reviewed. Run the `plan-review` skill on this before fanning out, same
> as AWS did — the gap inventory in §4 is best-effort domain knowledge until §3
> generates the authoritative list.

---

## 0. Architecture you must know before touching anything

GCP's provider is **not** shaped like AWS's. There is no `AwsFacade`-per-resource
hand-roll for everything; there are **two distinct codepaths**, and getting this
wrong wastes the whole effort.

### 0a. Compute services are CODE-GENERATED

`providers/gcp/gcp_compute_code_generator/` is a standalone `main` that:

1. Reads the Compute discovery doc `compute-api.json` (currently from the
   `$GOPATH/src/google.golang.org/api/compute/v1/compute-api.json` path — **a
   hack**, see §0c).
2. Iterates every Compute API resource that (a) appears in the
   `terraformResources` map in `resources.go` **and** (b) has a `list` method.
3. Emits one `providers/gcp/<resource>_gen.go` per resource + regenerates
   `providers/gcp/compute.go` (the `ComputeServices` registry map).

**Consequence:** to add a compute resource you add an entry to `resources.go`
and re-run the generator. You do **not** hand-write `*_gen.go` — they carry
`// AUTO-GENERATED CODE. DO NOT EDIT.` and will be clobbered. This is a massive
advantage over AWS: ~50 compute services came from one map.

The generator templates handle the three GCP locality shapes automatically:
- **global** — `List(project)` (e.g. `networks`, `globalAddresses`).
- **regional** — `List(project, region)` via `g.GetArgs()["region"].(compute.Region).Name`.
- **zonal** — loops `region.Zones`, calls `List(project, zone)`, and (optionally)
  prefixes the import ID with `zone/` (`ifIDWithZone`).

`basicGCPResource` covers the common case; `backendServices`,
`globalForwardingRules`, `instanceGroupManagers` show the override hooks
(`getAdditionalFields`, `ifNeedZone`, `ifIDWithZone`, `allowEmptyValues`).

### 0b. Everything else is HAND-WRITTEN

The non-compute services (`bigQuery`, `cloudFunctions`, `cloudsql`, `cloudtasks`,
`dataProc`, `dns`, `gcs`, `gke`, `iam`, `kms`, `logging`, `memoryStore`,
`monitoring`, `project`, `instances`, `pubsub`, `schedulerJobs`, `cloudbuild`)
are one `providers/gcp/<svc>.go` each, implementing `InitResources()` and
registered **by hand** in `gcp_provider.go` → `GetSupportedService()`. `instances`
is hand-written (not codegen) because of disk/refresh quirks — see the commented
block in `resources.go`.

These pull from **two SDK families** (this matters for the version bump, §7):
- **Discovery clients** `google.golang.org/api/<svc>/<ver>`: compute, bigquery,
  cloudfunctions/v2, cloudkms/v1, cloudresourcemanager/v1, dns/v1, dataproc/v1,
  pubsub/v1, redis/v1, storage/v1, sqladmin/v1beta4, cloudscheduler/v1beta1,
  container/v1beta1.
- **GAPIC clients** `cloud.google.com/go/<svc>/...`: logging/logadmin,
  iam/admin/apiv1, cloudbuild/apiv1, cloudtasks/apiv2, monitoring/apiv3.

New hand-written services should prefer the **discovery client**
(`google.golang.org/api/<svc>`) for consistency with the majority and because
the compute codegen + most services already use it; reach for a GAPIC client
only when the discovery client lacks the List method.

### 0c. The codegen `compute-api.json` path hack

`main.go` reads the discovery JSON from `$GOPATH/src/google.golang.org/api/...`,
which is the **old GOPATH layout**, not the Go-modules cache. Before the SDK bump
can regenerate compute correctly, fix this to read from the module cache, e.g.:

```
COMPUTE_JSON=$(go list -m -f '{{.Dir}}' google.golang.org/api)/compute/v1/compute-api.json
```

and pass it via flag/env instead of the hardcoded GOPATH path. Do this in
Phase 0 (foundations) — it's a one-file fix and it unblocks every compute
addition. (Module-cache files are read-only; the generator only reads it, so
that's fine.)

### 0d. Provider name / beta handling

`GCPFacade.PostConvertHook` rewrites `provider` to `google-beta` when
`--provider-type beta` is set; `Init` stores `providerType`. Any new generator
gets this for free **as long as it's wrapped in `GCPFacade`** in the registry.
Always register via `&GCPFacade{service: &XxxGenerator{}}`, never the bare
generator.

### 0e. Region/zone args contract

Services read locality from `g.GetArgs()`:
- `g.GetArgs()["project"].(string)`
- `g.GetArgs()["region"].(compute.Region)` — has `.Name` and `.Zones`
  (zonal loops iterate `.Zones`). For `--regions=global`, `getRegion` returns an
  empty `compute.Region{}`.

The cmd (`cmd/provider_cmd_google.go`) loops **projects × regions** and rewrites
the output path to `{provider}/<project>/{service}/<region>`. There is no
global/regional/east-only 3-pass split like AWS — each region is its own
top-level invocation pass. A global resource imported under `--regions=global`
just gets an empty region. Keep this contract; don't introduce AWS-style passes.

---

## 1. Hard constraints (read first — they bound "full coverage")

Same two bounds as AWS:

- **Terraform provider bound.** Terraformer emits HCL + tfstate then refreshes
  through the real `terraform-provider-google` plugin. If the provider has **no
  `google_*` resource** for a thing, terraformer cannot import it. The provider's
  resource set (**~1000+** `google_*` resources as of provider v7.x), not GCP's
  full API surface, is the source of truth. `terraform-provider-google-beta`
  adds more — reachable via `--provider-type beta`.
- **List-API bound.** Each generator must *enumerate* live resources via a
  `List`/`AggregatedList` call. A TF resource with no list API (singletons,
  project-level settings reachable only by Get with a known ID, IAM policy
  bindings discovered only by walking parents) cannot be auto-discovered without
  a parent walk.

**Excluded from scope** (no list-able importable infra — do not build generators):

- Pure data-plane / runtime APIs (BigQuery query jobs, Pub/Sub message contents,
  Logging entries themselves).
- Read/IAM-introspection-only surfaces with no standalone `google_*` resource.
- Project-singleton config that has no List API and is better left to the user.

> Anything marked "verify" in §4 must be confirmed against the provider schema
> (§3a) and the API discovery doc (has a `list` method?) before building.

---

## 2. Current coverage snapshot

- **~67** service keys registered (`ComputeServices` map ≈ 49 + ~18 hand-wired in
  `GetSupportedService()`).
- **88** distinct `google_*` resource types emitted across `providers/gcp/*.go`.

The authoritative current set is generated by §3b — commit it to
`docs/gcp-full-support/current-coverage.txt` so diffs are reproducible.

Note the asymmetry vs AWS: most GCP services here emit **exactly one** `google_*`
resource (the compute codegen is 1 resource per service). So the resource gap is
dominated by (a) whole missing services and (b) sub-resources of the few
multi-resource services (gcs, iam, monitoring, dns, kms, pubsub, cloudsql, gke).

---

## 3. Establish the authoritative gap list (do this once, after §7 SDK bump)

Hand-maintaining a 1000-row list is error-prone. Generate it. **Pin the provider
version** — terraformer refreshes against whatever `terraform-provider-google`
the user has installed at runtime; if the gap list is computed against a newer
provider than the user runs, you build generators that fail at refresh.

**a. Dump every resource terraform-provider-google supports** (source of truth):

```bash
GOOGLE_PROVIDER_VERSION="7.37.0"   # <- pin; bump deliberately. Latest as of 2026-06.
cd $(mktemp -d)
cat > main.tf <<EOF
terraform {
  required_providers { google = { source = "hashicorp/google", version = "= ${GOOGLE_PROVIDER_VERSION}" } }
}
EOF
terraform init >/dev/null
terraform providers schema -json \
  | jq -r '.provider_schemas[] | .resource_schemas | keys[]' \
  | sort > /tmp/tf-google-all-resources.txt
wc -l /tmp/tf-google-all-resources.txt   # ~1000+
```

Optionally repeat with `source = "hashicorp/google-beta"` →
`/tmp/tf-google-beta-all-resources.txt` for the beta surface.

Record the pinned version in `docs/gcp.md` as the supported provider floor.

**b. Dump what terraformer currently emits:**

```bash
cd providers/gcp
grep -rhoE '"google_[a-z0-9_]+"' *.go | tr -d '"' | sort -u \
  > ../../docs/gcp-full-support/current-coverage.txt
```

**c. The gap = (a) − (b):**

```bash
comm -23 /tmp/tf-google-all-resources.txt \
         docs/gcp-full-support/current-coverage.txt \
  > docs/gcp-full-support/missing-resources.txt
```

**d. Cross-check the List-API bound for compute** (so you don't add a
`resources.go` entry for a resource with no list method):

```bash
COMPUTE_JSON=$(go list -m -f '{{.Dir}}' google.golang.org/api)/compute/v1/compute-api.json
jq -r '.resources | to_entries[] | select(.value.methods.list or .value.methods.aggregatedList) | .key' \
  "$COMPUTE_JSON" | sort > /tmp/compute-listable.txt
```

Then bucket `missing-resources.txt` by `google_<service>_` prefix → that bucketing
drives the work streams. A prefix whose service is entirely absent = **new
service**; a prefix partially present = **resource gap** in an existing service.

> Prefix ≠ terraformer registry key. `google_compute_*` → the compute codegen.
> `google_container_*` → `gke`. `google_storage_*` → `gcs`. `google_sql_*` →
> `cloudsql`. `google_redis_*` → `memoryStore`. Map prefixes to keys when bucketing.

---

## 4. Gap inventory (best-effort; confirm via §3)

Three work streams, in ROI order.

### 4a. Missing COMPUTE resources (cheapest — just `resources.go` entries)

The Compute API lists far more than the ~49 generated. Candidates already
list-able and TF-supported but **absent from `resources.go`**:

| `resources.go` key | TF resource | Locality |
|---|---|---|
| `snapshots` | `google_compute_snapshot` | global (commented out — re-enable, see ignoreKeys note in resources.go) |
| `instances` | `google_compute_instance` | zonal (currently hand-written; keep hand-written) |
| `globalNetworkEndpointGroups` | `google_compute_global_network_endpoint_group` | global (commented out) |
| `regionNetworkEndpointGroups` | `google_compute_region_network_endpoint_group` | regional |
| `regionSslPolicies` | `google_compute_region_ssl_policy` | regional |
| `serviceAttachments` | `google_compute_service_attachment` | regional |
| `publicAdvertisedPrefixes` | `google_compute_public_advertised_prefix` | global |
| `publicDelegatedPrefixes` | `google_compute_public_delegated_prefix` | regional |
| `instantSnapshots` | (verify TF support) | zonal |
| `machineImages` | `google_compute_machine_image` | global |
| `networkAttachments` | `google_compute_network_attachment` | regional |
| `regionNetworkFirewallPolicies` / `networkFirewallPolicies` | `google_compute_network_firewall_policy` | regional/global |
| `regionSecurityPolicies` | `google_compute_region_security_policy` | regional |
| `interconnects` | `google_compute_interconnect` | global |
| `routerNats` | `google_compute_router_nat` (sub of router — may need parent walk) | regional |

> Add each as a `basicGCPResource{terraformName: "..."}` entry (or an override
> struct for additional fields), re-run the generator, build. One `resources.go`
> commit can carry several since regeneration is mechanical — but **commit per
> logical service** to keep refresh-validation isolated (§8).

### 4b. Partial hand-written services — known resource gaps (fill these)

| Registry key | Has | Missing (examples — confirm via §3) |
|---|---|---|
| `iam` | custom_role, member, service_account | `google_project_iam_binding`, `_policy`, `_audit_config`, `google_service_account_key`, `_iam_member/binding`, `google_organization_iam_*`, `google_folder_iam_*`, `google_iam_workload_identity_pool(_provider)` |
| `gcs` | bucket, acls, iam, notification, default_object_acl | `google_storage_bucket_object`, `_hmac_key`, `_bucket_lifecycle`/retention covered via bucket |
| `dns` | managed_zone, record_set | `google_dns_policy`, `_response_policy(_rule)`, `_managed_zone_iam_*` |
| `kms` | key_ring, crypto_key | `google_kms_crypto_key_version`, `_key_ring_iam_*`, `_crypto_key_iam_*`, `_secret_ciphertext` (n/a), `_key_ring_import_job` |
| `monitoring` | alert_policy, group, notification_channel, uptime_check | `google_monitoring_dashboard`, `_slo`, `_service`, `_custom_service`, `_monitored_project` |
| `pubsub` | topic, subscription | `google_pubsub_schema`, `_lite_topic`, `_lite_subscription`, `_lite_reservation`, `_topic_iam_*`, `_subscription_iam_*` |
| `cloudsql` | database, database_instance | `google_sql_user`, `_ssl_cert`, `_source_representation_instance` |
| `gke` | cluster, node_pool | (cluster sub-config is inlined; verify `google_gke_hub_*`, `google_gke_backup_*` belong to new services not gke) |
| `bigQuery` | dataset, table | `google_bigquery_job`(n/a—data plane), `_routine`, `_data_transfer_config`, `_connection`, `_reservation`, `_dataset_iam_*`, `_table_iam_*`, `google_bigquery_analytics_hub_*` (→ new service) |
| `cloudFunctions` | function, function2 | (covered) — verify gen2 location enumeration |
| `logging` | metric | `google_logging_project_sink`, `_billing_account_sink`, `_folder_sink`, `_organization_sink`, `_project_bucket_config`, `_metric` exists |
| `project` | project | `google_project_service`, `_iam_*` (→ iam), `_organization_policy` |

### 4c. Missing SERVICES that have terraform-provider-google resources (build these)

Grouped by priority. Each becomes one `providers/gcp/<name>.go` + a `GetSupportedService()` entry.

**P1 — common infra, high demand**

| Service key (proposed) | TF prefix | SDK (`google.golang.org/api/...` unless noted) | Notes |
|---|---|---|---|
| `secretManager` | `google_secret_manager_secret(_version)` | `secretmanager/v1` | very common |
| `serviceAccount`* | (in `iam` — expand instead) | — | see §4b iam |
| `cloudRun` | `google_cloud_run_v2_service`, `_v2_job`, `google_cloud_run_service` | `run/v2` | |
| `artifactRegistry` | `google_artifact_registry_repository` | `artifactregistry/v1` | replaces deprecated container registry |
| `spanner` | `google_spanner_instance`, `_database` | `spanner/v1` | |
| `bigtable` | `google_bigtable_instance`, `_table`, `_app_profile` | `bigtableadmin/v2` | |
| `filestore` | `google_filestore_instance`, `_backup`, `_snapshot` | `file/v1` | |
| `vpcAccess` | `google_vpc_access_connector` | `vpcaccess/v1` | serverless VPC |
| `serviceNetworking` | `google_service_networking_connection` | `servicenetworking/v1` | |
| `dataflow` | `google_dataflow_job` (verify list) | `dataflow/v1b3` | |
| `composer` | `google_composer_environment` | `composer/v1` | |
| `apigee` | `google_apigee_*` | `apigee/v1` | large |
| `notebooks` | `google_notebooks_instance`, `_runtime` | `notebooks/v1` | |
| `workflows` | `google_workflows_workflow` | `workflows/v1` | |
| `eventarc` | `google_eventarc_trigger` | `eventarc/v1` | |
| `certificateManager` | `google_certificate_manager_*` | `certificatemanager/v1` | |

**P2 — security / governance / networking**

| Service | TF prefix | SDK | Notes |
|---|---|---|---|
| `securityCenter` | `google_scc_*` | `securitycenter/v1` | source, notification_config, mute_config |
| `binaryAuthorization` | `google_binary_authorization_*` | `binaryauthorization/v1` | |
| `accessContextManager` | `google_access_context_manager_*` | `accesscontextmanager/v1` | perimeters, levels |
| `iap` | `google_iap_*` | `iap/v1` | brand, client, tunnel/web iam |
| `vpcsc` | (part of accessContextManager) | — | |
| `networkConnectivity` | `google_network_connectivity_hub`, `_spoke` | `networkconnectivity/v1` | |
| `networkServices` | `google_network_services_*` | `networkservices/v1` | mesh, gateway, edge cache |
| `networkSecurity` | `google_network_security_*` | `networksecurity/v1` | |
| `dataLossPrevention` (DLP) | `google_data_loss_prevention_*` | `dlp/v2` | |
| `essentialContacts` | `google_essential_contacts_contact` | `essentialcontacts/v1` | |
| `orgPolicy` | `google_org_policy_policy` | `orgpolicy/v2` | |
| `tags` | `google_tags_tag_key`, `_tag_value` | `cloudresourcemanager/v3` | |

**P3 — data / analytics / ML / app platform**

| Service | TF prefix | SDK | Notes |
|---|---|---|---|
| `datacatalog` | `google_data_catalog_*` | `datacatalog/v1` | |
| `datafusion` | `google_data_fusion_instance` | `datafusion/v1` | |
| `dataplex` | `google_dataplex_*` | `dataplex/v1` | lake, zone, asset |
| `firestore` | `google_firestore_database`, `_index`, `_field` | `firestore/v1` | |
| `datastream` | `google_datastream_*` | `datastream/v1` | |
| `vertexAI` | `google_vertex_ai_*` | `aiplatform/v1` | dataset, endpoint, featurestore — large |
| `documentAI` | `google_document_ai_processor` | `documentai/v1` | |
| `healthcare` | `google_healthcare_*` | `healthcare/v1` | dataset, fhir/dicom/hl7 stores |
| `gameServices` | `google_game_services_*` | (verify still supported) | |
| `looker` | `google_looker_instance` | `looker/v1` | |
| `analyticsHub` | `google_bigquery_analytics_hub_*` | `analyticshub/v1` | |
| `pubsubLite` | `google_pubsub_lite_*` | `pubsublite/v1` | (or fold into `pubsub`) |

**P4 — long tail (build on demand)**

`apphub`, `apigateway`, `backupdr`, `beyondcorp`, `biglake`, `blockchainnodeengine`,
`cloudasset` (inventory feeds), `cloudidentity` (groups/memberships),
`clouddeploy`, `clouddomains`, `cloudids`, `colab`, `databasemigrationservice`
(`google_database_migration_service_*`), `dialogflow(cx)`, `discoveryengine`,
`firebase*` (separate provider surface — verify), `gkehub`/`gkeonprem`,
`integrationconnectors`, `memcache` (`google_memcache_instance`),
`migrationcenter`, `oracledatabase`, `parallelstore`, `privateca`
(`google_privateca_*`), `publicca`, `recaptchaenterprise`, `redis_cluster`
(`google_redis_cluster` — fold into `memoryStore`), `securesourcemanager`,
`servicedirectory`, `storagetransfer` (`google_storage_transfer_job`),
`tpu` (`google_tpu_*`), `vmwareengine`, `workbench`, `workstations`.

> The §3 diff produces the *complete* per-service gap; the tables above are the
> high-value starter set, not exhaustive.

---

## 5. Recipe — add COMPUTE resources (via code generator)

This is the laziest, highest-leverage path — use it for anything `google_compute_*`.

1. Confirm the resource is **list-able** (`/tmp/compute-listable.txt` from §3d)
   and **TF-supported** (`/tmp/tf-google-all-resources.txt`).
2. Add an entry to `gcp_compute_code_generator/resources.go`:

```go
"serviceAttachments": basicGCPResource{
    terraformName: "google_compute_service_attachment",
},
```

   For resources needing extra refresh fields / empty-value allowances / a zoned
   import ID, use an override struct (see `backendServices`,
   `instanceGroupManagers`, `globalForwardingRules` and the `basicGcpResource.go`
   interface methods: `getAdditionalFields`, `getAdditionalFieldsForRefresh`,
   `getAllowEmptyValues`, `ifNeedRegion`, `ifNeedZone`, `ifIDWithZone`).
3. Re-run the generator (after the §0c path fix):

```bash
cd gcp_compute_code_generator && go run . && cd ..
gofmt -w providers/gcp/*_gen.go providers/gcp/compute.go
go build -v
```

   This (re)writes `providers/gcp/<key>_gen.go` and `providers/gcp/compute.go`.
   **Never hand-edit `*_gen.go`.**
4. The new key auto-registers in `ComputeServices`, which `GetSupportedService()`
   already merges — no `gcp_provider.go` change needed for compute.
5. Add the resource to `docs/gcp.md`.
6. Add a resource-connection mapping in `GetResourceConnections()` if it
   references another resource by self_link (optional, for `--connect`).

## 5b. Recipe — add a NEW hand-written service

For non-compute services. Reference: `providers/gcp/memoryStore.go` (simple
single-resource) and `gke.go` (multi-resource + PostConvertHook).

1. **Create `providers/gcp/<service>.go`:**

```go
package gcp

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"google.golang.org/api/<svc>/<ver>"
)

var <svc>AllowEmptyValues = []string{""}
var <svc>AdditionalFields = map[string]interface{}{}

type <Svc>Generator struct {
	GCPService
}

func (g *<Svc>Generator) InitResources() error {
	ctx := context.Background()
	svc, err := <svc>.NewService(ctx)
	if err != nil {
		return err
	}
	project := g.GetArgs()["project"].(string)

	// List/AggregatedList; for regional resources read region from args.
	resp, err := svc.Projects.Locations.<Things>.List(parent).Do()
	if err != nil {
		return err
	}
	for _, obj := range resp.<Things> {
		g.Resources = append(g.Resources, terraformutils.NewResource(
			obj.Name,            // import ID — MUST match terraform import (often projects/.../...)
			obj.Name,            // terraform resource name
			"google_<tf_type>",
			g.ProviderName,
			map[string]string{
				"name":    obj.Name,
				"project": project,
			},
			<svc>AllowEmptyValues,
			<svc>AdditionalFields,
		))
	}
	return nil
}
```

2. **Register it** in `gcp_provider.go` → `GetSupportedService()`:

```go
services["<service>"] = &GCPFacade{service: &<Svc>Generator{}}
```

   **Key naming:** camelCase matching the existing keys (`secretManager`,
   `artifactRegistry`, `cloudRun`). Keep new keys consistent; don't churn legacy
   ones.

3. **Docs:** add the key + emitted resources to `docs/gcp.md`.

**Per-resource gotchas (GCP-specific):**

- **Import ID is the killer detail.** GCP import IDs are frequently
  `projects/{project}/locations/{loc}/{collection}/{id}` or `{region}/{name}` or
  `{zone}/{name}`, **not** the bare name. Check each resource's "Import" section
  in the provider docs and build the ID to match, or refresh fails. The compute
  codegen already handles `zone/name` via `ifIDWithZone`.
- **Required refresh attributes.** The provider often needs `project`,
  `location`/`region`/`zone`, and sometimes a parent (`cluster`, `instance`) in
  the attributes map for `Refresh` to succeed — see `gke.go` node pools.
- **Discovery vs GAPIC pagination differ.** Discovery clients use
  `.Pages(ctx, func(page) error {...})` or `.Do()` + `NextPageToken`; GAPIC
  clients use `iterator.Done`. Match the SDK family.
- **Locality.** Many services are regional/zonal — honor
  `g.GetArgs()["region"].(compute.Region)`. Global services ignore it.
- **Never deref a nil pointer** from the SDK response; guard optional fields.

## 6. Recipe — expand an EXISTING hand-written service (resource gaps, §4b)

1. From §3's per-service gap, list missing `google_*` types for that file.
2. Add an enumeration block per missing type inside the existing
   `InitResources()` (or a helper). Reuse the already-built client when the SDK
   package is the same; create a new client otherwise.
3. IAM sub-resources (`*_iam_member/binding/policy`) usually require walking the
   parent resources and emitting per-parent — decide whether the value justifies
   the parent walk (often it does for `gcs`, `pubsub`, `kms`).
4. Add to that service's `PostConvertHook()` if attribute rewriting is needed
   (see `gke.go` for the pattern).
5. Update `docs/gcp.md`.

---

## 7. SDK upgrade (do FIRST — it sets the floor)

The user's first ask. Current pins: `google.golang.org/api v0.214.0` (Dec 2024),
assorted `cloud.google.com/go/* v1.x`. Latest `terraform-provider-google` is
**7.37.0** (2026-06). Steps:

1. **Bump the Go API client + GAPIC modules:**

```bash
go get google.golang.org/api@latest
go get cloud.google.com/go/storage@latest cloud.google.com/go/logging@latest \
       cloud.google.com/go/iam@latest cloud.google.com/go/monitoring@latest \
       cloud.google.com/go/cloudbuild@latest cloud.google.com/go/cloudtasks@latest
go mod tidy
go build -v ./...
```

2. **Fix the codegen path hack (§0c)** so the regenerator reads `compute-api.json`
   from the module cache, then **regenerate compute** to pick up new API fields:

```bash
cd gcp_compute_code_generator && go run . && cd ..
gofmt -w providers/gcp/*_gen.go providers/gcp/compute.go
git diff --stat providers/gcp/   # review what the bump changed in generated code
```

3. **Handle breaking changes.** The discovery clients are generally additive, but
   GAPIC majors (`cloud.google.com/go/*`) occasionally rename. Fix compile
   errors per package; keep beta-versioned imports (`container/v1beta1`,
   `sqladmin/v1beta4`, `cloudscheduler/v1beta1`) unless a stable GA version now
   exists — prefer GA where available.
4. **Re-pin the provider floor** in `docs/gcp.md` (`= 7.37.0`) and run §3 against
   it so the gap list matches what users will refresh against.
5. `go test ./...` + `golangci-lint run`. Commit the bump as **its own PR**
   ("chore(gcp): bump google.golang.org/api + regenerate compute") before any
   feature work — isolates the diff and the risk.

> Caveat: bumping `google.golang.org/api` may shift transitive deps shared with
> other providers (AWS uses different SDKs, so low risk, but `go mod tidy` will
> touch `go.sum`). Build the whole repo, not just `providers/gcp`.

---

## 8. Phased rollout

Each service = **one focused PR** (generator/file + registry + docs + connections),
committed and validated independently — matches the AWS cadence and the user's
"one service at a time" ask. Keeps reviews small and isolates SDK-version risk.

- **Phase 0 — foundations:** fix the codegen `compute-api.json` path (§0c); add a
  Makefile target / script to run the generator reproducibly; commit. Small,
  unblocks compute.
- **Phase 1 — SDK bump (§7):** bump `google.golang.org/api` + GAPIC modules,
  regenerate compute, fix breakage, re-pin provider floor. One PR.
- **Phase 2 — gap tooling (§3):** run the diff against pinned 7.37.0, commit
  `current-coverage.txt`, `tf-google-all-resources.txt`, `missing-resources.txt`,
  `compute-listable.txt`; bucket by service. Produces the real backlog.
- **Phase 3 — compute resource gaps (§4a):** cheapest wins — `resources.go`
  entries + regenerate. snapshots, machineImages, serviceAttachments,
  region/global NEGs, network firewall policies, publicAdvertised/DelegatedPrefixes.
- **Phase 4 — partial hand-written service gaps (§4b):** highest ROI on
  already-wired services: `iam` (bindings/policies/workload-identity), `gcs`
  (objects/hmac), `pubsub`, `kms`, `monitoring`, `dns`, `cloudsql`, `bigQuery`,
  `logging` sinks.
- **Phase 5 — P1 new services (§4c):** secretManager, cloudRun, artifactRegistry,
  spanner, bigtable, filestore, composer, vpcAccess, certificateManager,
  serviceNetworking, workflows, eventarc, notebooks.
- **Phase 6 — P2 security/governance/networking:** securityCenter, binaryAuth,
  accessContextManager, iap, networkConnectivity/Services/Security, DLP,
  orgPolicy, tags, essentialContacts.
- **Phase 7 — P3 data/ML/app:** firestore, dataplex, datacatalog, datafusion,
  datastream, vertexAI, healthcare, analyticsHub, looker.
- **Phase 8 — P4 long tail:** build on demand; decide per service (ML/IoT/media
  niche may not be worth maintenance).

Re-run §3 after each PR and shrink `missing-resources.txt`. Done = diff empty
modulo the documented "no list API" exclusions.

---

## 9. Validation per service

Mirror the AWS bar (the codebase's only generator test, `sg_test.go` on the AWS
side, tests a *pure transform*, not a mocked SDK — GCP has even fewer tests, so
hold the same line):

- **Unit (conditional):** only if `InitResources` does non-trivial result
  processing (parent-walk joins, the `gke` node-pool cluster-ref rewrite, ID
  composition) — extract it to a pure func and test that. Trivial
  list→`NewResource` mappers need none. **Do not mock the GCP SDK** — the
  codebase doesn't.
- **Codegen sanity (compute):** after regenerating, `go build` + `gofmt -l`
  (must be empty) + confirm the new `*_gen.go` + `compute.go` entry exist and
  `git diff` of unrelated generated files is empty (proves the generator is
  deterministic and the bump didn't churn unrelated services).
- **Integration (the correctness bar):**
  `go build -o terraformer && ./terraformer import google --resources=<svc> --projects=<proj> --regions=<region>`
  against a test project, then `terraform plan` on the output shows **no diff**
  (refresh round-trips cleanly). Use the provider version pinned in §3. Test the
  `--provider-type beta` path for any beta-only resource.
- `go test ./...` (full suite pulls the provider plugin via `providerwrapper`;
  needs module cache + network + a terraform-provider-google binary resolvable).
- `golangci-lint run` (gocritic/revive/unconvert/unparam + gofmt/goimports).

---

## 10. Effort estimate

- Compute resource gaps (§4a): ~15 `resources.go` entries, mechanical, 1–2 PRs.
- Partial hand-written gaps (§4b): ~10 services touched, ~50 resource blocks
  (IAM bindings dominate).
- New services (§4c): ~40 services across P1–P4.
- Total new `google_*` reachable: the §3 diff gives the exact count. Terraformer
  emits **88** today against a provider surface of **~1000+** — expect the gap to
  be **600–800** reachable (the rest are no-list-API singletons, IAM-policy-only,
  or data-plane).

Track by re-running §3 after each PR.

---

## 11. Performance & scale notes

- Service init runs **sequentially** (same engine as AWS; the `WaitGroup` in
  `cmd/import.go` is dead code). A full `--resources=*` across many projects ×
  regions = Σ(every service's serial paginated calls) × regions × projects —
  hours at full coverage.
- GCP adds a **projects × regions** outer loop in `provider_cmd_google.go`, so
  cost multiplies by both. Recommend users scope `--resources`, `--projects`,
  and `--regions` to what they need.
- Compute zonal resources loop **every zone in the region** — already handled by
  the codegen, but it's N API calls per zonal service per region.
- Deferred (shared with AWS, see `TODOS.md`): parallel service execution,
  per-service memory streaming, `context` timeouts on the `.Do()`/`.Pages()`
  calls (one hung call hangs the run — GCP calls use `context.Background()` with
  no deadline today).

---

## 12. Open decisions

> **Resolved 2026-06-23** (see `phases.md`): scope = full P4 (diff-empty);
> IAM = import **all** `*_iam_*` via parent walks (not just containers);
> beta = build **GA + beta**; provider floor pinned **7.37.0** (verified latest
> GA), SDK bump to `google.golang.org/api v0.286.0`. Remaining items below are
> per-service judgment calls, not blockers.

- **Un-listable resources:** maintain `docs/gcp-full-support/no-list-api.md` so
  "missing" never silently means "impossible" (IAM-policy-only resources,
  project singletons, data-plane).
- **beta vs GA:** decide per service whether to require `--provider-type beta`.
  Prefer GA discovery clients; document beta-only resources explicitly.
- **GAPIC vs discovery clients:** standardize new services on
  `google.golang.org/api/<svc>` (discovery) for consistency; use
  `cloud.google.com/go/<svc>` only when discovery lacks the List method.
- **Fold vs new service:** `google_redis_cluster` → extend `memoryStore` or new
  key? `google_pubsub_lite_*` → extend `pubsub` or new `pubsubLite`? Decide by
  whether the SDK package differs (different package → new service is cleaner).
- **Codegen scope creep:** the generator is compute-only. Resist generalizing it
  to other services unless ≥1 new service family is itself large and
  discovery-doc-driven — YAGNI until proven.
- **IAM explosion:** `*_iam_member/binding/policy` exist for nearly every
  resource type. Importing them requires parent walks and can 3× the resource
  count. Decide a policy: import IAM only for top-level containers
  (project/folder/org + bucket/topic/dataset) unless asked, and document it.
</content>
</invoke>
