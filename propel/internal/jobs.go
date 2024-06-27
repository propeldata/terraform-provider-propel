package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	pc "github.com/propeldata/terraform-provider-propel/propel_client"
)

func WaitForAddColumnJobSucceeded(ctx context.Context, client graphql.Client, id string, timeout time.Duration) error {
	createStateConf := &retry.StateChangeConf{
		Pending: []string{
			string(pc.JobStatusCreated),
			string(pc.JobStatusInProgress),
		},
		Target: []string{
			string(pc.JobStatusSucceeded),
			string(pc.JobStatusFailed),
		},
		Refresh: func() (any, string, error) {
			resp, err := pc.AddColumnToDataPoolJob(ctx, client, id)
			if err != nil {
				return 0, "", fmt.Errorf("error trying to read Add Column Job status: %s", err)
			}

			return resp, string(resp.AddColumnToDataPoolJob.Status), nil
		},
		Timeout:                   timeout - time.Minute,
		Delay:                     10 * time.Second,
		MinTimeout:                5 * time.Second,
		ContinuousTargetOccurence: 3,
	}

	resp, err := createStateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf("error waiting for Add Column Job to succeed: %s", err)
	}

	addColumnJobResponse, ok := resp.(*pc.AddColumnToDataPoolJobResponse)
	if !ok || addColumnJobResponse.AddColumnToDataPoolJob.Status == pc.JobStatusFailed {
		return fmt.Errorf("add column job failed: %s", addColumnJobResponse.AddColumnToDataPoolJob.Error.Message)
	}

	return nil
}
