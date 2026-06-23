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

- **Op absent in pinned SDK** (would not compile; no dep bump per plan §11):
  `aws_ecr_repository_creation_template` (ecr v1.24.5 lacks
  DescribeRepositoryCreationTemplates), `aws_eks_access_entry` /
  `aws_eks_access_policy_association` (eks v1.35.5 lacks ListAccessEntries),
  `aws_lambda_function_recursion_config` (lambda v1.49.5 lacks
  GetFunctionRecursionConfig), `aws_iam_organizations_features` (iam v1.28.5
  lacks ListOrganizationsFeatures). These become buildable only after the
  respective module is bumped.
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

## Result

`missing-resources.txt` = **215**, every entry mapped to one of the verdicts
above (the classifier in the build scripts reports 0 unclassified). Coverage
**1256** `aws_*` types. The buildable frontier is now genuinely exhausted for the
pinned SDK set; further gains require a deliberate `aws-sdk-go-v2` bump (the
"op absent in pinned SDK" group) or accepting the ambiguous-import risk above.
