// Copyright 2018 The Terraformer Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gcp

import (
	"context"
	"log"
	"strings"

	"google.golang.org/api/compute/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var computeIamAllowEmptyValues = []string{""}

var computeIamAdditionalFields = map[string]interface{}{}

// ComputeIamGenerator imports the resource-level `_iam_member` resources for the
// compute resource types whose SDK exposes GetIamPolicy. The base resources
// themselves are emitted by the generated *_gen.go files; this generator only
// adds the IAM member bindings (which the codegen has no template for).
type ComputeIamGenerator struct {
	GCPService
}

// computeIamRoleMember is a single flattened (role, member) pair from an IAM policy.
type computeIamRoleMember struct {
	Role   string
	Member string
}

// expandComputeBindings flattens a compute IAM policy's bindings into one entry
// per (role, member) pair — the granularity of a google_compute_*_iam_member.
func expandComputeBindings(bindings []*compute.Binding) []computeIamRoleMember {
	out := []computeIamRoleMember{}
	for _, b := range bindings {
		if b == nil {
			continue
		}
		for _, m := range b.Members {
			out = append(out, computeIamRoleMember{Role: b.Role, Member: m})
		}
	}
	return out
}

// emitMembers appends one resource per flattened binding for a single base resource.
func (g *ComputeIamGenerator) emitMembers(policy *compute.Policy, tfType, resourceName string, attrs map[string]string) {
	if policy == nil {
		return
	}
	for _, rm := range expandComputeBindings(policy.Bindings) {
		a := map[string]string{"role": rm.Role, "member": rm.Member, "project": attrs["project"]}
		for k, v := range attrs {
			a[k] = v
		}
		a["role"] = rm.Role
		a["member"] = rm.Member
		g.Resources = append(g.Resources, terraformutils.NewResource(
			resourceName+" "+rm.Role+" "+rm.Member,
			resourceName+"_"+rm.Role+"_"+rm.Member,
			tfType, g.ProviderName, a,
			computeIamAllowEmptyValues, computeIamAdditionalFields))
	}
}

// Generate TerraformResources from GCP API.
func (g *ComputeIamGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := compute.NewService(ctx)
	if err != nil {
		return err
	}
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region)
	regionName := region.Name

	// Global resources.
	if err := svc.Images.List(project).Pages(ctx, func(page *compute.ImageList) error {
		for _, o := range page.Items {
			if pol, e := svc.Images.GetIamPolicy(project, o.Name).Do(); e == nil {
				g.emitMembers(pol, "google_compute_image_iam_member", o.Name, map[string]string{"image": o.Name, "project": project})
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := svc.Snapshots.List(project).Pages(ctx, func(page *compute.SnapshotList) error {
		for _, o := range page.Items {
			if pol, e := svc.Snapshots.GetIamPolicy(project, o.Name).Do(); e == nil {
				g.emitMembers(pol, "google_compute_snapshot_iam_member", o.Name, map[string]string{"name": o.Name, "project": project})
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := svc.InstanceTemplates.List(project).Pages(ctx, func(page *compute.InstanceTemplateList) error {
		for _, o := range page.Items {
			if pol, e := svc.InstanceTemplates.GetIamPolicy(project, o.Name).Do(); e == nil {
				g.emitMembers(pol, "google_compute_instance_template_iam_member", o.Name, map[string]string{"name": o.Name, "project": project})
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	// Regional resources.
	if err := svc.Subnetworks.List(project, regionName).Pages(ctx, func(page *compute.SubnetworkList) error {
		for _, o := range page.Items {
			if pol, e := svc.Subnetworks.GetIamPolicy(project, regionName, o.Name).Do(); e == nil {
				g.emitMembers(pol, "google_compute_subnetwork_iam_member", o.Name, map[string]string{"subnetwork": o.Name, "region": regionName, "project": project})
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := svc.RegionDisks.List(project, regionName).Pages(ctx, func(page *compute.DiskList) error {
		for _, o := range page.Items {
			if pol, e := svc.RegionDisks.GetIamPolicy(project, regionName, o.Name).Do(); e == nil {
				g.emitMembers(pol, "google_compute_region_disk_iam_member", o.Name, map[string]string{"name": o.Name, "region": regionName, "project": project})
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := svc.RegionInstantSnapshots.List(project, regionName).Pages(ctx, func(page *compute.InstantSnapshotList) error {
		for _, o := range page.Items {
			if pol, e := svc.RegionInstantSnapshots.GetIamPolicy(project, regionName, o.Name).Do(); e == nil {
				g.emitMembers(pol, "google_compute_region_instant_snapshot_iam_member", o.Name, map[string]string{"name": o.Name, "region": regionName, "project": project})
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	// Zonal resources — iterate the region's zones (matching the generated compute files).
	for _, zoneLink := range region.Zones {
		zoneParts := strings.Split(zoneLink, "/")
		zone := zoneParts[len(zoneParts)-1]
		if err := svc.Disks.List(project, zone).Pages(ctx, func(page *compute.DiskList) error {
			for _, o := range page.Items {
				if pol, e := svc.Disks.GetIamPolicy(project, zone, o.Name).Do(); e == nil {
					g.emitMembers(pol, "google_compute_disk_iam_member", o.Name, map[string]string{"name": o.Name, "zone": zone, "project": project})
				}
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
		if err := svc.Instances.List(project, zone).Pages(ctx, func(page *compute.InstanceList) error {
			for _, o := range page.Items {
				if pol, e := svc.Instances.GetIamPolicy(project, zone, o.Name).Do(); e == nil {
					g.emitMembers(pol, "google_compute_instance_iam_member", o.Name, map[string]string{"instance_name": o.Name, "zone": zone, "project": project})
				}
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
		if err := svc.InstantSnapshots.List(project, zone).Pages(ctx, func(page *compute.InstantSnapshotList) error {
			for _, o := range page.Items {
				if pol, e := svc.InstantSnapshots.GetIamPolicy(project, zone, o.Name).Do(); e == nil {
					g.emitMembers(pol, "google_compute_instant_snapshot_iam_member", o.Name, map[string]string{"name": o.Name, "zone": zone, "project": project})
				}
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
		if err := svc.StoragePools.List(project, zone).Pages(ctx, func(page *compute.StoragePoolList) error {
			for _, o := range page.Items {
				if pol, e := svc.StoragePools.GetIamPolicy(project, zone, o.Name).Do(); e == nil {
					g.emitMembers(pol, "google_compute_storage_pool_iam_member", o.Name, map[string]string{"name": o.Name, "zone": zone, "project": project})
				}
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
	}
	return nil
}
