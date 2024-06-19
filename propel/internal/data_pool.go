package internal

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	pc "github.com/propeldata/terraform-provider-propel/propel_client"
)

func WaitForDataPoolLive(ctx context.Context, client graphql.Client, id string, timeout time.Duration) error {
	createStateConf := &retry.StateChangeConf{
		Pending: []string{
			string(pc.DataPoolStatusCreated),
			string(pc.DataPoolStatusPending),
		},
		Target: []string{
			string(pc.DataPoolStatusLive),
		},
		Refresh: func() (any, string, error) {
			resp, err := pc.DataPool(ctx, client, id)
			if err != nil {
				return 0, "", fmt.Errorf("error trying to read Data Pool status: %s", err)
			}

			return resp, string(resp.DataPool.Status), nil
		},
		Timeout:                   timeout - time.Minute,
		Delay:                     10 * time.Second,
		MinTimeout:                5 * time.Second,
		ContinuousTargetOccurence: 3,
	}

	_, err := createStateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf("error waiting for Data Pool to be LIVE: %s", err)
	}

	return nil
}

func WaitForDataPoolDeletion(ctx context.Context, client graphql.Client, id string, timeout time.Duration) error {
	tickerInterval := 10 // 10s
	timeoutSeconds := int(timeout.Seconds())
	n := 0

	ticker := time.NewTicker(time.Duration(tickerInterval) * time.Second)
	for range ticker.C {
		if n*tickerInterval > timeoutSeconds {
			ticker.Stop()
			break
		}

		_, err := pc.DataPool(ctx, client, id)
		if err != nil {
			ticker.Stop()

			if strings.Contains(err.Error(), "not found") {
				return nil
			}

			return fmt.Errorf("error trying to fetch Data Pool: %s", err)
		}

		n++
	}
	return nil
}
