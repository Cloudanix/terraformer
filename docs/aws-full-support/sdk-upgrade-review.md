# AWS services addable after the SDK upgrade — review (2026-06-23)

Scope: what net-new AWS coverage the recent `aws-sdk-go-v2` upgrade (84 service
modules to latest) actually unlocks. Method: diff the 1453-row TF resource
universe (`tf-aws-all-resources.txt`) against the 1346 `aws_*` types terraformer
emits; resolve the 184-row gap (`missing-resources.txt`) against the vendored
SDK module set and `no-list-api.md` exclusion classes.

**Headline:** net-new *services* are essentially complete. The upgrade already
landed its wins — `ecr_repository_creation_template`, `eks_access_entry`,
`eks_access_policy_association`, `lambda_function_recursion_config`,
`dynamodb_resource_policy`, `codebuild_fleet`, `vpc_security_group_vpc_association`,
`iam_organizations_features` are all emitted now. What remains is a short, exact
list below.

## Tier 1 — buildable now, SDK already vendored, NO dep bump (do these)

| Service | Add | API | Notes |
|---|---|---|---|
| evidently | `aws_evidently_feature`, `aws_evidently_launch`, `aws_evidently_segment` | `ListFeatures`/`ListLaunches` (per project), `ListSegments` (account) | `evidently.go` already paginates projects → iterate them for feature/launch; segment is account-level. evidently@v1.30.0 vendored. |
| computeoptimizer | `aws_computeoptimizer_recommendation_preferences` | `GetRecommendationPreferences` | Singleton config; needs a `resourceType` arg loop. Marginal value. |

evidently is the clean win: 3 resources, zero new deps, generator pattern already
in the file.

## Tier 2 — real services, need a one-module dep bump (`go get` each)

| Service | Add | API | Cost |
|---|---|---|---|
| drs (Elastic Disaster Recovery) | `aws_drs_replication_configuration_template` | `DescribeReplicationConfigurationTemplates` | `drs.go` exists but is an empty stub. +1 SDK module. |
| codecatalyst | `aws_codecatalyst_project`, `_dev_environment`, `_source_repository` | `ListSpaces`→`ListProjects`→`ListDevEnvironments`/`ListSourceRepositories` | `codecatalyst.go` empty stub. Multi-step (space-scoped). +1 SDK module. |
| serverlessapplicationrepository | `aws_serverlessapplicationrepository_cloudformation_stack` | `ListApplications` | Semantics iffy — TF resource *deploys* a stack; List returns owned apps. **Skip unless requested.** |

## Tier 3 — correctly excluded, do NOT add (confirmed)

- **AWS-deprecated:** simpledb, worklink.
- **Superseded:** lex v1 (`lexmodelbuildingservice`) — v2 `aws_lexv2models_*` already built.
- **Data-plane / meta / no-import:** redshiftdata, cloudcontrolapi,
  `aws_customerprofiles_profile`, snapshot/ami copy/launch-permission,
  spot_datafeed, ses identity verification, transfer_tag, etc.
- **~150 structural suffixes with no List API:** `*_attachment`, `*_policy`,
  `*_accepter`, `*_association`, `*_tag`, `*_validation`, `aws_default_*`,
  inline-policy conflicts (sqs/sns `*_policy`). Catalogued in `no-list-api.md`.

## Doc-hygiene fix (stale, hurts the audit's trust)

`no-list-api.md` → "SDK not vendored in the pinned aws-sdk-go-v2 module set" is
**out of date**. These are now vendored AND built, and must be removed from that
list: **evidently** (project), **devopsguru** (4 resources), **paymentcryptography**
(key, key_alias), **computeoptimizer** (enrollment_status), **costoptimizationhub**
(enrollment_status, preferences), **customerprofiles** (domain). Only still-unvendored:
cloudwatchevidently-v?/lex-v1/drs/codecatalyst/serverlessapplicationrepository/
simpledb/worklink/redshiftdata/cloudcontrolapi.

## Recommendation

1. Ship **evidently feature/launch/segment** — no dep bump, 3 resources, trivial.
2. Optionally **drs** (+1 dep, 1 clean resource) and **codecatalyst** (+1 dep, 3
   resources, multi-step) if disaster-recovery / dev-tooling coverage is wanted.
3. Correct the `no-list-api.md` unvendored section.
4. Everything else in the 184-row gap is a documented exclusion class — leave it.

## Outcome (2026-06-23, this session)

- **evidently — DROPPED.** The vendored SDK is marked deprecated ("AWS has
  deprecated this service… no longer available for use"). CloudWatch Evidently
  is EOL. Building feature/launch/segment for a dead service = waste; the
  existing `aws_evidently_project` stays only for historical reasons. Logged in
  `no-list-api.md` under AWS-deprecated.
- **computeoptimizer recommendation_preferences — DONE.** Iterates the
  importable `ResourceType` enum, paginates `GetRecommendationPreferences`, keys
  by `resourceType,scopeName,scopeValue`. (commit)
- **drs — DONE.** `aws_drs_replication_configuration_template` via
  `DescribeReplicationConfigurationTemplates`; SDK vendored through the
  curl→file-proxy workflow. (commit)
- **codecatalyst — SKIPPED (not cleanly importable).** `dev_environment` has no
  `terraform import`; `project` / `source_repository` import by a bare name but
  are space/project-scoped, so a generated import would not round-trip on
  refresh. Dep dropped via `go mod tidy`. Logged in `no-list-api.md`.

Net new this session: **2 resource types** (computeoptimizer prefs, drs template).
Confirms the headline: the SDK upgrade left essentially no net-new *services* to
add — the residue is deprecated, data-plane, or non-importable.
