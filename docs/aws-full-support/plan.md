# Plan: Full AWS Service & Resource Coverage for Terraformer

Goal: import every AWS resource that is **importable** — i.e. every resource the
`terraform-provider-aws` plugin supports **and** that AWS exposes a List/Describe
API for. Two work streams:

1. **Missing services** — add a generator for each service not in the registry.
2. **Missing resources** — fill gaps inside services already registered.

> **Reviewed 2026-06-20 (plan-review).** Decisions folded in below: shared
> generic helper (no per-service copy-paste), SDK retryer, declarative region
> scope table + assertion, pinned provider version for the gap diff, nil-safe
> skeleton, factored policy/allow-empty helpers, CLI-verbatim key naming,
> unit tests for the high-fan-in helpers + per-region cache. Deferred (see
> `TODOS.md`): parallel service execution, per-service memory streaming,
> `context.TODO()` timeout sweep.

## 0. Foundations to build FIRST (before any generator)

These are the shared pieces every generator depends on — build + test them
before fanning out, or the boilerplate gets copy-pasted 90× and a bug lands
everywhere.

- **`appendSimpleResources[T]` helper** (`aws_service.go`) — factors the
  page→item→`NewSimpleResource` loop so simple services are ~5 lines, not 25.
  Nil-safe (uses `aws.ToString`, skips empty import IDs). Nested/composite
  services (ecs, api_gateway) stay hand-written.
- **`wrapPolicyAttributes(g, attr, types...)` helper** — the IAM/policy heredoc +
  `escapeAwsInterpolation` logic currently inline in `ecr.go`, factored once.
  Reused by every policy-bearing service (iam, sns, sqs, s3, kms, secretsmanager…).
- **`defaultAllowEmptyValues = []string{"tags."}`** package var — replaces the
  per-file `var xAllowEmptyValues` duplication.
- **SDK retryer** in `buildBaseConfig` — adaptive retry + explicit max-attempts
  cap (e.g. 5) so throttled large imports degrade instead of aborting/hanging.
- **`serviceScope` map + assertion test** — single declarative source mapping
  each registry key to `regional | global | eastOnly`, asserted against the
  existing `SupportedGlobalResources` / `SupportedEastOnlyResources` lists so a
  misclassification fails the build, not a live import.

```
generateConfig() per-region cache + region-pass flow  (aws_service.go)
─────────────────────────────────────────────────────────────────────
provider_cmd_aws.go drives 3 SEQUENTIAL passes in ONE process:

  pass 1: aws-global ─┐
  pass 2: us-east-1   ├─ each calls Import() → services → generateConfig()
  pass 3: regions...  ─┘

generateConfig(region):
  ┌───────────────────────────────────────────────┐
  │ configCacheMu.Lock()                           │
  │ hit  configCache[region] ? → return copy ──────┼─► SDK calls use baked-in
  │ miss → buildBaseConfig (sets AWS_REGION),       │   region (correct even if
  │        retrieve creds, set creds env once,      │   global env drifts)
  │        configCache[region] = cfg                │
  └───────────────────────────────────────────────┘
  Key = region (NOT a singleton): a shared cache froze every later
  pass to pass-1's endpoint → wrong-region signing. See commit c127700b.
```

---

## 1. Hard constraints (read first — they bound "full coverage")

Terraformer is **not** an AWS-API dumper. It is bounded on two sides:

- **Terraform provider bound.** Terraformer emits HCL + tfstate, then refreshes
  each resource through the real `terraform-provider-aws` plugin
  (`terraformutils/providerwrapper`). If the provider has **no resource** for a
  thing, terraformer **cannot** import it. The provider's resource set
  (~1400 `aws_*` resources), **not** the CLI service list (433 entries), is the
  source of truth.
- **List-API bound.** Each generator's `InitResources()` must *enumerate*
  existing resources via a `List*`/`Describe*` paginator. A TF resource with no
  list/describe API (e.g. some singleton/runtime configs) cannot be
  auto-discovered.

**Excluded from scope** (no importable infrastructure — do not build generators):

- Data-plane / runtime APIs: `*-data`, `*-runtime`, `bedrock-runtime`,
  `kinesis-video-media`, `dynamodbstreams`, `cloudsearchdomain`,
  `rds-data`, `redshift-data`, `lex-runtime`, `*-events` data planes.
- CLI meta/non-service entries: `configure`, `cli-dev`, `login`, `logout`,
  `signin`, `history`, `agent-toolkit`, `nova-act`, `help`.
- Pure read/report/analytics services with no `aws_*` resource: `sts`, `pricing`,
  `ce`*(has 2 resources — see table)*, `compute-optimizer`, `health`,
  `discovery`, `savingsplans`, `support`, `trustedadvisor`, `resourcegroupstaggingapi`,
  `cloudcontrol`, `taxsettings`, `transcribe`, `translate`, `polly`,
  `rekognition`*(no TF)*, `comprehend`, `forecast`, `personalize`, `mturk`,
  `workdocs`, `mgn`, `drs`, `snowball`.

> Any service marked "verify" below must be confirmed against the provider
> schema (§3) before building — this list is best-effort from domain knowledge.

---

## 2. Current coverage snapshot

- **90** service keys registered in `AWSProvider.GetSupportedService()`
  (`providers/aws/aws_provider.go`).
- **250** distinct `aws_*` resource types emitted across `providers/aws/*.go`.

The authoritative current set is generated by the script in §3 — keep it in
`docs/aws-full-support/current-coverage.txt` so diffs are reproducible.

---

## 3. Establish the authoritative gap list (do this first, once)

Hand-maintaining a 1400-row list is error-prone. Generate it.

**a. Dump every resource terraform-provider-aws supports** (source of truth):

```bash
# requires terraform + the aws provider installed in a scratch dir
# PIN the version — do NOT use -upgrade. Terraformer refreshes against whatever
# terraform-provider-aws the USER installed at runtime (external plugin via
# ~/.terraform.d/plugins or versions.tf). If the gap list is computed against a
# newer provider than the user runs, you build generators for resources that
# fail at refresh. Pin to the declared supported floor and document it.
AWS_PROVIDER_VERSION="5.80.0"   # <- the supported floor; bump deliberately
cd $(mktemp -d)
cat > main.tf <<EOF
terraform {
  required_providers { aws = { source = "hashicorp/aws", version = "= ${AWS_PROVIDER_VERSION}" } }
}
EOF
terraform init >/dev/null
terraform providers schema -json \
  | jq -r '.provider_schemas[] | .resource_schemas | keys[]' \
  | sort > /tmp/tf-aws-all-resources.txt
wc -l /tmp/tf-aws-all-resources.txt   # ~1400
```

Record the pinned version in `docs/aws.md` as the supported provider floor so
the gap list and the user's runtime provider stay aligned.

**b. Dump what terraformer currently emits:**

```bash
cd providers/aws
grep -rhoE '"aws_[a-z0-9_]+"' *.go | tr -d '"' | sort -u \
  > docs/aws-full-support/current-coverage.txt
```

**c. The gap = (a) − (b):**

```bash
comm -23 /tmp/tf-aws-all-resources.txt \
         docs/aws-full-support/current-coverage.txt \
  > docs/aws-full-support/missing-resources.txt
```

Then bucket `missing-resources.txt` by service prefix → that bucketing drives
both work streams. A prefix entirely absent from the registry = **new service**;
a prefix partially present = **resource gap** in an existing service.

> Note: TF resource prefix ≠ terraformer registry key (e.g. `aws_appautoscaling_*`
> → would be service `application-autoscaling`; `aws_elasticsearch_domain` lives
> under registry key `es`). Map prefixes to registry keys when bucketing.

---

## 4. Gap inventory (best-effort; confirm via §3)

### 4a. Missing services that HAVE terraform-provider-aws resources (build these)

Grouped by priority. Each becomes one `providers/aws/<name>.go` + a registry entry.

**P1 — common infra, high demand**

| Service (CLI) | TF resource prefix | Notes |
|---|---|---|
| application-autoscaling | `aws_appautoscaling_*` | target, policy, scheduled_action |
| backup | `aws_backup_*` | vault, plan, selection, framework |
| dynamodb (gaps) | `aws_dynamodb_*` | global_table, kinesis_streaming_dest (resource gap, §4b) |
| datasync | `aws_datasync_*` | tasks, locations, agents |
| dlm | `aws_dlm_lifecycle_policy` | single resource |
| dms | `aws_dms_*` | replication instance/task/endpoint |
| ds (directory service) | `aws_directory_service_directory` | |
| servicediscovery | `aws_service_discovery_*` | namespaces, services |
| sesv2 | `aws_sesv2_*` | newer SES (existing `ses` covers v1 only) |
| ssm (gaps) | `aws_ssm_*` | document, association, maintenance_window, patch_baseline |
| route53resolver | `aws_route53_resolver_*` | endpoints, rules |
| globalaccelerator | `aws_globalaccelerator_*` | |
| servicequotas | `aws_servicequotas_service_quota` | |
| transfer | `aws_transfer_*` | server, user |
| fsx | `aws_fsx_*` | windows/lustre/ontap/openzfs |
| storagegateway | `aws_storagegateway_*` | |
| glacier | `aws_glacier_vault` | |
| dax | `aws_dax_*` | cluster, parameter/subnet group |

**P2 — security / governance / observability**

| Service | TF prefix | Notes |
|---|---|---|
| guardduty | `aws_guardduty_*` | detector, filter, member |
| inspector2 | `aws_inspector2_*` | + legacy `aws_inspector_*` |
| macie2 | `aws_macie2_*` | |
| fms | `aws_fms_*` | policy |
| shield | `aws_shield_*` | protection, protection_group |
| securitylake | `aws_securitylake_*` | |
| detective | `aws_detective_*` | |
| network-firewall | `aws_networkfirewall_*` | |
| ram | `aws_ram_*` | resource_share, principal/resource assoc |
| sso-admin | `aws_ssoadmin_*` | permission_set, account_assignment |
| acm-pca | `aws_acmpca_*` | certificate_authority |
| signer | `aws_signer_*` | |
| rolesanywhere | `aws_rolesanywhere_*` | |
| cur | `aws_cur_report_definition` | (us-east-1 only) |
| oam | `aws_oam_*` | sink, link |
| synthetics | `aws_synthetics_canary` | |
| rum | `aws_rum_app_monitor` | |
| internetmonitor | `aws_internetmonitor_monitor` | |
| networkmonitor | `aws_networkmonitor_*` | |

**P3 — data / analytics / ML / app platform**

| Service | TF prefix | Notes |
|---|---|---|
| athena | `aws_athena_*` | workgroup, database, named_query |
| lakeformation | `aws_lakeformation_*` | |
| neptune | `aws_neptune_*` | cluster, instance |
| memorydb | `aws_memorydb_*` | |
| timestream-write | `aws_timestreamwrite_*` | |
| timestream-influxdb | `aws_timestreaminfluxdb_db_instance` | |
| keyspaces | `aws_keyspaces_*` | |
| kinesisanalyticsv2 | `aws_kinesisanalyticsv2_application` | |
| kinesisvideo | `aws_kinesis_video_stream` | |
| emr-serverless | `aws_emrserverless_application` | |
| emr-containers | `aws_emrcontainers_virtual_cluster` | |
| mwaa | `aws_mwaa_environment` | |
| mskconnect (kafkaconnect) | `aws_mskconnect_*` | |
| opensearch | `aws_opensearch_domain` | (existing `es` = elasticsearch only) |
| opensearchserverless | `aws_opensearchserverless_*` | |
| redshift-serverless | `aws_redshiftserverless_*` | |
| docdb-elastic | `aws_docdbelastic_cluster` | |
| sagemaker | `aws_sagemaker_*` | large — model/endpoint/notebook/domain |
| quicksight | `aws_quicksight_*` | |
| pipes | `aws_pipes_pipe` | |
| scheduler | `aws_scheduler_schedule*` | |
| schemas | `aws_schemas_*` | |
| eventbridge (events gaps) | `aws_cloudwatch_event_bus/connection/api_destination` | resource gap, §4b |

**P4 — long tail (build if needed)**

`amplify`, `apprunner`, `appconfig`, `appflow`, `appmesh`, `appstream`,
`appautoscaling-plans` (`aws_autoscalingplans_scaling_plan`), `chime-sdk-voice`,
`cloudfront-keyvaluestore` (`aws_cloudfront_key_value_store`), `codeartifact`,
`codestar-connections`/`codeconnections` (`aws_codestarconnections_*`),
`codestar-notifications`, `codegurureviewer`, `codeguruprofiler`, `connect`,
`controltower`, `customer-profiles`, `dataexchange`, `datazone`, `deadline`,
`devopsguru`, `finspace`, `fis`, `gamelift`, `grafana`, `greengrassv2`,
`groundstation`, `healthlake`, `imagebuilder`, `iotanalytics`/`iotevents`/
`iotsitewise`/`iottwinmaker`/`iotwireless`, `ivs`, `kendra`, `lexv2-models`,
`license-manager`, `lightsail`, `location`, `m2`, `mediaconvert`,
`mediapackagev2`, `neptune-graph`, `osis`, `payment-cryptography`,
`pca-connector-ad`, `pcs`, `pinpoint`, `proton`, `qbusiness`, `rbin`,
`resiliencehub`, `resource-explorer-2`, `route53-recovery-*`, `route53domains`,
`route53profiles`, `s3control`, `s3outposts`, `s3tables`, `securityhub` (gaps),
`serverlessrepo`, `servicecatalog-appregistry`, `ssm-contacts`, `ssm-incidents`,
`ssm-sap`, `verifiedpermissions`, `vpc-lattice`, `wellarchitected`,
`workspaces-web`, `xray` (gaps).

### 4b. Partial services — known resource gaps (fill these)

| Registry key | Has | Missing (examples) |
|---|---|---|
| `codedeploy` | app | `aws_codedeploy_deployment_group`, `_deployment_config` |
| `servicecatalog` | portfolio | `_product`, `_constraint`, `_principal_portfolio_association`, `_product_portfolio_association` |
| `ssm` | parameter | document, association, maintenance_window(+task/target), patch_baseline, patch_group, resource_data_sync |
| `cloudwatch`/events | rule, target, alarm, dashboard, log_group | `aws_cloudwatch_event_bus`, `_connection`, `_api_destination`, `_log_metric_filter`, `_log_resource_policy`, `_composite_alarm` |
| `ses` | v1 identities/rules/templates | `aws_sesv2_*` (separate service) |
| `ec2`/`ec2_instance` | instance, ebs, eni, eip, vpc, subnet, sg, route_table, nacl, igw, nat, tgw, vpn, vpc_endpoint/peering, customer_gateway | `aws_ami`, `aws_key_pair`, `aws_placement_group`, `aws_flow_log`, `aws_spot_*`, `aws_ec2_fleet`, `aws_ec2_managed_prefix_list`, `aws_vpc_dhcp_options`, `aws_egress_only_internet_gateway`, `aws_default_*` |
| `route53` | zone, record, health_check | `aws_route53_delegation_set`, `_query_log`, `_vpc_association_authorization`, `_resolver_*` (→ new service) |
| `s3` | bucket, bucket_policy | bucket sub-resources now split out: `aws_s3_bucket_versioning`, `_lifecycle_configuration`, `_server_side_encryption_configuration`, `_public_access_block`, `_cors_configuration`, `_logging`, `_acl`, `_notification`, `_website_configuration`, `_ownership_controls`, `_object_lock_configuration`, `_replication_configuration`, `aws_s3_access_point` |
| `cognito` | user_pool(+client), identity_pool | `aws_cognito_user_pool_domain`, `_resource_server`, `_user_group`, `_identity_provider`, `_risk_configuration`, `_identity_pool_roles_attachment` |
| `iam` | user/group/role/policy + attachments, instance_profile, access_key | `aws_iam_account_password_policy`, `_account_alias`, `_openid_connect_provider`, `_saml_provider`, `_server_certificate`, `_service_linked_role`, `_signing_certificate`, `_virtual_mfa_device` |
| `elasticache` | cluster, replication_group, param/subnet group | `aws_elasticache_user`, `_user_group`, `_global_replication_group` |
| `dynamodb` | table | `aws_dynamodb_global_table`, `_kinesis_streaming_destination`, `_contributor_insights`, `_table_replica` |
| `kinesis` | stream | `aws_kinesis_stream_consumer` |
| `sns` | topic, subscription | `aws_sns_topic_policy`, `_sns_platform_application`, `_sns_sms_preferences` |
| `sqs` | queue | `aws_sqs_queue_policy`, `_redrive_policy/allow` |
| `lambda` | function, layer, perm, esm, invoke_config | `aws_lambda_alias`, `_function_url`, `_provisioned_concurrency_config`, `_code_signing_config` |
| `cloudfront` | distribution, cache_policy | `aws_cloudfront_origin_access_control/identity`, `_function`, `_response_headers_policy`, `_origin_request_policy`, `_key_group`, `_field_level_encryption_*`, `_monitoring_subscription` |
| `kms` | key, alias, grant | `aws_kms_external_key`, `_replica_key`, `_custom_key_store` |
| `apigatewayv2` | api, authorizer, model, route(+resp), vpc_link | `aws_apigatewayv2_integration`, `_stage`, `_deployment`, `_domain_name`, `_api_mapping` |
| `securityhub` | account, member, standards_subscription | `aws_securityhub_action_target`, `_insight`, `_finding_aggregator`, `_product_subscription`, `_organization_*` |
| `eks` | cluster, node_group | `aws_eks_fargate_profile`, `_addon`, `_identity_provider_config`, `_access_entry` |
| `ecs` | cluster, service, task_definition | `aws_ecs_capacity_provider`, `_cluster_capacity_providers`, `_account_setting_default`, `_tag` |
| `redshift` | cluster, subnet/param group, snapshot schedule, event_sub | `aws_redshift_endpoint_access`, `_scheduled_action`, `_usage_limit`, `_authentication_profile`, `_hsm_*` |
| `glue` | catalog db/table, crawler, job, trigger | `aws_glue_connection`, `_workflow`, `_registry`, `_schema`, `_partition`, `_security_configuration`, `_user_defined_function`, `_data_catalog_encryption_settings` |

> The §3 diff produces the *complete* per-service gap; the table above is the
> high-value starter set, not exhaustive.

---

## 5. Recipe — add a NEW service generator

Reference implementation: `providers/aws/ecr.go` (simple) and
`providers/aws/cloudwatch.go` (multi-resource). Steps:

1. **Create `providers/aws/<service>.go`** using the shared helper from §0 — do
   NOT hand-roll the page loop:

```go
package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/<sdkpkg>"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type <Svc>Generator struct {
	AWSService
}

func (g *<Svc>Generator) InitResources() error {
	config, e := g.generateConfig()      // region-aware SDK config (see §0 diagram)
	if e != nil {
		return e
	}
	svc := <sdkpkg>.NewFromConfig(config)

	p := <sdkpkg>.NewList<Thing>sPaginator(svc, &<sdkpkg>.List<Thing>sInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())  // ponytail: TODO ctx repo-wide, see TODOS.md T3
		if err != nil {
			return err
		}
		for _, item := range page.<Things> {
			id := aws.ToString(item.<Id>)         // nil-safe: never *item.X
			if id == "" {
				continue                          // skip un-importable items
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				aws.ToString(item.<Name>),        // terraform resource name (sanitized)
				"aws_<tf_type>",
				"aws",
				defaultAllowEmptyValues))         // shared var from §0, not a per-file slice
		}
	}
	return nil
}
```

> The §0 `appendSimpleResources[T]` helper collapses the inner two loops to one
> call for the common case; use it. Keep the explicit loop only when items need
> per-item branching (extra API calls, conditional resource types — see `ecr.go`).

2. **Register it** in `aws_provider.go` → `GetSupportedService()`:

```go
"<service>": &AwsFacade{service: &<Svc>Generator{}},
```

   **Key naming convention:** new keys = the **AWS CLI service name verbatim**
   (what users know from the CLI list), e.g. `application-autoscaling`, `backup`,
   `route53resolver`. Legacy keys that don't follow this (`auto_scaling`,
   `ec2_instance`, `api_gateway`) stay for back-compat — document, don't churn.

3. **Region classification** — add the key to the `serviceScope` map (§0) with
   `regional` / `global` / `eastOnly`; the assertion test keeps it in sync with
   `SupportedGlobalResources` / `SupportedEastOnlyResources`. Default `regional`
   still needs an explicit entry so the assertion can prove completeness.

4. **`PostConvertHook()`** (optional): for policy-JSON resources call the §0
   `wrapPolicyAttributes` helper instead of re-implementing the heredoc/escape;
   otherwise use it to drop computed-only fields or fix interpolation.

5. **Docs**: append the service key to `docs/aws.md`.

6. **Tests** (per §0 + `sg_test.go` convention): if `InitResources` does any
   non-trivial result processing, extract it into a pure function taking the
   already-fetched SDK structs and unit-test that (see `findSgsToMoveOut` in
   `sg_test.go`). A trivial list→`NewSimpleResource` mapper needs no test. Do
   **not** mock the SDK client — the codebase doesn't, and it's not worth it.

**Per-resource gotchas:**

- **Never `*item.X`** — AWS SDK fields are pointers; use `aws.ToString` /
  `aws.ToInt32` etc. A nil optional field deref panics and kills the whole run.
- The **import ID** (1st arg) must match what `terraform import aws_x.y <ID>`
  expects — check the provider docs' "Import" section per resource; it is often
  not the same as Name/ARN.
- Sub-resources that need a parent ID (e.g. ECS service needs cluster) require
  the parent loop to nest the child enumeration — see `ecs.go`, `api_gateway.go`.
- Composite import IDs use `terraformutils.NewResource(...)` with explicit
  attributes instead of `NewSimpleResource`.

## 6. Recipe — expand an EXISTING service (resource gaps)

1. From §3's per-service gap, list missing `aws_*` types for that file.
2. Add an enumeration block per missing type inside the existing
   `InitResources()` (or split into a helper). Reuse the already-built `svc`
   client when same SDK package; create a new client otherwise.
3. Add to that service's `PostConvertHook()` if needed.
4. Update the test data / canary if the service has one.

---

## 7. Phased rollout

- **Phase 0a — foundations (§0):** build + unit-test the shared helpers
  (`appendSimpleResources`, `wrapPolicyAttributes`, `defaultAllowEmptyValues`),
  the SDK retryer, and the `serviceScope` map + assertion test. Plus the
  per-region config-cache unit test. **Nothing fans out before this lands** —
  it's what stops the boilerplate (and its bugs) from being copied 90×.
- **Phase 0b — tooling (1 day):** run §3 (pinned version), commit
  `current-coverage.txt`, `tf-aws-all-resources.txt`, `missing-resources.txt`;
  bucket by service. Produces the *real* backlog (this doc is the map; §3 is the
  territory).
- **Phase 1 — partial-service gaps (§4b):** highest ROI, services already wired.
  Start with `s3`, `ec2`, `iam`, `cognito`, `ssm`, `cloudwatch/events`,
  `lambda`, `cloudfront`.
- **Phase 2 — P1 missing services (§4a):** application-autoscaling, backup,
  dms, datasync, dlm, ds, sesv2, route53resolver, transfer, fsx, dax,
  globalaccelerator, servicediscovery, storagegateway, glacier, servicequotas.
- **Phase 3 — P2 security/governance:** guardduty, inspector2, macie2, fms,
  shield, securitylake, detective, network-firewall, ram, sso-admin, acm-pca,
  signer, oam, synthetics, rum.
- **Phase 4 — P3 data/ML/app:** athena, lakeformation, neptune, memorydb,
  timestream*, opensearch(+serverless), sagemaker, quicksight, pipes, scheduler,
  mwaa, emr-serverless/containers, redshift-serverless, kinesisanalyticsv2.
- **Phase 5 — P4 long tail:** build on demand.

Each service = one focused PR (generator + registry + docs + region class).
Keeps reviews small and isolates SDK-version bumps.

---

## 8. Validation per service

Test strategy (matches the codebase — `sg_test.go` is the only generator test
and it tests a *pure transform*, not a mocked SDK):

- **Unit (required):** the §0 shared helpers — `appendSimpleResources` (nil ID
  skipped, empty page, mapping), `wrapPolicyAttributes` (missing attr, non-string
  value, interpolation escaping), the `serviceScope` assertion (every key
  classified + consistent with the global/east lists), and the per-region config
  cache (same region → same config, different regions → distinct). These are the
  high-fan-in pieces; a bug here is a bug in all 90 generators.
- **Unit (per generator, conditional):** only if `InitResources` has non-trivial
  result processing — extract it to a pure func and test it (`sg_test.go`
  pattern). Trivial mappers need none. No SDK mocking.
- **Integration (the correctness bar):**
  `go build -o terraformer && ./terraformer import aws --resources=<svc> --regions=<r>`
  against a test account, then `terraform plan` on the output shows **no diff**
  (refresh round-trips cleanly). Use the provider version pinned in §3.
- `go test ./providers/aws/...` (full suite pulls the provider plugin via
  `providerwrapper`; needs module cache + network).
- `golangci-lint run` (gocritic/revive/unconvert/unparam enabled).
- The `serviceScope` assertion replaces eyeballing the region passes — if a new
  global/east-only service is misclassified, the test fails instead of a live
  import going to the wrong path / failing to sign.

---

## 9. Effort estimate

- Partial-service gaps (§4b): ~25 services touched, ~120 resource blocks.
- New services with TF resources (§4a): ~90 services (P1–P4).
- Total new `aws_*` resource types reachable: the §3 diff gives the exact count
  (expect the gap to be **800–1000** of the provider's ~1400, since terraformer
  currently covers ~250 and some provider resources have no list API).

Track progress by re-running §3 after each PR and shrinking
`missing-resources.txt`. Done = diff is empty modulo the documented
"no list API" exclusions.

---

## 10. Performance & scale notes

Service init runs **sequentially** (the `WaitGroup` at `cmd/import.go:155` is
dead code — no goroutine/`Wait`). A full `--resources=*` multi-region import =
Σ(every service's serial paginated calls) × every region, which at full coverage
can run **hours**.

- **Mitigation now:** recommend users scope `--resources` to what they need;
  the SDK retryer (§0) is capped (adaptive + max ~5 attempts) so throttling
  degrades instead of hanging unboundedly.
- **Deferred (see `TODOS.md`):** parallel service execution (T1 — needs
  thread-safe `ProvidersMapping` + env first; **not** filed as a TODO by choice,
  noted here only), per-service memory streaming (T2 — all refreshed state is
  held in memory before write; reachable OOM on large accounts), and the
  `context.TODO()` timeout sweep (T3 — one hung call hangs the run).
- The per-region config cache is already optimal for the read path: first service
  in a region builds the config, the rest hit the cache (no repeated
  `LoadDefaultConfig`/STS per service).

## 11. Open decisions

- **Un-listable resources:** maintain an explicit `docs/aws-full-support/no-list-api.md`
  exclusion list so "missing" never silently means "impossible".
- **SDK migration:** existing code is on `aws-sdk-go-v2`; all new generators use
  v2 paginators. Don't reintroduce v1.
- **Volume vs. value:** P4 long-tail services (ML/IoT/media niche) may not be
  worth maintenance cost — decide per service before building.
- **Generic lister not used:** AWS Cloud Control / Resource Explorer could list
  generically but couple to a different discovery model + import-ID scheme;
  deliberately out of scope. Revisit only if hand-written list APIs stall.
