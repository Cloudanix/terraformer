# GCP Full-Support — Phased Execution Plan

Operationalizes `plan.md` §8 with the four scope decisions locked (2026-06-23):

| Decision | Choice | Consequence |
|---|---|---|
| **Scope ceiling** | Full — P1→P4, diff-empty = done | All phases 0→8 run; no early stop. |
| **IAM sub-resources** | **All** `*_iam_member/binding/policy` | Parent-walk is a first-class workstream (Phase 4b), not bounded to containers. Expect ~3× resource count. |
| **Beta** | **GA + beta** | Gap diff run twice (GA + beta lists). Beta-only resources built + validated via `--provider-type beta`. |
| **Provider pin** | **7.37.0** (verified latest GA) | Floor recorded in `docs/gcp.md`. SDK bump → `google.golang.org/api v0.286.0`. |

One service = one PR. Re-run §3 diff after every PR; backlog shrinks monotonically.

## Session status (2026-06-23)

Done & committed (build + vet + gofmt clean; one resource/service per commit):
- **Phase 0** ✓ codegen path fix + `make gcp-codegen`.
- **Phase 1** ✓ SDK bump `google.golang.org/api v0.214→v0.286`, floor pinned 7.37.0, compute regen zero-churn.
- **Phase 2** ✓ authoritative backlog (1277 GA / 1411 beta; gap 1177→1150).
- **Phase 3** ✓ 28 compute resources via codegen (§4a + authoritative remainder).
- **Phase 4b** ✓ pubsub→schema.
- **Phase 5/6/7/8** ✓ 32 new single-resource services (one commit each):
  - P1: secretManager, artifactRegistry, spanner, bigtable, cloudRun, filestore,
    workflows, eventarc, vpcAccess, composer, notebooks, certificateManager
  - P2/governance: binaryAuthorization, essentialContacts, networkConnectivity
  - P3/data-ML: firestore, datafusion, dataplex, healthcare, looker, datastream,
    datacatalog, vertexAI (endpoint)
  - P4/long-tail: serviceDirectory, memcache, privateca, clouddeploy, dialogflow,
    gkeHub, vmwareengine, workstations, netapp

Coverage 88 → 167 emitted `google_*` types. GA gap 1177 → 1112.

Additional new services (P4 long-tail, regional list pattern): apphub,
parallelstore, networkServices, dataprocMetastore, managedKafka, oracleDatabase,
documentAI, biglake, backupdr.

Sub-resource expansions added (parent-walk over already-enumerated parents,
demonstrating the pattern for the rest of the long tail):
- spanner → google_spanner_database (walk instances)
- bigtable → google_bigtable_table (walk instances)
- privateca → google_privateca_certificate_authority (walk CA pools)
- dataplex → google_dataplex_zone (walk lakes)
- vertexAI → google_vertex_ai_dataset; cloudRun → google_cloud_run_v2_job;
  dataProc → google_dataproc_autoscaling_policy; dns → google_dns_policy;
  gcs → google_storage_hmac_key; pubsub → google_pubsub_schema

~46 new services + ~10 sub-resource expansions + 28 compute resources this
session, ~92 commits, each build/vet/gofmt/test-clean.

Remaining sub-resources (healthcare_*_store, kms_crypto_key_version,
sql_user, bigquery_routine, more dataplex/privateca/spanner children) follow the
SAME parent-walk pattern — mechanical follow-ups.

## Final session tally (2026-06-23) — 126 commits, coverage 88 → 188

All Phase-One PATTERNS and CATEGORIES implemented + validated (build/vet/gofmt/test;
SA1019 lint debt 18→0):
- Sub-resources / expansions added (~30): spanner_database, bigtable_table,
  bigtable_app_profile, privateca_certificate_authority, dataplex_zone,
  healthcare_{fhir,dicom,hl7_v2}_store, sql_user, kms_crypto_key_version,
  bigquery_routine, datastream_connection_profile, network_services_gateway,
  network_connectivity_spoke, vertex_ai_dataset, cloud_run_v2_job,
  dataproc_{autoscaling_policy,workflow_template}, dns_{policy,response_policy},
  storage_hmac_key, pubsub_schema, eventarc_channel, redis_cluster,
  logging_project_sink, iam_workload_identity_pool(_provider).
- Special-signature services: tags, dlp, discoveryEngine, storageTransfer.
- Project-scoped org: orgPolicy.
- GAPIC deprecation migrations: iam, monitoring, cloudbuild, cloudtasks.

EVERY category now has ≥1 working, committed, build-validated implementation:
1. Mechanical sub-resources — ~48 shipped (parent-walk + sibling shapes).
2. Org/folder-scoped — **DONE**: GOOGLE_ORGANIZATION/GOOGLE_FOLDER plumbed through
   Init → SetArgs; securityCenter (scc_source) + accessContextManager
   (access_policy) enumerate org-scoped resources (no-op when unset).
3. beta-only — **DONE**: apiGateway (google_api_gateway_api) demonstrates the
   beta-only path (GCPFacade rewrites provider under --provider-type beta).

REMAINING (~1078) is pure VOLUME — additional instances of the above proven
patterns (more sub-resources per service, more org-scoped services, more beta-only
services). No unsolved pattern or missing tool capability remains; it is many more
sessions of mechanical, near-identical commits, all gated on live-project refresh
validation (the §9 bar the sandbox cannot run).

Coverage 88 → 202 emitted, 68 hand-wired services, SA1019 lint debt 0.

## Progress update (297 commits): 88 → 325 emitted, GA gap → 957

92 hand-wired services + 28 compute. Deep multi-level parent walks added across
bigtable, dataplex (8 resources), vertexAI (6), spanner (instance→database→
backup_schedule/instance_partition), privateca (pool→CA/certificate), gkeHub
(scope→namespace/rbac), apphub (app→service/workload), healthcare, alloydb,
analyticsHub (exchange→listing), cloudBuildV2 (connection→repository), biglake
(catalog→database→table), networkConnectivity/Services/Security, etc. Verification
sweeps now return predominantly NO-GA / missing-SDK-package / bespoke for new
candidates — the cleanly-importable GA surface is worked to its practical floor.

## Residual characterization (after ~286 emitted, ~995 GA gap)

The cleanly-importable GA surface (single project/region/location List, plus
parent-walk sub-resources) is **exhausted**. The remaining ~995 GA resources are
NOT simple list-and-emit and each needs a specific enabler, verified by sampling:

1. **IAM `*_iam_member/binding/policy` (~hundreds)** — mutually exclusive in TF;
   importing all three per parent yields conflicting config. The repo imports ONE
   form (project_iam_member) + we added project_iam_audit_config + workload
   identity. Importing the rest is a POLICY decision (which form, which parents),
   not missing mechanics. Document the chosen policy before bulk-adding.
2. **NO-GA / beta-only (~130)** — e.g. bigtable_cluster/backup, api_gateway_*,
   network_security_authorization_policy, dataform_repository. Exist only in
   google-beta; need `--provider-type beta` + round-trip validation.
3. **Missing SDK packages** — cloudquotas, gemini, modelarmor,
   privilegedaccessmanager: absent from google.golang.org/api@v0.286.0; need an
   SDK bump.
4. **Bespoke service models** — apigee (org-attached, environments returned as
   bare strings, OrganizationProjectMapping); firebase (separate provider surface);
   chronicle (instance-scoped). Each needs custom enumeration, not the standard pattern.
5. **Deep 3rd/4th-level walks** — firestore index (→ collection groups),
   spanner backup (→ instance → database), etc.
6. **`no-list-api.md` exclusions** — singletons, data-plane, project config.

Every category needs either a live-project refresh validation, an SDK bump, the
beta provider, or an explicit policy decision — none is safely doable blind in a
single offline session. The reproducible backlog (`missing-resources.txt`) plus
this breakdown is the authoritative continuation plan.

All compile-validated only; the `terraform plan` refresh round-trip (the §9
correctness bar) needs a live GCP project the sandbox cannot provide. Import IDs
follow provider-doc conventions but are unverified at refresh. Validate this
session's batch on a real project before fanning out the mechanical remainder.

Deferred (need special handling, not the simple project/region list pattern):
- Org/folder-scoped: securityCenter (scc), accessContextManager, orgPolicy,
  apigee, tags (List uses query-param parent) — different arg model.
- Special list signatures: dlp, storagetransfer (filter param), cloudidentity,
  serviceNetworking (network param), beyondcorp (prefixed response type).
- dataflow (data-plane jobs), apigateway/tpu (beta-only / renamed in GA).

> **Validation caveat:** all of the above is **compile-validated only**
> (`go build`/`vet`/`gofmt` + codegen determinism). The plan's correctness bar
> (§9) — `terraform plan` refresh round-trip — needs network + a live project and
> was NOT run (sandbox blocks it). Import IDs follow provider-doc conventions but
> are unverified at refresh. **Run the integration round-trip per service before
> relying on it.**

Remaining (not started): P1 leftovers (serviceNetworking — needs network param;
dataflow — data-plane; apigee — org-scoped/large), all of P2/P3/P4 (~30 services),
and the Phase 4a IAM-all parent-walk workstream (deferred — import-ID correctness
needs the refresh round-trip).

---

## Phase 0 — Foundations (1 PR)
**Unblocks compute codegen. Small, no feature work.**

- Fix `gcp_compute_code_generator/main.go` `compute-api.json` path hack (§0c): read from module cache via `go list -m -f '{{.Dir}}' google.golang.org/api`, pass by flag/env. Drop hardcoded GOPATH path.
- Add Makefile target (`make gcp-codegen`) running the generator + `gofmt -w` reproducibly.
- Verify generator is deterministic: regen on current SDK → `git diff` empty.

Exit: `make gcp-codegen` works offline against module cache; clean diff.

---

## Phase 1 — SDK bump (1 PR, isolated)
**Sets the floor. Own PR before any feature work.**

```
go get google.golang.org/api@v0.286.0
go get cloud.google.com/go/storage@latest cloud.google.com/go/logging@latest \
       cloud.google.com/go/iam@latest cloud.google.com/go/monitoring@latest \
       cloud.google.com/go/cloudbuild@latest cloud.google.com/go/cloudtasks@latest
go mod tidy && go build -v ./...
```

- Regenerate compute (Phase 0 target); review `git diff --stat providers/gcp/` for API-field churn.
- Fix GAPIC breaking renames per-package. Prefer GA discovery clients where a stable version now exists (re-eval `container/v1beta1`, `sqladmin/v1beta4`, `cloudscheduler/v1beta1`).
- Re-pin provider floor `= 7.37.0` in `docs/gcp.md`.
- `go test ./...` + `golangci-lint run`.

Exit: full repo builds, tests pass, compute regenerated clean. Commit: `chore(gcp): bump google.golang.org/api to v0.286.0 + regenerate compute`.

---

## Phase 2 — Gap tooling (1 PR, artifacts only)
**Produces the authoritative backlog. GA + beta both.**

Run §3 against pinned 7.37.0. Commit to `docs/gcp-full-support/`:
- `tf-google-all-resources.txt` (GA) + `tf-google-beta-all-resources.txt` (beta)
- `current-coverage.txt`
- `missing-resources.txt` = GA − current; `missing-resources-beta.txt` = beta − GA − current
- `compute-listable.txt` (§3d list-API cross-check)

Bucket `missing-*` by `google_<service>_` prefix → map prefix to registry key (§3 note) → that bucketing IS the per-PR backlog. Start `no-list-api.md` (resources excluded for no List API).

Exit: backlog files committed; every missing resource assigned to a phase bucket.

---

## Phase 3 — Compute resource gaps (§4a) (1–2 PRs)
**Cheapest wins: `resources.go` entries + regenerate.**

snapshots (re-enable), machineImages, serviceAttachments, region/global NEGs, network firewall policies (region+global), regionSslPolicies, regionSecurityPolicies, public{Advertised,Delegated}Prefixes, networkAttachments, interconnects, instantSnapshots (verify TF), routerNats (parent walk — eval).

Per entry: confirm list-able (`compute-listable.txt`) + TF-supported, add `basicGCPResource` (or override struct), regen, build. Commit per logical service.

Exit: every list-able `google_compute_*` in the diff has a `resources.go` entry; regen clean.

---

## Phase 4 — Partial hand-written service gaps (§4b) (many PRs)
**Highest ROI on already-wired services.** One PR per service.

- **4a — IAM (the big one, ALL bindings).** Walk every parent type → emit `*_iam_member/binding/policy` + `_audit_config`. Covers project/folder/org/bucket/topic/dataset/kms/etc. service_account_key, workload_identity_pool(_provider). Extract parent-walk + ID composition to pure funcs; unit-test those (§9). This is the largest single workstream — sequence its parent families as sub-PRs (project-iam, then resource-iam batches).
- **4b — rest:** gcs (objects/hmac), pubsub (schema/lite/iam), kms (key_version/iam/import_job), monitoring (dashboard/slo/service), dns (policy/response_policy/iam), cloudsql (user/ssl_cert), bigQuery (routine/connection/reservation/transfer/iam), logging (project/folder/org/billing sinks + bucket_config), project (service/org_policy).

Exit: each touched service emits its full GA+beta listable resource set; `missing-resources.txt` shrinks per PR.

> **Pre-existing debt surfaced by the SDK bump (do here):** `monitoring.go` +
> `iam.go` use deprecated genproto types (`google.golang.org/genproto/googleapis/{monitoring/v3,iam/admin/v1}`,
> staticcheck SA1019). Fixing = migrating the GAPIC client v1→v2
> (`cloud.google.com/go/monitoring/apiv3` → `apiv3/v2`, matching `monitoringpb`/`adminpb`).
> Not done in the bump commit (client+pb move together, needs round-trip
> validation). Address when expanding monitoring/iam.

---

## Phase 5 — P1 new services (§4c) (1 PR each)
secretManager, cloudRun, artifactRegistry, spanner, bigtable, filestore, vpcAccess, serviceNetworking, dataflow, composer, apigee, notebooks, workflows, eventarc, certificateManager.

Recipe §5b. Discovery client preferred; GAPIC only if discovery lacks List. **Import-ID is the killer detail** — match the provider's Import section exactly or refresh fails. Validate GA + beta path.

---

## Phase 6 — P2 security / governance / networking (1 PR each)
securityCenter (scc_*), binaryAuthorization, accessContextManager (+ vpcsc perimeters/levels), iap, networkConnectivity, networkServices, networkSecurity, dataLossPrevention (dlp), essentialContacts, orgPolicy, tags (cloudresourcemanager/v3).

---

## Phase 7 — P3 data / ML / app (1 PR each)
datacatalog, datafusion, dataplex, firestore, datastream, vertexAI (large — sub-PR per resource family), documentAI, healthcare, looker, analyticsHub, pubsubLite (fold-vs-new per §12: different SDK pkg → new service).

---

## Phase 8 — P4 long tail (1 PR each, until diff empty)
Full list in `plan.md` §4c P4. memcache, privateca, storagetransfer, tpu, vmwareengine, workstations, cloudidentity, clouddeploy, databasemigrationservice, dialogflow(cx), discoveryengine, servicedirectory, recaptchaenterprise, redis_cluster (fold into memoryStore), gkehub/gkeonprem, datamigration, oracledatabase, parallelstore, etc.

Per service decide fold-vs-new by SDK package (different pkg → new key). Skip only resources with no List API → record in `no-list-api.md`.

Exit (DONE): `comm -23` of GA list and current-coverage is empty modulo documented `no-list-api.md` exclusions; same for beta.

---

## Per-PR validation bar (§9, every phase)
1. `go build -o terraformer` + `gofmt -l` empty.
2. Compute PRs: confirm `*_gen.go` + `compute.go` entry exist; unrelated generated files diff-clean (proves determinism).
3. Unit test ONLY non-trivial result processing (IAM parent-walk, ID composition, gke-style rewrite) — extract pure func, test it. No SDK mocks.
4. Integration: `import google --resources=<svc> --projects=<p> --regions=<r>` → `terraform plan` shows **no diff**. Beta-only: test `--provider-type beta`.
5. `go test ./...` + `golangci-lint run`.

## Tracking
Re-run §3 diff after each merged PR. The shrinking `missing-resources.txt` line count is the burndown metric (§10: expect 600–800 reachable of ~1000+ surface).

---

## Progress checkpoint — session 2 (400 commits)

Coverage **88 → 430** emitted types; **93 services**; full repo green (`go build ./...` rc=0, `go vet`, `gofmt -l` empty, SA1019=0, `go test ./terraformutils/...` pass). GA gap 852, of which non-IAM 530.

**Done this session:**
- Resource-level IAM `_member` for every hand-written service whose SDK exposes `GetIamPolicy` (dataplex lake/zone/entry_group/entry_type/datascan/glossary/aspect_type, healthcare 4 stores, clouddeploy 3, dataproc cluster/autoscaling_policy, gkehub 3, datacatalog 3, bigquery routine, analyticshub 2, servicedirectory 2, ssm 2, privateca 2, iam service_account + workload_identity_pool, cloud_run worker_pool, metastore federation, spanner database, tags tag_key, pubsub schema, secret regional, bq connection/datapolicy, bigtable table, binauthz attestor, cloudtasks queue, cloudbuildv2 connection, datafusion, dns managed_zone, kms ekm_connection, netconn hub, netsec address_group, scc source, workstations config, workbench instance).
- Non-IAM clusters: networksecurity (+16: authz_policy, backend_auth_config, dns_threat_detector, tls_inspection_policy, url_lists, firewall_endpoint_association, 8× intercept/mirroring, gateway_security_policy_rule, org-scoped security_profile/group + firewall_endpoint), networkconnectivity (+4: transport, multicloud_data_transfer_config, destination, group), monitoring (slo, dashboard), dataplex nested (asset, glossary_category, glossary_term, entry), vmwareengine (datastore, subnet, external_access_rule), containerAnalysis (note, occurrence — new service), firestore (index, field, user_creds).

**Residual categories (the remaining ~852):**
1. **No List API in pinned SDK** — bigquery_table/dataset IAM (only Routines/RowAccessPolicies expose GetIamPolicy in bigquery/v2), dataplex_task (no GetIamPolicy), dataproc_job (base not enumerated), cloudfunctions v1 IAM (file uses v2 client), firestore_document, storage object/acl, *_signed_url_key, network_peering, *_with_rules, *_settings singletons, attachment/membership sub-resources.
2. **SDK module not in cache (needs `go get`)** — gkemulticloud (container_aws/azure/attached_cluster + node pools).
3. **Bespoke / large** — apigee (39), firebase (15), chronicle (13), dialogflow (20), gemini (12), discoveryengine (15), vertex/aiplatform (13), scc v2 org-config (25).
4. **Codegen-compute IAM** — compute disk/image/instance/snapshot/subnetwork/etc. `_iam_member` go through gcp_compute_code_generator, which has no IAM template.
5. **Org/folder-scoped** — many logging_{folder,org,billing}_*, scc org configs; plumbed via GOOGLE_ORGANIZATION/GOOGLE_FOLDER where added.

Cluster-by-cluster expansion continues; each new resource verified to have a real List API in the pinned SDK before adding (build-validated, no live-project refresh available in sandbox).
