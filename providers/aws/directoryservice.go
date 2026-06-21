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

package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/directoryservice"
	"github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
)

type DirectoryServiceGenerator struct {
	AWSService
}

// InitResources enumerates AWS Directory Service directories. The Terraform
// import ID for aws_directory_service_directory is the directory ID.
func (g *DirectoryServiceGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := directoryservice.NewFromConfig(config)

	p := directoryservice.NewDescribeDirectoriesPaginator(svc, &directoryservice.DescribeDirectoriesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.DirectoryDescriptions, "aws_directory_service_directory",
			defaultAllowEmptyValues,
			func(d types.DirectoryDescription) string { return StringValue(d.DirectoryId) },
			func(d types.DirectoryDescription) string { return StringValue(d.DirectoryId) })
	}
	return nil
}
