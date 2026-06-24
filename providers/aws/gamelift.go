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
	"github.com/aws/aws-sdk-go-v2/service/gamelift"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type GameLiftGenerator struct {
	AWSService
}

// InitResources enumerates GameLift fleets and builds. Import IDs are the ids.
func (g *GameLiftGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := gamelift.NewFromConfig(config)
	ctx := awsContext()

	fleets := gamelift.NewListFleetsPaginator(svc, &gamelift.ListFleetsInput{})
	for fleets.HasMorePages() {
		page, err := fleets.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, id := range page.FleetIds {
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_gamelift_fleet", "aws", defaultAllowEmptyValues))
		}
	}

	builds := gamelift.NewListBuildsPaginator(svc, &gamelift.ListBuildsInput{})
	for builds.HasMorePages() {
		page, err := builds.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, build := range page.Builds {
			id := StringValue(build.BuildId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(build.Name), "aws_gamelift_build", "aws", defaultAllowEmptyValues))
		}
	}

	for aliases := gamelift.NewListAliasesPaginator(svc, &gamelift.ListAliasesInput{}); aliases.HasMorePages(); {
		page, err := aliases.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, a := range page.Aliases {
			id := StringValue(a.AliasId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(a.Name), "aws_gamelift_alias", "aws", defaultAllowEmptyValues))
		}
	}

	for groups := gamelift.NewListGameServerGroupsPaginator(svc, &gamelift.ListGameServerGroupsInput{}); groups.HasMorePages(); {
		page, err := groups.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, grp := range page.GameServerGroups {
			name := StringValue(grp.GameServerGroupName)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_gamelift_game_server_group", "aws", defaultAllowEmptyValues))
		}
	}

	for queues := gamelift.NewDescribeGameSessionQueuesPaginator(svc, &gamelift.DescribeGameSessionQueuesInput{}); queues.HasMorePages(); {
		page, err := queues.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, q := range page.GameSessionQueues {
			name := StringValue(q.Name)
			if name == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_gamelift_game_session_queue", "aws", defaultAllowEmptyValues))
		}
	}

	for scripts := gamelift.NewListScriptsPaginator(svc, &gamelift.ListScriptsInput{}); scripts.HasMorePages(); {
		page, err := scripts.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, s := range page.Scripts {
			id := StringValue(s.ScriptId)
			if id == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(s.Name), "aws_gamelift_script", "aws", defaultAllowEmptyValues))
		}
	}
	return nil
}
