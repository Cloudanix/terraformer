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
- **Phase 5 P1** ✓ 12 new services: secretManager, artifactRegistry, spanner,
  bigtable, cloudRun, filestore, workflows, eventarc, vpcAccess, composer,
  notebooks, certificateManager. Plus pubsub→schema (Phase 4b).

Coverage 88 → 129 emitted `google_*` types.

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
