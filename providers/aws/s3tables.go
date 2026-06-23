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
			if _, err := svc.GetTableBucketPolicy(context.TODO(), &s3tables.GetTableBucketPolicyInput{TableBucketARN: bucket.Arn}); err == nil {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					arn, StringValue(bucket.Name), "aws_s3tables_table_bucket_policy", "aws", defaultAllowEmptyValues))
			}

			bucketARN := arn
			for np := s3tables.NewListNamespacesPaginator(svc, &s3tables.ListNamespacesInput{TableBucketARN: &bucketARN}); np.HasMorePages(); {
				npage, err := np.NextPage(context.TODO())
				if err != nil {
					break
				}
				for _, ns := range npage.Namespaces {
					if len(ns.Namespace) == 0 {
						continue
					}
					name := ns.Namespace[0]
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						bucketARN+";"+name, name, "aws_s3tables_namespace", "aws", defaultAllowEmptyValues))
				}
			}
			for tp := s3tables.NewListTablesPaginator(svc, &s3tables.ListTablesInput{TableBucketARN: &bucketARN}); tp.HasMorePages(); {
				tpage, err := tp.NextPage(context.TODO())
				if err != nil {
					break
				}
				for _, t := range tpage.Tables {
					if len(t.Namespace) == 0 {
						continue
					}
					ns, tn := t.Namespace[0], StringValue(t.Name)
					if tn == "" {
						continue
					}
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						bucketARN+";"+ns+";"+tn, ns+"_"+tn, "aws_s3tables_table", "aws", defaultAllowEmptyValues))
					if _, err := svc.GetTablePolicy(context.TODO(), &s3tables.GetTablePolicyInput{
						TableBucketARN: &bucketARN, Namespace: &ns, Name: &tn,
					}); err == nil {
						g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
							bucketARN+";"+ns+";"+tn, ns+"_"+tn, "aws_s3tables_table_policy", "aws", defaultAllowEmptyValues))
					}
				}
			}
		}
	}
	return nil
}
