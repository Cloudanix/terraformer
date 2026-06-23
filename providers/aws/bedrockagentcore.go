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

	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type BedrockAgentCoreGenerator struct {
	AWSService
}

// InitResources enumerates Bedrock AgentCore control-plane resources. Import IDs
// are the resource id (or ARN where the resource has no short id).
func (g *BedrockAgentCoreGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := bedrockagentcorecontrol.NewFromConfig(config)
	ctx := context.TODO()
	emit := func(id, name, tfType string) {
		if id != "" {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, name, tfType, "aws", defaultAllowEmptyValues))
		}
	}

	for p := bedrockagentcorecontrol.NewListAgentRuntimesPaginator(svc, &bedrockagentcorecontrol.ListAgentRuntimesInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, x := range page.AgentRuntimes {
			emit(StringValue(x.AgentRuntimeId), StringValue(x.AgentRuntimeId), "aws_bedrockagentcore_agent_runtime")
		}
	}
	for p := bedrockagentcorecontrol.NewListBrowsersPaginator(svc, &bedrockagentcorecontrol.ListBrowsersInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range page.BrowserSummaries {
			emit(StringValue(x.BrowserId), StringValue(x.BrowserId), "aws_bedrockagentcore_browser")
		}
	}
	for p := bedrockagentcorecontrol.NewListCodeInterpretersPaginator(svc, &bedrockagentcorecontrol.ListCodeInterpretersInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range page.CodeInterpreterSummaries {
			emit(StringValue(x.CodeInterpreterId), StringValue(x.CodeInterpreterId), "aws_bedrockagentcore_code_interpreter")
		}
	}
	for p := bedrockagentcorecontrol.NewListGatewaysPaginator(svc, &bedrockagentcorecontrol.ListGatewaysInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range page.Items {
			emit(StringValue(x.GatewayId), StringValue(x.Name), "aws_bedrockagentcore_gateway")
		}
	}
	for p := bedrockagentcorecontrol.NewListMemoriesPaginator(svc, &bedrockagentcorecontrol.ListMemoriesInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range page.Memories {
			emit(StringValue(x.Id), StringValue(x.Id), "aws_bedrockagentcore_memory")
		}
	}
	for p := bedrockagentcorecontrol.NewListWorkloadIdentitiesPaginator(svc, &bedrockagentcorecontrol.ListWorkloadIdentitiesInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range page.WorkloadIdentities {
			emit(StringValue(x.Name), StringValue(x.Name), "aws_bedrockagentcore_workload_identity")
		}
	}
	for p := bedrockagentcorecontrol.NewListEvaluatorsPaginator(svc, &bedrockagentcorecontrol.ListEvaluatorsInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range page.Evaluators {
			emit(StringValue(x.EvaluatorId), StringValue(x.EvaluatorId), "aws_bedrockagentcore_evaluator")
		}
	}
	for p := bedrockagentcorecontrol.NewListPoliciesPaginator(svc, &bedrockagentcorecontrol.ListPoliciesInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range page.Policies {
			emit(StringValue(x.PolicyArn), StringValue(x.PolicyArn), "aws_bedrockagentcore_policy")
		}
	}
	for p := bedrockagentcorecontrol.NewListPolicyEnginesPaginator(svc, &bedrockagentcorecontrol.ListPolicyEnginesInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range page.PolicyEngines {
			emit(StringValue(x.PolicyEngineArn), StringValue(x.PolicyEngineArn), "aws_bedrockagentcore_policy_engine")
		}
	}
	for p := bedrockagentcorecontrol.NewListHarnessesPaginator(svc, &bedrockagentcorecontrol.ListHarnessesInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range page.Harnesses {
			emit(StringValue(x.HarnessId), StringValue(x.HarnessId), "aws_bedrockagentcore_harness")
		}
	}
	for p := bedrockagentcorecontrol.NewListApiKeyCredentialProvidersPaginator(svc, &bedrockagentcorecontrol.ListApiKeyCredentialProvidersInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range page.CredentialProviders {
			emit(StringValue(x.Name), StringValue(x.Name), "aws_bedrockagentcore_api_key_credential_provider")
		}
	}
	for p := bedrockagentcorecontrol.NewListOnlineEvaluationConfigsPaginator(svc, &bedrockagentcorecontrol.ListOnlineEvaluationConfigsInput{}); p.HasMorePages(); {
		page, err := p.NextPage(ctx)
		if err != nil {
			break
		}
		for _, x := range page.OnlineEvaluationConfigs {
			emit(StringValue(x.OnlineEvaluationConfigArn), StringValue(x.OnlineEvaluationConfigArn), "aws_bedrockagentcore_online_evaluation_config")
		}
	}
	return nil
}
