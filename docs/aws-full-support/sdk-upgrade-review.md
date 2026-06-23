# AWS coverage review after the SDK upgrade (2026-06-23)

Two passes. **Pass 1** diffed the committed schema dump (`tf-aws-all-resources.txt`,
1453 rows) and concluded "net-new services essentially complete." **Pass 2** went
to the authoritative live source and found the dump was **stale** — so pass 1
understated the gap. This doc supersedes pass 1's conclusion.

## Method (pass 2 — authoritative)

- Live resource universe = filenames under `website/docs/r/` in
  `hashicorp/terraform-provider-aws@main` (exact resource names, fetched via the
  GitHub tree API): **1669** `aws_*` resources.
- Service registry = `names/data/names_data.hcl`: **366** services.
- Terraformer emits **1348** distinct `aws_*` types across **225/366** services.
- SDK List/Describe ops verified against the **go.mod-pinned** SDK version (not
  whatever's first in the module cache — that bug made an earlier check report
  false negatives on cloudfront/appsync/redshift/securityhub).

**Gap vs live provider = 352** (not 184). The committed dump missed **236**
resources the provider has added since it was generated.

## Net-new top-level services: 0 addable

9 services terraformer covers 0 of — all correctly excluded:

| Service | Why excluded |
|---|---|
| lexmodels (lex v1) | Superseded by `aws_lexv2models_*` (built) |
| codecatalyst | Not cleanly reverse-importable (dev_environment no import; project/source_repository import by bare name but are space/project-scoped) |
| cloudcontrol (`aws_cloudcontrolapi_resource`) | Generic meta-resource |
| redshiftdata (`aws_redshiftdata_statement`) | Data-plane (runs SQL) |
| cloudfrontkeyvaluestore (`_key`, `_keys_exclusive`) | Data-plane KV items in a store (container already built) |
| elb (`aws_elb_attachment`) | Structural attachment (elb built) |
| serverlessrepo (`_cloudformation_stack`) | Deploy action, not durable infra |
| outposts (`aws_outposts_capacity_task`) | Action resource |

The headline holds for *services* — but the SDK upgrade opened a real backlog of
**new child-resources on services we already cover**.

## Addable now — SDK already vendored, no dep bump (~30)

Each verified to have a List/Describe op in the go.mod-pinned SDK. Standard
`List → NewSimpleResource` generators; add to the existing service file.

| Resource | Service file | List/Describe op |
|---|---|---|
| `aws_cloudfront_vpc_origin` | cloudfront | `ListVpcOrigins` |
| `aws_cloudfront_anycast_ip_list` | cloudfront | `ListAnycastIpLists` |
| `aws_cloudfront_distribution_tenant` | cloudfront | `ListDistributionTenants` |
| `aws_cloudfront_connection_group` | cloudfront | `ListConnectionGroups` |
| `aws_cloudfront_trust_store` | cloudfront | `ListTrustStores` |
| `aws_bedrock_inference_profile` | bedrock | `ListInferenceProfiles` |
| `aws_bedrockagent_flow` | bedrockagent | `ListFlows` |
| `aws_bedrockagent_prompt` | bedrockagent | `ListPrompts` |
| `aws_appsync_api` | appsync | `ListApis` |
| `aws_appsync_channel_namespace` | appsync | `ListChannelNamespaces` (per-api) |
| `aws_athena_capacity_reservation` | athena | `ListCapacityReservations` |
| `aws_opensearch_application` | opensearch | `ListApplications` |
| `aws_opensearchserverless_collection_group`* | opensearchserverless | verify (`ListCollections` is the collection, not group) |
| `aws_memorydb_multi_region_cluster` | memorydb | `DescribeMultiRegionClusters` |
| `aws_timestreaminfluxdb_db_cluster` | timestreaminfluxdb | `ListDbClusters` |
| `aws_vpclattice_resource_gateway` | vpclattice | `ListResourceGateways` |
| `aws_wafv2_api_key` | wafv2 | `ListAPIKeys` (per-scope) |
| `aws_workmail_domain` | workmail | `ListMailDomains` (per-org) |
| `aws_cloudwatch_log_anomaly_detector` | logs | `ListLogAnomalyDetectors` |
| `aws_cloudwatch_log_delivery` | logs | `DescribeDeliveries` |
| `aws_cloudwatch_log_delivery_destination` | logs | `DescribeDeliveryDestinations` |
| `aws_cloudwatch_log_delivery_source` | logs | `DescribeDeliverySources` |
| `aws_cloudwatch_contributor_insight_rule` | cloudwatch | `DescribeInsightRules` |
| `aws_lakeformation_lf_tag_expression` | lakeformation | `ListLFTagExpressions` |
| `aws_rds_shard_group` | rds | `DescribeDBShardGroups` |
| `aws_redshift_integration` | redshift | `DescribeIntegrations` |
| `aws_route53domains_domain` | route53domains | `ListDomains` |
| `aws_sagemaker_model_card` | sagemaker | `ListModelCards` |
| `aws_sagemaker_algorithm` | sagemaker | `ListAlgorithms` |
| `aws_sagemaker_mlflow_app` | sagemaker | `ListMlflowTrackingServers` |
| `aws_transfer_web_app` | transfer | `ListWebApps` |
| `aws_transfer_host_key` | transfer | `ListHostKeys` (per-server) |
| `aws_glue_catalog` | glue | `GetCatalogs` |
| `aws_securityhub_automation_rule_v2` | securityhub | `ListAutomationRulesV2` |

\* opensearchserverless_collection_group needs the right op confirmed before build.

**Add with care (filter required, else noise):**
- `aws_rds_custom_db_engine_version` — `DescribeDBEngineVersions` returns all AWS
  engine versions; filter to customer-owned (`CustomEngineVersion`) only.
- `aws_elastictranscoder_preset` — `ListPresets` returns AWS system presets too;
  filter to non-system.

**Per-parent enumeration (multi-step, lower priority):** appsync_channel_namespace
(per api), wafv2_api_key (per scope), workmail_domain (per org), transfer_host_key
(per server) — need a parent loop like the evidently pattern.

## Excluded (the remaining ~290 of the 352 gap)

Documented classes in [no-list-api.md](no-list-api.md):
- **Structural suffixes (~136):** `*_association` (31), `*_policy` (22),
  `*_accepter` (16), `*_attachment` (12), `*_exclusive` (iam/route53/ram/ssoadmin/
  vpc_security_group_rules — conflict with parent, double-manage), `*_tag`,
  `*_permission`, `*_membership`, `*_default_*`.
- **Data-plane / action / no-import:** kms_ciphertext, lambda_invocation,
  iot_logging_options, secretsmanager_secret_version, dataexchange_revision*,
  customerprofiles_profile, msk_topic, ebs_fast_snapshot_restore,
  vpc_ipam_preview_next_cidr, quicksight_ingestion, grafana_workspace_api_key,
  iam_user_login_profile, snapshot/ami copy+permission.
- **Transient job executions (not durable infra):** sagemaker_training_job,
  sagemaker_labeling_job, sagemaker_hyper_parameter_tuning_job,
  sagemaker_model_card_export_job.
- **Two-account handshake:** dx_hosted_*, dx_bgp_peer, dx_connection_confirmation,
  directory_service_shared_directory_accepter.
- **Superseded:** kinesis_analytics_application (v1; v2 built), lex v1.
- **Org-level singletons (low value):** organizations_aws_service_access,
  ram_sharing_with_organization, servicecatalog_organizations_access,
  vpc_ipam_organization_admin_account, cloudtrail_organization_delegated_admin_account.

## Status of this session's edits

- `computeoptimizer_recommendation_preferences` — DONE (committed).
- `drs_replication_configuration_template` — DONE (committed, dep vendored).
- evidently feature/launch/segment — DROPPED (SDK marks service AWS-EOL).
- codecatalyst — SKIPPED (not cleanly importable).

## Pass 3 — backlog implemented (2026-06-23)

The "addable now" backlog was built out. Emitted `aws_*` types **1346 → 1375**.
One commit per service file:

| Commit | Resources added |
|---|---|
| cloudfront | vpc_origin, anycast_ip_list, distribution_tenant, connection_group, trust_store |
| logs | log_anomaly_detector, log_delivery, log_delivery_destination, log_delivery_source |
| cloudwatch | contributor_insight_rule |
| bedrock/bedrockagent/appsync/athena | bedrock_inference_profile (application-only), bedrockagent_flow, bedrockagent_prompt, appsync_api, athena_capacity_reservation |
| opensearch/memorydb/timestreaminfluxdb/vpclattice | opensearch_application, memorydb_multi_region_cluster, timestreaminfluxdb_db_cluster, vpclattice_resource_gateway |
| lakeformation/rds/redshift | lakeformation_lf_tag_expression, rds_shard_group, redshift_integration |
| sagemaker/transfer/glue/securityhub | sagemaker_model_card, sagemaker_algorithm, transfer_web_app, glue_catalog, securityhub_automation_rule_v2 |

**Dropped during build (with reason):**
- `route53domains_domain` — would double-emit domains already covered by
  `aws_route53domains_registered_domain` (same `ListDomains`).
- `sagemaker_mlflow_app` — no matching list op (`ListMlflowTrackingServers` is a
  different resource).
- `opensearchserverless_collection_group` — list op for the *group* not confirmed.
- `rds_custom_db_engine_version`, `elastictranscoder_preset` — deferred (need a
  customer-owned/system filter to avoid emitting AWS-managed entries).
- Per-parent leaves (appsync_channel_namespace, wafv2_api_key, workmail_domain,
  transfer_host_key) — deferred; need a parent loop, lower value.

All builds + `TestServiceScope`/`TestEveryServiceDocumented`/
`TestAllServicesInstantiable` green; gofmt clean.

## Recommendation

The ~30 "addable now" rows are the real post-upgrade backlog: clean
`List→NewSimpleResource` adds on already-registered services, zero new deps. Ship
in batches by service file (cloudfront, cloudwatchlogs, sagemaker, bedrock(agent),
appsync, rds/redshift, transfer, misc), one commit per service, refreshing
`docs/aws.md`. Verify each resource's import-ID format against the TF docs before
building (some are composite, e.g. `name:parent`). Regenerate
`tf-aws-all-resources.txt` (it is ~236 resources stale) when the schema dump can
run outside the sandbox.
