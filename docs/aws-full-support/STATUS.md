# AWS Full-Support — Implementation Status

Tracks progress against [plan.md](plan.md). Updated 2026-06-21.

## Done

### Foundations (§0) — complete
Shared helpers (`appendSimpleResources`, `wrapPolicyAttribute`,
`defaultAllowEmptyValues`), SDK retryer, per-region config cache, and the
`serviceScope` map + assertion test all landed earlier.

### New services (§4a P1/P2 — the offline-buildable subset)
application-autoscaling, backup, datasync, dax, directory-service (ds), dlm,
dms, fsx, glacier, **globalaccelerator** (partition-global), route53resolver,
servicediscovery, **servicequotas** (T5, change-history filter), sesv2,
**sso-admin**, storagegateway, transfer.

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

## Blocked / deferred

### §4a remaining new services — BLOCKED (no network)
guardduty, inspector2, macie2, fms, shield, securitylake, detective,
network-firewall, ram, acm-pca, signer, athena, lakeformation, neptune,
memorydb, timestream*, opensearch(+serverless), sagemaker, quicksight, pipes,
scheduler, mwaa, and the rest of P2/P3/P4.

**Why:** the sandbox blocks Go module downloads (proxy TLS failure + the module
VCS cache dir is read-only). Every one of these needs an
`aws-sdk-go-v2/service/<svc>` module that is **not** in the local cache, so it
cannot be added to `go.mod` offline. All 98 cached SDK modules already map to
registered services. **Unblock:** run `go get` for the needed service modules
once network is available, then follow the §5 recipe per service.

### §3 authoritative gap list — BLOCKED (no network)
Generating `tf-aws-all-resources.txt` needs `terraform init` + provider schema
download. Run §3 once network is available to produce the exact remaining diff.

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
