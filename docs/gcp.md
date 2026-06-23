### Use with GCP

**Supported provider floor:** `terraform-provider-google` / `-google-beta` **= 7.37.0**.
Terraformer refreshes generated state through the real provider plugin, so the
plugin you run must be at or above this version. SDK floor: `google.golang.org/api v0.286.0`.

In order to access the information from your Google Project, you need to provide authentication credentials
by setting up the environment variable `GOOGLE_APPLICATION_CREDENTIALS` with the file path of the JSON
file that contains your service account key. 

[![asciicast](https://asciinema.org/a/243961.svg)](https://asciinema.org/a/243961)

Example:

```
terraformer import google --resources=gcs,forwardingRules,httpHealthChecks --connect=true --regions=europe-west1,europe-west4 --projects=aaa,fff
terraformer import google --resources=gcs,forwardingRules,httpHealthChecks --filter=compute_firewall=rule1:rule2:rule3 --regions=europe-west1 --projects=aaa,fff
```

For google-beta provider:

```
terraformer import google --resources=gcs,forwardingRules,httpHealthChecks --regions=europe-west4 --projects=aaa --provider-type beta
```

List of supported GCP services:

*   `addresses`
    * `google_compute_address`
*   `autoscalers`
    * `google_compute_autoscaler`
*   `artifactRegistry`
    * `google_artifact_registry_repository`
*   `binaryAuthorization`
    * `google_binary_authorization_attestor`
*   `backupdr`
    * `google_backup_dr_management_server`
*   `backendBuckets`
    * `google_compute_backend_bucket`
*   `backendServices`
    * `google_compute_backend_service`
*   `bigQuery`
    * `google_bigquery_dataset`
    * `google_bigquery_table`
*   `biglake`
    * `google_biglake_catalog`
*   `bigtable`
    * `google_bigtable_instance`
*   `cloudRun`
    * `google_cloud_run_v2_service`
*   `composer`
    * `google_composer_environment`
*   `certificateManager`
    * `google_certificate_manager_certificate`
*   `cloudFunctions`
    * `google_cloudfunctions_function`
    * `google_cloudfunctions2_function`
*   `cloudbuild`
    * `google_cloudbuild_trigger`
*   `clouddeploy`
    * `google_clouddeploy_delivery_pipeline`
*   `cloudsql`
    * `google_sql_database`
    * `google_sql_database_instance`
*   `datacatalog`
    * `google_data_catalog_entry_group`
*   `dataProc`
    * `google_dataproc_cluster`
*   `dataplex`
    * `google_dataplex_lake`
*   `datafusion`
    * `google_data_fusion_instance`
*   `datastream`
    * `google_datastream_stream`
*   `dialogflow`
    * `google_dialogflow_cx_agent`
*   `disks`
    * `google_compute_disk`
*   `dns`
    * `google_dns_managed_zone`
    * `google_dns_record_set`
*   `externalVpnGateways`
    * `google_compute_external_vpn_gateway`
*   `firewall`
    * `google_compute_firewall`
*   `essentialContacts`
    * `google_essential_contacts_contact`
*   `documentAI`
    * `google_document_ai_processor`
*   `eventarc`
    * `google_eventarc_trigger`
*   `filestore`
    * `google_filestore_instance`
*   `firestore`
    * `google_firestore_database`
*   `forwardingRules`
    * `google_compute_forwarding_rule`
*   `gcs`
    * `google_storage_bucket`
    * `google_storage_bucket_acl`
    * `google_storage_bucket_iam_binding`
    * `google_storage_bucket_iam_member`
    * `google_storage_bucket_iam_policy`
    * `google_storage_default_object_acl`
    * `google_storage_notification`
*   `gke`
    * `google_container_cluster`
    * `google_container_node_pool`
*   `gkeHub`
    * `google_gke_hub_membership`
*   `globalAddresses`
    * `google_compute_global_address`
*   `globalForwardingRules`
    * `google_compute_global_forwarding_rule`
*   `healthChecks`
    * `google_compute_health_check`
*   `healthcare`
    * `google_healthcare_dataset`
*   `httpHealthChecks`
    * `google_compute_http_health_check`
*   `httpsHealthChecks`
    * `google_compute_https_health_check`
*   `iam`
    * `google_project_iam_custom_role`
    * `google_project_iam_member`
    * `google_service_account`
*   `images`
    * `google_compute_image`
*   `instanceGroupManagers`
    * `google_compute_instance_group_manager`
*   `instanceGroups`
    * `google_compute_instance_group`
*   `instanceTemplates`
    * `google_compute_instance_template`
*   `instances`
    * `google_compute_instance`
*   `interconnectAttachments`
    * `google_compute_interconnect_attachment`
*   `kms`
    * `google_kms_crypto_key`
    * `google_kms_key_ring`
*   `looker`
    * `google_looker_instance`
*   `logging`
    * `google_logging_metric`
*   `memcache`
    * `google_memcache_instance`
*   `memoryStore`
    * `google_redis_instance`
*   `monitoring`
    * `google_monitoring_alert_policy`
    * `google_monitoring_group`
    * `google_monitoring_notification_channel`
    * `google_monitoring_uptime_check_config`
*   `notebooks`
    * `google_notebooks_instance`
*   `netapp`
    * `google_netapp_storage_pool`
*   `networkConnectivity`
    * `google_network_connectivity_hub`
*   `networks`
    * `google_compute_network`
*   `nodeGroups`
    * `google_compute_node_group`
*   `nodeTemplates`
    * `google_compute_node_template`
*   `packetMirrorings`
    * `google_compute_packet_mirroring`
*   `privateca`
    * `google_privateca_ca_pool`
*   `project`
    * `google_project`
*   `pubsub`
    * `google_pubsub_subscription`
    * `google_pubsub_topic`
*   `regionAutoscalers`
    * `google_compute_region_autoscaler`
*   `regionBackendServices`
    * `google_compute_region_backend_service`
*   `regionDisks`
    * `google_compute_region_disk`
*   `regionHealthChecks`
    * `google_compute_region_health_check`
*   `regionInstanceGroupManagers`
    * `google_compute_region_instance_group_manager`
*   `regionInstanceGroups`
    * `google_compute_region_instance_group`
*   `regionSslCertificates`
    * `google_compute_region_ssl_certificate`
*   `regionTargetHttpProxies`
    * `google_compute_region_target_http_proxy`
*   `regionTargetHttpsProxies`
    * `google_compute_region_target_https_proxy`
*   `regionUrlMaps`
    * `google_compute_region_url_map`
*   `reservations`
    * `google_compute_reservation`
*   `resourcePolicies`
    * `google_compute_resource_policy`
*   `routers`
    * `google_compute_router`
*   `routes`
    * `google_compute_route`
*   `schedulerJobs`
    * `google_cloud_scheduler_job`
*   `secretManager`
    * `google_secret_manager_secret`
*   `securityPolicies`
    * `google_compute_security_policy`
*   `serviceDirectory`
    * `google_service_directory_namespace`
*   `spanner`
    * `google_spanner_instance`
*   `sslCertificates`
    * `google_compute_managed_ssl_certificate`
*   `sslPolicies`
    * `google_compute_ssl_policy`
*   `vertexAI`
    * `google_vertex_ai_endpoint`
*   `vmwareengine`
    * `google_vmwareengine_private_cloud`
*   `vpcAccess`
    * `google_vpc_access_connector`
*   `workstations`
    * `google_workstations_workstation_cluster`
*   `workflows`
    * `google_workflows_workflow`
*   `subnetworks`
    * `google_compute_subnetwork`
*   `targetHttpProxies`
    * `google_compute_target_http_proxy`
*   `targetHttpsProxies`
    * `google_compute_target_https_proxy`
*   `targetInstances`
    * `google_compute_target_instance`
*   `targetPools`
    * `google_compute_target_pool`
*   `targetSslProxies`
    * `google_compute_target_ssl_proxy`
*   `targetTcpProxies`
    * `google_compute_target_tcp_proxy`
*   `targetVpnGateways`
    * `google_compute_vpn_gateway`
*   `urlMaps`
    * `google_compute_url_map`
*   `vpnTunnels`
    * `google_compute_vpn_tunnel`

Your `tf` and `tfstate` files are written by default to
`generated/gcp/zone/service`.
