// Copyright 2020 The Terraformer Authors.
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

	"github.com/aws/aws-sdk-go-v2/service/qldb"
	"github.com/aws/aws-sdk-go-v2/service/qldb/types"
)

type QLDBGenerator struct {
	AWSService
}

func (g *QLDBGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := qldb.NewFromConfig(config)
	p := qldb.NewListLedgersPaginator(svc, &qldb.ListLedgersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		g.Resources = appendSimpleResources(g.Resources, page.Ledgers, "aws_qldb_ledger",
			defaultAllowEmptyValues,
			func(l types.LedgerSummary) string { return StringValue(l.Name) },
			func(l types.LedgerSummary) string { return StringValue(l.Name) })
	}
	return nil
}
