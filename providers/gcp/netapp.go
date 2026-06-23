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
	"google.golang.org/api/netapp/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var netappAllowEmptyValues = []string{""}

var netappAdditionalFields = map[string]interface{}{}

type NetappGenerator struct {
	GCPService
}

// Run on storagePoolsList and create for each TerraformResource
func (g NetappGenerator) createResources(ctx context.Context, storagePoolsList *netapp.ProjectsLocationsStoragePoolsListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := storagePoolsList.Pages(ctx, func(page *netapp.ListStoragePoolsResponse) error {
		for _, obj := range page.StoragePools {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_netapp_storage_pool",
				g.ProviderName,
				map[string]string{
					"name":     name,
					"project":  g.GetArgs()["project"].(string),
					"location": location,
				},
				netappAllowEmptyValues,
				netappAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
func (g *NetappGenerator) InitResources() error {
	ctx := context.Background()
	netappService, err := netapp.NewService(ctx)
	if err != nil {
		return err
	}

	parent := "projects/" + g.GetArgs()["project"].(string) + "/locations/" + g.GetArgs()["region"].(compute.Region).Name
	storagePoolsList := netappService.Projects.Locations.StoragePools.List(parent)
	g.Resources = append(g.Resources, g.createResources(ctx, storagePoolsList)...)

	loc := g.GetArgs()["region"].(compute.Region).Name
	proj := g.GetArgs()["project"].(string)

	volumesList := netappService.Projects.Locations.Volumes.List(parent)
	volumeResources := g.createVolumesResources(ctx, volumesList)
	g.Resources = append(g.Resources, volumeResources...)
	for _, vr := range volumeResources {
		volFull := vr.InstanceState.ID
		volName := strings.Split(volFull, "/")[len(strings.Split(volFull, "/"))-1]
		if serr := netappService.Projects.Locations.Volumes.Snapshots.List(volFull).Pages(ctx, func(p *netapp.ListSnapshotsResponse) error {
			for _, o := range p.Snapshots {
				t := strings.Split(o.Name, "/")
				g.Resources = append(g.Resources, terraformutils.NewResource(
					o.Name, t[len(t)-1], "google_netapp_volume_snapshot", g.ProviderName,
					map[string]string{"name": t[len(t)-1], "volume_name": volName, "location": loc, "project": proj},
					netappAllowEmptyValues, netappAdditionalFields))
			}
			return nil
		}); serr != nil {
			log.Println(serr)
		}
		if rerr := netappService.Projects.Locations.Volumes.Replications.List(volFull).Pages(ctx, func(p *netapp.ListReplicationsResponse) error {
			for _, o := range p.Replications {
				t := strings.Split(o.Name, "/")
				g.Resources = append(g.Resources, terraformutils.NewResource(
					o.Name, t[len(t)-1], "google_netapp_volume_replication", g.ProviderName,
					map[string]string{"name": t[len(t)-1], "volume_name": volName, "location": loc, "project": proj},
					netappAllowEmptyValues, netappAdditionalFields))
			}
			return nil
		}); rerr != nil {
			log.Println(rerr)
		}
		if qerr := netappService.Projects.Locations.Volumes.QuotaRules.List(volFull).Pages(ctx, func(p *netapp.ListQuotaRulesResponse) error {
			for _, o := range p.QuotaRules {
				t := strings.Split(o.Name, "/")
				g.Resources = append(g.Resources, terraformutils.NewResource(
					o.Name, t[len(t)-1], "google_netapp_volume_quota_rule", g.ProviderName,
					map[string]string{"name": t[len(t)-1], "location": loc, "project": proj},
					netappAllowEmptyValues, netappAdditionalFields))
			}
			return nil
		}); qerr != nil {
			log.Println(qerr)
		}
	}
	if err := netappService.Projects.Locations.ActiveDirectories.List(parent).Pages(ctx, func(p *netapp.ListActiveDirectoriesResponse) error {
		for _, o := range p.ActiveDirectories {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_netapp_active_directory", g.ProviderName,
				map[string]string{"name": name, "project": proj, "location": loc},
				netappAllowEmptyValues, netappAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := netappService.Projects.Locations.BackupVaults.List(parent).Pages(ctx, func(p *netapp.ListBackupVaultsResponse) error {
		for _, o := range p.BackupVaults {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_netapp_backup_vault", g.ProviderName,
				map[string]string{"name": name, "project": proj, "location": loc},
				netappAllowEmptyValues, netappAdditionalFields))
			vaultName := name
			if berr := netappService.Projects.Locations.BackupVaults.Backups.List(o.Name).Pages(ctx, func(bp *netapp.ListBackupsResponse) error {
				for _, b := range bp.Backups {
					bt := strings.Split(b.Name, "/")
					g.Resources = append(g.Resources, terraformutils.NewResource(
						b.Name, bt[len(bt)-1], "google_netapp_backup", g.ProviderName,
						map[string]string{"name": bt[len(bt)-1], "vault_name": vaultName, "location": loc, "project": proj},
						netappAllowEmptyValues, netappAdditionalFields))
				}
				return nil
			}); berr != nil {
				log.Println(berr)
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := netappService.Projects.Locations.KmsConfigs.List(parent).Pages(ctx, func(p *netapp.ListKmsConfigsResponse) error {
		for _, o := range p.KmsConfigs {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_netapp_kmsconfig", g.ProviderName,
				map[string]string{"name": name, "project": proj, "location": loc},
				netappAllowEmptyValues, netappAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := netappService.Projects.Locations.BackupPolicies.List(parent).Pages(ctx, func(p *netapp.ListBackupPoliciesResponse) error {
		for _, o := range p.BackupPolicies {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_netapp_backup_policy", g.ProviderName,
				map[string]string{"name": name, "project": proj, "location": loc},
				netappAllowEmptyValues, netappAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := netappService.Projects.Locations.HostGroups.List(parent).Pages(ctx, func(p *netapp.ListHostGroupsResponse) error {
		for _, o := range p.HostGroups {
			t := strings.Split(o.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				o.Name, name, "google_netapp_host_group", g.ProviderName,
				map[string]string{"name": name, "project": proj, "location": loc},
				netappAllowEmptyValues, netappAdditionalFields))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}

// Run on volumesList and create for each TerraformResource
func (g NetappGenerator) createVolumesResources(ctx context.Context, list *netapp.ProjectsLocationsVolumesListCall) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	location := g.GetArgs()["region"].(compute.Region).Name
	if err := list.Pages(ctx, func(page *netapp.ListVolumesResponse) error {
		for _, obj := range page.Volumes {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			resources = append(resources, terraformutils.NewResource(
				obj.Name, name, "google_netapp_volume", g.ProviderName,
				map[string]string{"name": name, "project": g.GetArgs()["project"].(string), "location": location},
				netappAllowEmptyValues, netappAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}
