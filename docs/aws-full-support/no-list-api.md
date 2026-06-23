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

## Structurally not independently listable (the bulk of current missing-resources.txt)
Per the §3 diff (`missing-resources.txt`, currently **240** entries after the
buildable tail was ground down, coverage 250→1231), the large majority are
resource *classes* that have no standalone List/Describe API — they are
attachments, associations, sub-policies, role/permission assignments, default-*
singletons, accepters, and per-parent settings. Terraform models them as their
own resources, but AWS only exposes them as fields of (or actions on) a parent,
so they cannot be enumerated to import on their own. Current breakdown by
suffix (from `missing-resources.txt`):

    *_policy / *_policies            31   (resource/access policies inlined on the parent or set via a Put* action)
    *_association / *_associations   26
    *_accepter                       16   (the *accepting* side of a cross-account handshake — no list)
    *_attachment                     15
    *_tag / *_tags                    7
    *_*exclusive                      6   (Terraform-only management constructs, no AWS object)
    *_permission(s)                   5
    *_default* / *_setting(s)         4   (account/region singletons set via Put*/Update*, no list)

These are excluded by the same provider/list-API bound as §1: AWS exposes them
only through a Get/Describe on the *named parent* (you must already know the
ID), or only through a Put/Associate action, never through an enumerable list.
Where a parent-scoped list *does* exist (e.g. `aws_route53_resolver_firewall_rule`
via `ListFirewallRules`, `aws_quicksight_group` via `ListGroups`), the resource
is **built**, not excluded — see the per-service generators.

The balance of `missing-resources.txt` (240 total) splits into: (a) resources
whose backing SDK is not vendored in the pinned module set (see below), (b)
data-plane/report-only objects with no `terraform import` (next sections), and
(c) `aws_default_*` adopt-existing resources (`aws_default_vpc`,
`aws_default_subnet`, `aws_default_route_table`, `aws_default_security_group`,
`aws_default_vpc_dhcp_options`) which represent pre-existing AWS defaults a user
*adopts* rather than infrastructure to reverse-import. The clean-import,
non-conflicting buildable tail of plan §9 has been ground down to empty —
every remaining entry maps to one of these documented exclusion classes.

## SDK not vendored in the pinned aws-sdk-go-v2 module set
Generators cannot be written against SDKs absent from `go.mod`'s pinned
versions, so these provider resources are not buildable in this tree until the
dependency is added (deliberately out of scope to avoid an unrequested dep bump):
- **cloudwatchevidently** (`aws_evidently_feature`, `aws_evidently_launch`, `aws_evidently_segment`)
- **lexmodelbuildingservice** v1 (`aws_lex_bot`, `aws_lex_intent`, `aws_lex_slot_type`, `aws_lex_bot_alias`) — note the v2 `aws_lexv2models_*` set IS built
- **devopsguru** (`aws_devopsguru_*` — notification_channel, resource_collection, service_integration, event_sources_config)
- **drs** (`aws_drs_replication_configuration_template`)
- **eventbridge** `ListEndpoints` (`aws_cloudwatch_event_endpoint`) — the pinned
  `cloudwatchevents` v1 SDK has no global-endpoints API
- **paymentcryptography** (`aws_paymentcryptography_key_alias`)
- **codecatalyst** (`aws_codecatalyst_project`, `aws_codecatalyst_dev_environment`, `aws_codecatalyst_source_repository`)
- **computeoptimizer** (`aws_computeoptimizer_enrollment_status`, `aws_computeoptimizer_recommendation_preferences`)
- **costoptimizationhub** (`aws_costoptimizationhub_enrollment_status`, `aws_costoptimizationhub_preferences`)
- **simpledb** (`aws_simpledb_domain`), **worklink** (`aws_worklink_fleet`, `aws_worklink_website_certificate_authority_association`)
- **customerprofiles** (`aws_customerprofiles_profile` — data-plane), **dataexchange** revisions (data-plane)
- Operations absent from pinned SDKs: `aws_ecr_repository_creation_template`
  (ECR `ListRepositoryCreationTemplates`), `aws_eks_access_entry` /
  `aws_eks_access_policy_association` (`ListAccessEntries`),
  `aws_lambda_function_recursion_config` (`GetFunctionRecursionConfig`),
  `aws_dynamodb_resource_policy` (`GetResourcePolicy`),
  `aws_codebuild_fleet` (`ListFleets`)

## No import / data-plane object resources (not reverse-importable)
Provider resources that either have no `terraform import` support or represent
data-plane objects/actions rather than durable infrastructure, so terraformer
cannot reverse them: `aws_s3_object`, `aws_s3_bucket_object`, `aws_s3_object_copy`,
`aws_dynamodb_table_item`, `aws_kms_ciphertext`,
`aws_ec2_instance_state`, `aws_*_snapshot_copy` / `aws_*_snapshot_import`
(create-from-source actions), `aws_iot_logging_options` /
`aws_iot_indexing_configuration` / `aws_iot_event_configurations` (no import),
and the dx hosted-VIF / accepter / confirmation / proposal handshake resources
(the *accepting*/*requesting* side of a two-account flow, with no list API).

Specifically excluded (verified individually against the SDK):
- `aws_acmpca_certificate` — IssueCertificate is an action; issued certs are not
  enumerable and the resource has no import.
- `aws_cognito_managed_user_pool_client` — same ListUserPoolClients as
  `aws_cognito_user_pool_client` (already built); a separate generator would
  double-emit the same clients. Managed-vs-unmanaged isn't distinguishable from
  the list.
- `aws_db_instance_automated_backups_replication` — cross-region replication
  *action* defined on the destination side from a source ARN; no reliable way to
  distinguish a replicated automated backup from an original via
  DescribeDBInstanceAutomatedBackups, and import is ambiguous.
- `aws_dx_bgp_peer`, `aws_dx_connection_confirmation`, `aws_dx_hosted_connection`,
  `aws_dx_hosted_{private,public,transit}_virtual_interface` — Direct Connect
  hosted/allocation + BGP-peer resources have no `terraform import` support
  (provider docs) and/or are the allocate-to-other-account action side.
- `aws_ec2_instance_metadata_defaults` — account+region singleton set via
  ModifyInstanceMetadataDefaults; no per-resource import identity.
- `aws_servicecatalog_organizations_access` — org-access enable/disable singleton.
- `aws_vpc_block_public_access_options` — region singleton
  (DescribeVpcBlockPublicAccessOptions); no per-resource import identity.
- `aws_securityhub_standards_control_association` /
  `aws_securityhub_configuration_policy_association` — the enable/disable *state*
  of every control in every enabled standard; `aws_securityhub_standards_control`
  (already built, via DescribeStandardsControls) captures the same control state,
  so emitting the association too would double-manage it.
- `aws_opensearch_authorize_vpc_endpoint_access` — ListVpcEndpointAccess lists the
  authorized principals, but the resource's `terraform import` identity (domain +
  account composite) is not reliably documented; left out rather than emit an
  un-refreshable import.
- `aws_vpc_endpoint_service_allowed_principal` — no `terraform import` support
  (provider docs); the allow-list is managed via ModifyVpcEndpointServicePermissions.

Also intentionally skipped: `aws_elastictranscoder_preset` — Elastic Transcoder's
SDK is AWS-deprecated ("no longer available for use"); adding new generator code
against the deprecated package trips staticcheck SA1019. (ListPresets also returns
the hundreds of AWS-managed *system* presets, which would be noise, not coverage.)
The pre-existing `aws_elastictranscoder_pipeline` generator predates the
deprecation and is left as-is.

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
