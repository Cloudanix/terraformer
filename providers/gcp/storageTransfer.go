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

	"google.golang.org/api/storagetransfer/v1"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

var storageTransferAllowEmptyValues = []string{""}

var storageTransferAdditionalFields = map[string]interface{}{}

type StorageTransferGenerator struct {
	GCPService
}

// Generate TerraformResources from GCP API,
func (g *StorageTransferGenerator) InitResources() error {
	ctx := context.Background()
	storageTransferService, err := storagetransfer.NewService(ctx)
	if err != nil {
		return err
	}
	project := g.GetArgs()["project"].(string)

	// TransferJobs.List requires a JSON filter naming the project.
	jobsList := storageTransferService.TransferJobs.List(`{"projectId":"` + project + `"}`)
	if err := jobsList.Pages(ctx, func(page *storagetransfer.ListTransferJobsResponse) error {
		for _, obj := range page.TransferJobs {
			t := strings.Split(obj.Name, "/")
			name := t[len(t)-1]
			g.Resources = append(g.Resources, terraformutils.NewResource(
				obj.Name,
				name,
				"google_storage_transfer_job",
				g.ProviderName,
				map[string]string{
					"name":    obj.Name,
					"project": project,
				},
				storageTransferAllowEmptyValues,
				storageTransferAdditionalFields,
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return nil
}
