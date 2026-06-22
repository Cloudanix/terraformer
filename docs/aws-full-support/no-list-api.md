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

## Data-plane / report-only (plan §1 exclusion list)
`sts`, `pricing`, `compute-optimizer`, `health`, `support`, `trustedadvisor`,
`resourcegroupstaggingapi`, `cloudcontrol`, `*-runtime`, `*-data`, and the CLI
meta entries (`configure`, `login`, …). These have no importable infrastructure.

## Needs more than a single-paginator generator (built where feasible)
Built with extra handling: `quicksight`/`s3control` (account-id scoped),
`mediaconvert` (DescribeEndpoints first), `globalaccelerator`/`shield`
(partition-global), `route53domains` (us-east-1 only).

Not yet built — multi-step / region-specific control plane, revisit on demand:
- **route53-recovery-control-config**, **route53-recovery-readiness** — recovery
  control plane with its own regional endpoints.
