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
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas/types"
)

func TestQuotasFromChangeHistory(t *testing.T) {
	changes := []types.RequestedServiceQuotaChange{
		{ServiceCode: aws.String("ec2"), QuotaCode: aws.String("L-1216C47A"), QuotaName: aws.String("Running On-Demand instances")},
		// duplicate of the first quota (another change request) → collapsed
		{ServiceCode: aws.String("ec2"), QuotaCode: aws.String("L-1216C47A"), QuotaName: aws.String("Running On-Demand instances")},
		// missing quota code → skipped (un-importable)
		{ServiceCode: aws.String("lambda"), QuotaCode: nil, QuotaName: aws.String("Concurrent executions")},
		// missing service code → skipped
		{ServiceCode: nil, QuotaCode: aws.String("L-XXXX"), QuotaName: aws.String("orphan")},
		// missing name → id used as name
		{ServiceCode: aws.String("vpc"), QuotaCode: aws.String("L-F678F1CE"), QuotaName: nil},
	}

	got := quotasFromChangeHistory(changes)
	want := []quotaRef{
		{id: "ec2/L-1216C47A", name: "Running On-Demand instances"},
		{id: "vpc/L-F678F1CE", name: "vpc/L-F678F1CE"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("quotasFromChangeHistory() = %#v, want %#v", got, want)
	}
}

func TestQuotasFromChangeHistoryEmpty(t *testing.T) {
	if got := quotasFromChangeHistory(nil); len(got) != 0 {
		t.Errorf("expected no quotas from empty history, got %#v", got)
	}
}
