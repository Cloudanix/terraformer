# Thorough resource buildability review

A systematic audit of every entry in `missing-resources.txt` to separate
"genuinely buildable" from "documented exclusion" — and to catch any resource
previously mis-classified as impossible.

## Method (reproducible)

1. **Map** each missing `aws_<svc>_<thing>` to its service's vendored SDK package
   by finding already-built sibling generators sharing the prefix (see the probe
   script in commit history). 199/232 had a built sibling; 33 were orphans
   (new/unvendored services or prefix mismatches).
2. **Probe** each candidate for a `List*`/`Describe*`/`Get*`/`Search*`/`BatchGet*`
   op whose name shares a token with the resource — surfacing real APIs.
3. **CRITICAL pin check:** verify the op exists in the **go.mod-pinned** SDK
   version, not merely in some newer cached version. (The first probe pass over
   *all* cached versions produced false positives — e.g. `eks.ListAccessEntries`
   exists in cached v1.87.0 but **not** in pinned v1.35.5, so it does not
   compile. Always check `service/<pkg>@<go.mod version>`.)
4. **Assess** each pinned-confirmed op for a clean `terraform import` ID and
   absence of double-management with an inlined parent attribute.
5. **Build** the ones that pass; **document** the rest here + in `no-list-api.md`.

## Found buildable and BUILT (this review)

Resources previously sitting in `missing-resources.txt` that the review proved
buildable against the pinned SDKs, now implemented (one commit each):

- `aws_vpc_endpoint_route_table_association`, `_subnet_association`,
  `_security_group_association` — DescribeVpcEndpoints RouteTableIds/SubnetIds/Groups
- `aws_eip_domain_name` — DescribeAddressesAttribute(domain-name)
- `aws_ssoadmin_account_assignment` — ListAccountsForProvisionedPermissionSet → ListAccountAssignments
- `aws_glue_partition`, `aws_glue_catalog_table_optimizer` — GetPartitions / GetTableOptimizer
- `aws_cloudwatch_log_stream` — DescribeLogStreams
- `aws_sagemaker_device` — ListDevices per device fleet
- `aws_resourcegroups_resource` — ListGroupResources (manual/non-query groups only)
- `aws_media_store_container_policy` — GetContainerPolicy
- `aws_route53profiles_association`, `_resource_association` — ListProfile(Resource)Associations
- `aws_amplify_domain_association` — ListDomainAssociations
- `aws_resiliencehub_resiliency_policy` — ListResiliencyPolicies
- `aws_redshiftserverless_custom_domain_association` — ListCustomDomainAssociations
- `aws_organizations_resource_policy` — DescribeResourcePolicy (singleton)
- `aws_cloudsearch_domain_service_access_policy` — DescribeServiceAccessPolicies
- `aws_ecs_account_setting_default` — ListAccountSettings
- `aws_verifiedaccess_instance_logging_configuration` — DescribeVerifiedAccessInstanceLoggingConfigurations
- `aws_servicecatalog_budget_resource_association` — ListBudgetsForResource
- `aws_bedrockagent_agent_knowledge_base_association` — ListAgentKnowledgeBases
- `aws_servicequotas_template_association` — GetAssociationForServiceQuotaTemplate (singleton)
- `aws_auditmanager_framework_share` — ListAssessmentFrameworkShareRequests(SENT)

## Confirmed NOT buildable (verdicts → see no-list-api.md for the catalogue)

- **Op absent in pinned SDK** — **RESOLVED by the aws-sdk-go-v2 upgrade.** The
  84-module bump (eks v1.35.5→v1.87.0, ecr v1.24.5→v1.58.4, lambda
  v1.49.5→v1.93.0, iam v1.28.5→v1.54.5, codebuild→v1.69.4, dynamodb→v1.59.0,
  ec2→v1.307.1, s3tables) vendored the missing ops, so all of these are now
  BUILT: `aws_ecr_repository_creation_template`, `aws_eks_access_entry`,
  `aws_eks_access_policy_association`, `aws_lambda_function_recursion_config`,
  `aws_iam_organizations_features`, `aws_codebuild_fleet`,
  `aws_dynamodb_resource_policy`, `aws_vpc_security_group_vpc_association`,
  `aws_s3tables_table_bucket_policy`, `aws_s3tables_table_policy`.
- **Unvendored SDK entirely:** codecatalyst, drs, evidently, lex (v1),
  paymentcryptography, computeoptimizer, costoptimizationhub, simpledb, worklink,
  customerprofiles, dataexchange.
- **Redundant with an already-built resource:** `aws_cloudformation_stack_instances`
  (same ListStackInstances as the built `aws_cloudformation_stack_set_instance`),
  `aws_securityhub_standards_control_association` (same control state as built
  `aws_securityhub_standards_control`), `aws_cognito_managed_user_pool_client`
  (same ListUserPoolClients as built `aws_cognito_user_pool_client`).
- **Ambiguous / unreliable import ID:** `aws_redshift_data_share_consumer_association`
  (consumer-side composite not reliably reconstructable),
  `aws_sagemaker_servicecatalog_portfolio_status` (region/account-derived id),
  `aws_opensearch_authorize_vpc_endpoint_access` (domain+account composite).
- **Deferred — multi-level / cross-service fan-out, low value:** datazone
  (asset_type/form_type/glossary/glossary_term/user_profile need
  domain+owning-project context), `aws_controltower_control` (ListEnabledControls
  needs per-OU target enumeration from Organizations).
- **No `terraform import` / data-plane / singleton-without-identity / would
  double-manage an inlined attribute / cross-account handshake / `aws_default_*`
  adopt-existing / tag-per-parent:** the large structural majority — fully
  catalogued in `no-list-api.md`.

## Post-upgrade service review (all OTHER addable services)

After the aws-sdk-go-v2 upgrade, audited every terraform-provider-aws **service
prefix with zero terraformer coverage** (17 found) to see what whole services
could now be added. Findings:

- **No net-new service from already-vendored SDKs** — every SDK that is in
  `go.mod` and has terraform-provider-aws resources is already a registered
  generator. The cached modules that the earlier review flagged "unvendored"
  (evidently, lexmodelsv2, paymentcryptography, customerprofiles, dataexchange,
  bcmdataexports, chatbot) turned out to be **already registered**.

- **Added this review** (SDK already vendored, ops now usable):
  - `aws_paymentcryptography_key_alias` (ListAliases)
  - ELB-classic policy family on the existing `elb` service:
    `aws_load_balancer_policy`, `aws_app_cookie_stickiness_policy`,
    `aws_lb_cookie_stickiness_policy`, `aws_proxy_protocol_policy`,
    `aws_load_balancer_listener_policy`,
    `aws_load_balancer_backend_server_policy` (DescribeLoadBalancerPolicies +
    listener/backend descriptions; elasticloadbalancing classic SDK).

- **Blocked: SDK module not vendored** — these services would each be a clean
  new generator, but their `aws-sdk-go-v2/service/<x>` module is absent from
  `go.mod` (and the sandbox blocks fetching new modules). Adding any of them is a
  deliberate `go get` + dependency-addition decision:
  | Service (SDK module) | TF resources | List API exists |
  |---|---|---|
  | `account` | account_region, account_alternate_contact, account_primary_contact | ListRegions / GetAlternateContact / GetContactInformation |
  | `codecatalyst` | project, dev_environment, source_repository | ListProjects / ListSourceRepositories |
  | `computeoptimizer` | enrollment_status, recommendation_preferences | GetEnrollmentStatus / GetRecommendationPreferences (singletons) |
  | `costoptimizationhub` | enrollment_status, preferences | singletons |
  | `devopsguru` | notification_channel, resource_collection, … | ListNotificationChannels / GetResourceCollection |
  | `lexmodelbuildingservice` (lex v1) | lex_bot, lex_intent, lex_slot_type, lex_bot_alias | GetBots / GetIntents / GetSlotTypes |
  | `simpledb` | simpledb_domain | ListDomains (service AWS-deprecated) |
  | `worklink` | worklink_fleet, … | ListFleets (service AWS-deprecated) |
  | `cloudfrontkeyvaluestore` | cloudfrontkeyvaluestore_key | ListKeys (data-plane keys) |
  | `serverlessapplicationrepository` | cloudformation_stack | deploy action, no import |

  Recommended to actually add (real infra, supported services): **account**,
  **codecatalyst**, **devopsguru**. The rest are deprecated/data-plane/no-import.

- **Remain not addable regardless of SDK** (data-plane / no-import / action):
  `cloudcontrolapi_resource`, `redshiftdata_statement`,
  `snapshot_create_volume_permission`. See no-list-api.md.

## Result

`missing-resources.txt` = **198**, every entry mapped to one of the verdicts
above. Coverage **1273** `aws_*` types. After the aws-sdk-go-v2 upgrade resolved
the "op absent" group, the buildable frontier is exhausted: remaining gains need
an SDK module that is **entirely unvendored** (codecatalyst, drs, evidently, lex
v1, paymentcryptography, computeoptimizer, costoptimizationhub, simpledb,
worklink, customerprofiles, dataexchange — adding those is a larger, separate
dependency decision) or accepting the documented ambiguous-import / redundant /
data-plane exclusions.
