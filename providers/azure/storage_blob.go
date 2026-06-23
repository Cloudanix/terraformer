// Copyright 2019 The Terraformer Authors.
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

package azure

import (
	"context"
	"fmt"
	"log"
	"net/url"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/Azure/azure-storage-blob-go/azblob"
)

const (
	blobFormatString = `https://%s.blob.core.windows.net`
	blobIDFormat     = `https://%s.blob.core.windows.net/%s/%s`
)

type StorageBlobGenerator struct {
	AzureService
}

// getAccountPrimaryKey fetches a storage account's primary key via the Track 2
// armstorage management SDK (the key is then used for data-plane blob listing).
func (g StorageBlobGenerator) getAccountPrimaryKey(ctx context.Context, accountName, accountGroupName string) string {
	subscriptionID, cred, opts := g.getClientOptions()
	client, err := armstorage.NewAccountsClient(subscriptionID, cred, opts)
	if err != nil {
		log.Fatalf("failed to build storage accounts client: %v", err)
	}
	response, err := client.ListKeys(ctx, accountGroupName, accountName, nil)
	if err != nil {
		log.Fatalf("failed to list keys: %v", err)
	}
	if len(response.Keys) == 0 || response.Keys[0].Value == nil {
		return ""
	}
	return *response.Keys[0].Value
}

func (g StorageBlobGenerator) getContainerURL(ctx context.Context, accountName, accountGroupName, containerName string) (azblob.ContainerURL, error) {
	accountPrimaryKey := g.getAccountPrimaryKey(ctx, accountName, accountGroupName)
	sharedKeyCredential, err := azblob.NewSharedKeyCredential(accountName, accountPrimaryKey)
	if err != nil {
		return azblob.ContainerURL{}, err
	}

	p := azblob.NewPipeline(sharedKeyCredential, azblob.PipelineOptions{})
	accountURL, err := url.Parse(fmt.Sprintf(blobFormatString, accountName))
	if err != nil {
		return azblob.ContainerURL{}, err
	}

	serviceURL := azblob.NewServiceURL(*accountURL, p)
	containerURL := serviceURL.NewContainerURL(containerName)

	return containerURL, nil
}

func (g StorageBlobGenerator) getBlobsFromContainer(ctx context.Context, accountName, accountGroupName, containerName string) ([]azblob.BlobItem, error) {
	containerURL, err := g.getContainerURL(ctx, accountName, accountGroupName, containerName)
	if err != nil {
		return nil, err
	}

	blobListResponse, err := containerURL.ListBlobsFlatSegment(
		ctx,
		azblob.Marker{},
		azblob.ListBlobsSegmentOptions{
			Details: azblob.BlobListingDetails{
				Snapshots: true,
			},
		})
	if err != nil {
		return nil, err
	}

	return blobListResponse.Segment.BlobItems, nil
}

func (g *StorageBlobGenerator) InitResources() error {
	if _, cred, _ := g.getClientOptions(); cred == nil {
		return nil
	}
	ctx := context.Background()

	// Reuse the container generator (sharing this generator's Args, incl. the
	// Track 2 credential) to enumerate containers, then list blobs per container
	// via the data-plane azblob client.
	containerGen := &StorageContainerGenerator{}
	containerGen.SetArgs(g.Args)
	containers, err := containerGen.ListBlobContainers()
	if err != nil {
		return err
	}

	for _, container := range containers {
		parsedContainerID, err := ParseAzureResourceID(container.InstanceState.ID)
		if err != nil {
			return err
		}
		accountName := container.InstanceState.Attributes["storage_account_name"]
		containerName := container.InstanceState.Attributes["name"]
		blobs, err := g.getBlobsFromContainer(ctx, accountName, parsedContainerID.ResourceGroup, containerName)
		if err != nil {
			return err
		}
		for _, blobItem := range blobs {
			g.AppendSimpleResource(
				fmt.Sprintf(blobIDFormat, accountName, containerName, blobItem.Name),
				blobItem.Name,
				"azurerm_storage_blob")
		}
	}
	return nil
}
