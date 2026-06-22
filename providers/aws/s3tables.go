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

	"github.com/aws/aws-sdk-go-v2/service/s3tables"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type S3TablesGenerator struct {
	AWSService
}

// InitResources enumerates S3 Tables table buckets. Import ID is the bucket ARN.
func (g *S3TablesGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := s3tables.NewFromConfig(config)

	p := s3tables.NewListTableBucketsPaginator(svc, &s3tables.ListTableBucketsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, bucket := range page.TableBuckets {
			arn := StringValue(bucket.Arn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(bucket.Name), "aws_s3tables_table_bucket", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
