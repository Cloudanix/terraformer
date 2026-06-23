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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3vectors"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type S3VectorsGenerator struct {
	AWSService
}

// InitResources enumerates S3 Vectors vector buckets, their indexes, and bucket
// policies. Import IDs: bucket name; "<bucket>,<index>"; bucket name.
func (g *S3VectorsGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := s3vectors.NewFromConfig(config)
	ctx := awsContext()
	for p := s3vectors.NewListVectorBucketsPaginator(svc, &s3vectors.ListVectorBucketsInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, b := range page.VectorBuckets {
			name := StringValue(b.VectorBucketName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_s3vectors_vector_bucket", "aws", defaultAllowEmptyValues))
			if _, err := svc.GetVectorBucketPolicy(ctx, &s3vectors.GetVectorBucketPolicyInput{VectorBucketName: aws.String(name)}); err == nil {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					name, name, "aws_s3vectors_vector_bucket_policy", "aws", defaultAllowEmptyValues))
			}
			for ip := s3vectors.NewListIndexesPaginator(svc, &s3vectors.ListIndexesInput{VectorBucketName: aws.String(name)}); ip.HasMorePages(); {
				ipage, err := ip.NextPage(ctx)
				if err != nil {
					break
				}
				for _, idx := range ipage.Indexes {
					iname := StringValue(idx.IndexName)
					if iname == "" {
						continue
					}
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						name+","+iname, name+"_"+iname, "aws_s3vectors_index", "aws", defaultAllowEmptyValues))
				}
			}
		}
	}
	return nil
}
