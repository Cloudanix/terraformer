# Resources excluded from auto-import (not "missing" — impossible/deferred)

Tracks `google_*` resources that appear in `missing-resources.txt` but **cannot**
be imported by the basic compute generator or a simple list→NewResource mapper.
Keeps "missing" honest: these need a different mechanism, not a generator entry.

## Categories

### A. IAM sub-resources → Phase 4 IAM workstream (parent walk)
`*_iam_member`, `*_iam_binding`, `*_iam_policy`, `*_iam_audit_config` for every
parent type. Not list-able directly; discovered by walking parent resources and
reading their IAM policy. Handled in the IAM expansion, not the compute codegen.
Examples: `google_compute_{disk,image,instance,subnetwork,snapshot,storage_pool,
instant_snapshot,instance_template,region_disk,region_instant_snapshot}_iam_*`.

### B. Sub-resources requiring a parent walk (not project/region/zone scoped)
The compute codegen only handles list methods parameterized by project/region/zone.
These need a parent ID (router, instance-group-manager, cross-site-network, policy):
- `google_compute_router_nat`, `_router_peer`, `_router_interface`, `_router_route_policy` (parent: router)
- `google_compute_per_instance_config`, `_region_per_instance_config` (parent: IGM)
- `google_compute_instance_group_named_port`, `_instance_group_membership`
- `google_compute_network_peering`, `_network_peering_routes_config` (parent: network)
- `google_compute_resize_request` / `instanceGroupManagerResizeRequests` (parent: IGM)
- `google_compute_wire_group` (list needs `crossSiteNetwork` parent)
- `google_compute_*_firewall_policy_rule`, `_association` (parent: firewall policy)
- `google_compute_*_security_policy_rule` (parent: security policy)
- `google_compute_resource_policy_attachment`, `_disk_resource_policy_attachment`
- `google_compute_backend_*_signed_url_key`, `_disk_async_replication`

### C. Org/folder-scoped (list needs parentId, not project)
- `google_compute_firewall_policy`, `_firewall_policy_association`, `_firewall_policy_rule`
- `google_compute_organization_security_policy`, `_association`, `_rule`

### D. Project singletons / no standalone list
- `google_compute_project_metadata`, `_project_metadata_item`
- `google_compute_shared_vpc_host_project`, `_shared_vpc_service_project`
- `google_compute_project_default_network_tier`, `_project_cloud_armor_tier`
- `google_compute_instance_settings`, `_snapshot_settings`
- `google_compute_attached_disk`, `_instance_from_template` (synthetic / action resources)
- `google_compute_network_endpoint`, `_network_endpoints`, `_global_network_endpoint` (members of a NEG; parent walk)

> Categories A–C are buildable later (IAM workstream / hand-written parent-walk
> generators); D is genuinely not auto-discoverable and left to the user.
> Re-evaluate when the per-service hand-written work reaches each parent type.
