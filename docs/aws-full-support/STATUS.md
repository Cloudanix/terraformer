# AWS Full-Support — Implementation Status

Tracks progress against [plan.md](plan.md). Updated 2026-06-22.

## Done

### Foundations (§0) — complete
Shared helpers (`appendSimpleResources`, `wrapPolicyAttribute`,
`defaultAllowEmptyValues`), SDK retryer, per-region config cache, and the
`serviceScope` map + assertion test all landed earlier.

### New services (§4a)
**P1:** application-autoscaling, backup, datasync, dax, directory-service (ds),
dlm, dms, fsx, glacier, **globalaccelerator** (partition-global),
route53resolver, servicediscovery, **servicequotas** (T5, change-history
filter), sesv2, **sso-admin**, storagegateway, transfer.

**P2/P3 (added via the module-fetch workaround, see below):** guardduty,
macie2, **shield** (partition-global), fms, detective, ram, acm-pca, signer,
inspector2, synthetics, rum, oam, networkmonitor, internetmonitor,
network-firewall, rolesanywhere, securitylake, athena, opensearch, memorydb,
neptune.

### Network workaround (the original blocker is solved)
Go's own module download fails in the sandbox (TLS), **but `curl` reaches
`proxy.golang.org` fine.** So new SDK modules are fetched via curl into a local
file-GOPROXY and built with a relocated, writable module cache:

```sh
# fetch a module:  $TMPDIR/getmod.sh <module-path>   (resolves @latest + curls .info/.mod/.zip)
export GOMODCACHE="$TMPDIR/gomodcache" GOCACHE="$TMPDIR/gocache"
export GOPROXY="file://$TMPDIR/goproxy,file:///Users/puru/go/pkg/mod/cache/download"
export GOSUMDB=off GOFLAGS=-mod=mod
go get <module>@<version>   # version must be explicit (file proxy has no @latest)
```

The existing on-disk download cache (`.../pkg/mod/cache/download`) is chained as
a second file-proxy so all already-cached deps resolve. This unblocks every
remaining §4a service — fetch its module, then follow the §5 recipe.

### Partial-service gaps (§4b) — complete except as noted
| Service | Added |
|---|---|
| iam | account_alias, account_password_policy, openid_connect_provider, saml_provider, server_certificate, virtual_mfa_device |
| lambda | alias, function_url, code_signing_config |
| cloudfront | function, key_group, origin_access_control, origin_access_identity, origin_request_policy, response_headers_policy |
| ec2 | ami, key_pair, placement_group, flow_log, ec2_managed_prefix_list, vpc_dhcp_options, egress_only_internet_gateway |
| eks | addon, fargate_profile, identity_provider_config |
| ecs | capacity_provider |
| kinesis | stream_consumer |
| route53 | delegation_set, query_log |
| glue | connection, workflow, security_configuration, registry |
| elasticache | user, user_group |
| codedeploy | deployment_group, deployment_config |
| securityhub | action_target, insight, finding_aggregator |
| redshift | scheduled_action, usage_limit, authentication_profile, endpoint_access |
| servicecatalog | product |
| cloudwatch | composite_alarm, event_bus, event_connection, event_api_destination |
| sns | platform_application |
| dynamodb | global_table |
| kms | custom_key_store |
| ssm | document, maintenance_window, patch_baseline, association |
| cognito | user_group, resource_server, identity_provider |
| s3 | versioning, lifecycle, SSE, public_access_block, cors, logging, website, ownership_controls, object_lock, replication, accelerate, notification (probed; emitted only where configured) |

Every change: one focused commit + docs/aws.md + serviceScope entry where new.
`go build ./...` and `go test ./...` green.

## Milestone: all net-new services registered
Every AWS service in the provider that (a) has a terraform-provider-aws resource
and (b) exposes a List/Describe API is now registered — **234 services**. The
remaining `missing-resources.txt` entries are exclusively (1) sub-resources of
already-registered services (deepening work) or (2) documented exclusions in
`no-list-api.md` (no independent list API, data-plane, deprecated, aliases, or
account singletons). No buildable net-new *service* remains.

### Plan §4 completeness audit (verified 2026-06-22)
Mechanical check of every service/resource named in plan.md §4a + §4b against
the live registry (`aws_provider.go`) and `current-coverage.txt`:

- **§4a P1/P2/P3 services (56)** — all registered: application-autoscaling,
  backup, datasync, dlm, dms, ds, servicediscovery, sesv2, route53resolver,
  globalaccelerator, servicequotas, transfer, fsx, storagegateway, glacier, dax,
  guardduty, inspector2, macie2, fms, shield, securitylake, detective,
  network-firewall, ram, sso-admin, acm-pca, signer, rolesanywhere, oam,
  synthetics, rum, internetmonitor, networkmonitor, athena, lakeformation,
  neptune, memorydb, timestream-write, timestream-influxdb, keyspaces,
  kinesisanalyticsv2, kinesisvideo, emr-serverless, emr-containers, mwaa,
  kafkaconnect, opensearch, opensearchserverless, redshift-serverless,
  docdb-elastic, sagemaker, quicksight, pipes, scheduler, schemas. **0 missing.**
- **§4a P4 long-tail services (55)** — all registered: amplify, apprunner,
  appconfig, appflow, appmesh, appstream, chime-sdk-voice, codeartifact, connect,
  controltower, dataexchange, datazone, finspace, fis, gamelift, grafana,
  greengrassv2, groundstation, healthlake, imagebuilder, iotevents, iotsitewise,
  iottwinmaker, ivs, kendra, lexv2-models, license-manager, lightsail, location,
  m2, mediaconvert, mediapackagev2, neptune-graph, osis, pca-connector-ad, pcs,
  pinpoint, proton, qbusiness, rbin, resiliencehub, resource-explorer-2,
  route53domains, route53profiles, s3control, s3outposts, s3tables,
  servicecatalog-appregistry, ssm-contacts, ssm-incidents, verifiedpermissions,
  vpc-lattice, wellarchitected, workspaces-web, xray. **0 missing.**
- **§4b partial-service gaps** — all 100 named gap resource types
  (codedeploy/servicecatalog/ssm/cloudwatch+events/ec2/route53/s3/cognito/iam/
  elasticache/dynamodb/kinesis/sns/lambda/cloudfront/kms/apigatewayv2/securityhub/
  eks/ecs/redshift/glue …) present in `current-coverage.txt`. **0 missing.**

Both §4 work streams are therefore complete; coverage 250→1232 was reached
cumulatively across the rollout (§7 phases 0–5). What this last session added is
the §9 deepening tail — per-parent associations/policies enumerable via List/Get
on already-registered services — not net-new services, which were already done.

## Coverage tally

Current coverage: 234 services / **1239** `aws_*` resource types (baseline 90 / 250). §3 gap
`missing-resources.txt` = **232** — the clean-import buildable tail is now empty
(incl. the high-cardinality per-parent leaves §9 named: glue_partition,
cloudwatch_log_stream, sagemaker_device, glue_catalog_table_optimizer);
every remaining entry maps to a documented exclusion class in
[no-list-api.md](no-list-api.md) (structural attachment/policy/accepter/tag/
exclusive suffixes, unvendored SDKs, data-plane/no-import objects, and
`aws_default_*` adopt-existing resources). Historically:
~215 structurally-unlistable suffixes (association/attachment/policy/accepter/
settings), unvendored-SDK resources (evidently, lex v1, devopsguru, drs,
eventbridge endpoints, paymentcryptography, simpledb), and data-plane/no-import
objects — all enumerated in [no-list-api.md](no-list-api.md). The residual
buildable items are high-cardinality per-parent leaves (e.g. glue partitions)
and two-account handshake resources.

**234 services registered** in `GetSupportedService()` (was 90 at the start of
this effort). All §4b partial-service gaps done; §4a P1/P2/P3 done; §4a P4 done
across 6 batches. The §4b deepening tail is being ground down service-by-service
(coverage 706→893 this session): each batch adds the sub-resources / per-parent
children / account singletons exposed by an already-registered service's
SDK List/Describe/Get APIs, with `docs/aws.md` and the gap artifacts refreshed
and a separate commit per service. `go build ./providers/aws/` and the
cross-cutting tests (`TestScope`, `TestEveryServiceDocumented`,
`TestAllServicesInstantiable`) stay green; `serviceScope` assertion passes;
`gofmt`/`golangci-lint` clean on changed lines.

## Remaining / deferred

Special-handling services now BUILT: quicksight, s3control (account-id scoped),
mediaconvert (DescribeEndpoints), s3outposts, cloudfront key_value_store.

Genuinely excluded (no TF resource / deprecated / multi-step control plane) are
catalogued in [no-list-api.md](no-list-api.md): deadline, iotwireless,
devopsguru, serverlessrepo, ssm-sap, iotanalytics (AWS-deprecated, removed), and
route53-recovery-* (revisit on demand).

apigatewayv2 is now rewritten to aws-sdk-go-v2 (no more v1 in providers/aws) with
the §4b gaps (integration, stage, deployment, domain_name, api_mapping) added.
Still open: sqs/sns `*_policy` (conflict with parent inline policy).

### §3 gap list — current-coverage.txt committed; schema dump sandbox-blocked
`docs/aws-full-support/gen-gap-list.sh` runs the full §3 (with a curl-populated
filesystem mirror so `terraform init` works without registry TLS). `current-
coverage.txt` (523 resource types) is committed. `tf-aws-all-resources.txt` /
`missing-resources.txt` need `terraform providers schema`, which launches the
provider plugin over go-plugin. Proven blocked at the syscall level in this
sandbox — the plugin logs:

    plugin init error: listen unix /tmp/.../plugin…: bind: operation not permitted

i.e. the sandbox denies the unix-socket `bind` go-plugin requires. Confirmed
the denial is at the `bind(2)` syscall, universally (not a path/permission
issue): binding a unix socket in the repo dir, `$TMPDIR`, and `/Users/puru/code`
ALL fail `operation not permitted`, and TCP loopback `bind(127.0.0.1:0)` also
fails — so there is no transport go-plugin could use here. This is the SAME wall
as the integration `terraform plan` round-trip and any real `terraformer
import`. Run `gen-gap-list.sh` in a normal environment to produce the diff.

## Tests (cover every service + the logic-bearing/foundation code)
- `TestAllServicesInstantiable` — every registered service's facade+generator is
  non-nil and accepts the cmd/import.go wiring (per-service smoke test, all 234).
- `TestServiceScopeMatchesRegistry` — every service region-classified+consistent.
- `TestEveryServiceDocumented` — every service present in docs/aws.md.
- `TestQuotasFromChangeHistory` + §0 helper tests (`TestAppendSimpleResources*`,
  `TestWrapPolicyAttribute*`, `TestGenerateConfigCache*`) + `sg_test`.
The live List/Describe path of each generator is not unit-tested: the codebase
deliberately does not mock the AWS SDK, and the provider-plugin alternative is
the bind-blocked round-trip above.

## Tests
- `TestServiceScope` — every registered service has a consistent region scope.
- `TestEveryServiceDocumented` — every registered service appears in docs/aws.md.
- `TestQuotasFromChangeHistory` — servicequotas dedup/filter pure function.
Per the codebase convention (no SDK mocking; `sg_test.go` tests a pure
transform), trivial list→resource generators rely on the two cross-cutting
assertions above plus the integration round-trip (import → `terraform plan`).

### §3 authoritative gap list — runnable now
`terraform init` + `terraform providers schema -json` can run via the same
network (curl works). Not yet executed; run §3 to produce the exact remaining
per-resource diff.

### Intentionally not built
- **apigatewayv2** gaps (integration, stage, deployment, domain_name,
  api_mapping): the existing `api_gatewayv2.go` is on **aws-sdk-go v1**;
  extending it would reintroduce v1 (plan §11 forbids). Needs a v2 rewrite first.
- **sqs queue_policy / sns topic_policy**: conflict with the parent's inline
  `policy` attribute — would double-manage. Left out deliberately.
- **kms external_key/replica_key**, dynamodb table_replica/contributor_insights,
  servicecatalog associations: need per-item Describe or composite IDs; lower
  value, deferred.

### Other deferred (TODOS.md)
T2 (memory streaming), T3 (context timeout sweep), T4 (fixed-region scope —
globalaccelerator shipped as partition-global instead).
