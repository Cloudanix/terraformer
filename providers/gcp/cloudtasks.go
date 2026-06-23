package gcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"google.golang.org/api/iterator"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	taskspb "cloud.google.com/go/cloudtasks/apiv2/cloudtaskspb"
	iampb "cloud.google.com/go/iam/apiv1/iampb"
	"google.golang.org/api/compute/v1"
)

var cloudTasksAllowEmptyValues = []string{}

var cloudTasksAdditionalFields = map[string]interface{}{}

type CloudTaskGenerator struct {
	GCPService
}

func (g *CloudTaskGenerator) loadCloudTaskQueues(ctx context.Context, client *cloudtasks.Client) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name

	req := &taskspb.ListQueuesRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", project, region),
	}

	queueIterator := client.ListQueues(ctx, req)
	for {
		resp, err := queueIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		splitName := strings.Split(resp.Name, "/")
		queueName := splitName[len(splitName)-1]

		g.Resources = append(g.Resources, terraformutils.NewResource(
			resp.Name,
			queueName,
			"google_cloud_tasks_queue",
			g.ProviderName,
			map[string]string{
				"id":       fmt.Sprintf("projects/%s/locations/%s/queues/%s", project, region, queueName),
				"name":     queueName,
				"project":  project,
				"location": region,
			},
			cloudTasksAllowEmptyValues,
			cloudTasksAdditionalFields,
		))
		if policy, perr := client.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{Resource: resp.Name}); perr == nil {
			for _, b := range policy.Bindings {
				for _, m := range b.Members {
					g.Resources = append(g.Resources, terraformutils.NewResource(
						resp.Name+" "+b.Role+" "+m, queueName+"_"+b.Role+"_"+m,
						"google_cloud_tasks_queue_iam_member", g.ProviderName,
						map[string]string{"name": queueName, "role": b.Role, "member": m, "project": project, "location": region},
						cloudTasksAllowEmptyValues, cloudTasksAdditionalFields))
				}
			}
		}
	}
	return nil
}

// Generate TerraformResources from GCP API,
// from each cloud task queue create 1 TerraformResource
func (g *CloudTaskGenerator) InitResources() error {
	ctx := context.Background()
	client, err := cloudtasks.NewClient(ctx)
	if err != nil {
		return err
	}

	if err := g.loadCloudTaskQueues(ctx, client); err != nil {
		return err
	}

	return client.Close()
}
