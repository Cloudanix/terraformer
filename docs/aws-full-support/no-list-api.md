# Deliberately excluded AWS services / resources

Per plan.md §1 and §11: terraformer can only import a resource that (a) has a
`terraform-provider-aws` resource AND (b) exposes a List/Describe API to
enumerate existing instances. This file records services/resources intentionally
**not** built, so "missing from coverage" never silently means "impossible".

## No terraform-provider-aws resource (provider bound, §1)
The service exists in the AWS SDK but the Terraform AWS provider has no `aws_*`
resource to import into, so a generator would emit un-refreshable HCL:

- **deadline** — AWS Deadline Cloud; no `aws_deadline_*` resources.
- **iotwireless** — no `aws_iotwireless_*` resources.
- **devopsguru** — account-level configuration only; no enumerable `aws_*` resource.
- **ssm-sap** — no `aws_ssmsap_*` resources.
- **serverlessrepo** — only `aws_serverlessapplicationrepository_cloudformation_stack`
  (a deployment action), nothing to list/enumerate.

## Would double-manage an attribute terraformer already inlines
- **aws_sqs_queue_policy**, **aws_sns_topic_policy** — the `aws_sqs_queue` /
  `aws_sns_topic` generators already emit the policy as an inline attribute
  (see their `PostConvertHook`s wrapping `policy` in a heredoc). Emitting a
  separate `_policy` resource would have two resources manage the same policy,
  which Terraform rejects. (Contrast `s3`, which does NOT inline the bucket
  policy and so DOES emit a separate `aws_s3_bucket_policy`.) Revisit only if the
  parent generators stop inlining the policy.

## AWS-deprecated services
- **iotanalytics** — AWS has deprecated the service ("no longer available for
  use"); the SDK package itself is marked deprecated. Removed from the registry.

## Structurally not independently listable (≈212 of current missing-resources.txt)
Per the §3 diff (`missing-resources.txt`, currently 507 entries), ~212 are
resource *classes* that have no standalone List/Describe API — they are
attachments, associations, sub-policies, role/permission assignments, default-*
singletons, accepters, and per-parent settings. Terraform models them as their
own resources, but AWS only exposes them as fields of (or actions on) a parent,
so they cannot be enumerated to import on their own. Current breakdown by
suffix (from `missing-resources.txt`):

    *_association / *_associations   64
    *_policy / *_policies            61   (resource/access policies inlined on the parent or set via a Put* action)
    *_attachment                     25
    *_default* / *_setting(s)        18   (account/region singletons set via Put*/Update*, no list)
    *_accepter                       16   (the *accepting* side of a cross-account handshake — no list)
    *_tag / *_tags                    9
    *_permission(s)                   8
    *_member(ship)                    6
    *_principal*                      5
    *_assignment                      4

These are excluded by the same provider/list-API bound as §1: AWS exposes them
only through a Get/Describe on the *named parent* (you must already know the
ID), or only through a Put/Associate action, never through an enumerable list.
Where a parent-scoped list *does* exist (e.g. `aws_route53_resolver_firewall_rule`
via `ListFirewallRules`, `aws_quicksight_group` via `ListGroups`), the resource
is **built**, not excluded — see the per-service generators.

The balance of `missing-resources.txt` (~295) splits into: (a) the remaining
buildable per-parent/composite tail of plan §9 — built service-by-service each
batch (re-run `gen-gap-list.sh` to watch it shrink; coverage 250→964 so far),
and (b) resources whose backing SDK is not vendored in the pinned module set
(see below) or that are data-plane/report-only (next section).

## SDK not vendored in the pinned aws-sdk-go-v2 module set
Generators cannot be written against SDKs absent from `go.mod`'s pinned
versions, so these provider resources are not buildable in this tree until the
dependency is added (deliberately out of scope to avoid an unrequested dep bump):
- **cloudwatchevidently** (`aws_evidently_feature`, `aws_evidently_launch`, `aws_evidently_segment`)
- **lexmodelbuildingservice** v1 (`aws_lex_bot`, `aws_lex_intent`, `aws_lex_slot_type`, `aws_lex_bot_alias`) — note the v2 `aws_lexv2models_*` set IS built
- **devopsguru** (`aws_devopsguru_*` — notification_channel, resource_collection, service_integration, event_sources_config)
- **drs** (`aws_drs_replication_configuration_template`)
- **eventbridge** `ListEndpoints` (`aws_cloudwatch_event_endpoint`)
- **paymentcryptography** (unregistered service: `aws_paymentcryptography_key`, `aws_paymentcryptography_key_alias`)

## No import / data-plane object resources (not reverse-importable)
Provider resources that either have no `terraform import` support or represent
data-plane objects/actions rather than durable infrastructure, so terraformer
cannot reverse them: `aws_s3_object`, `aws_s3_bucket_object`, `aws_s3_object_copy`,
`aws_dynamodb_table_item`, `aws_dynamodb_table_export`, `aws_kms_ciphertext`,
`aws_ec2_instance_state`, `aws_*_snapshot_copy` / `aws_*_snapshot_import`
(create-from-source actions), `aws_iot_logging_options` /
`aws_iot_indexing_configuration` / `aws_iot_event_configurations` (no import),
and the dx hosted-VIF / accepter / confirmation / proposal handshake resources
(the *accepting*/*requesting* side of a two-account flow, with no list API).

Also intentionally skipped: `aws_elastictranscoder_preset` (ListPresets returns
the hundreds of AWS-managed *system* presets, not user infrastructure — emitting
them is noise, not coverage).

## Data-plane / report-only (plan §1 exclusion list)
`sts`, `pricing`, `compute-optimizer`, `health`, `support`, `trustedadvisor`,
`resourcegroupstaggingapi`, `cloudcontrol`, `*-runtime`, `*-data`, and the CLI
meta entries (`configure`, `login`, …). These have no importable infrastructure.

## Needs more than a single-paginator generator (built where feasible)
Built with extra handling: `quicksight`/`s3control` (account-id scoped),
`mediaconvert` (DescribeEndpoints first), `globalaccelerator`/`shield`
(partition-global), `route53domains` (us-east-1 only),
`route53recoverycontrolconfig`/`route53recoveryreadiness` (nested
cluster→panel→routing-control/safety-rule and readiness check/group/resource-set).
