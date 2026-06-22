package aws

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/xray"
)

var xrayAllowEmptyValues = []string{"tags."}

type XrayGenerator struct {
	AWSService
}

func (g *XrayGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := xray.NewFromConfig(config)

	p := xray.NewGetSamplingRulesPaginator(svc, &xray.GetSamplingRulesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, samplingRule := range page.SamplingRuleRecords {
			// NOTE: Builtin rule with unmodifiable name and 10000 prirority (lowest)
			if *samplingRule.SamplingRule.RuleName != "Default" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					*samplingRule.SamplingRule.RuleName,
					*samplingRule.SamplingRule.RuleName,
					"aws_xray_sampling_rule",
					"aws",
					xrayAllowEmptyValues))
			}
		}
	}

	for gp := xray.NewGetGroupsPaginator(svc, &xray.GetGroupsInput{}); gp.HasMorePages(); {
		page, err := gp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, grp := range page.Groups {
			arn := StringValue(grp.GroupARN)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(grp.GroupName), "aws_xray_group", "aws", xrayAllowEmptyValues))
		}
	}

	// Encryption config is a region-level singleton; import ID is the region.
	if region := config.Region; region != "" {
		if _, err := svc.GetEncryptionConfig(context.TODO(), &xray.GetEncryptionConfigInput{}); err == nil {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				region, region, "aws_xray_encryption_config", "aws", xrayAllowEmptyValues))
		}
	}

	return nil
}
