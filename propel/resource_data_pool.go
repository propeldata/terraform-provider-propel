package propel

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	pc "github.com/propeldata/terraform-provider/propel_client"
)

func resourceDataPool() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDataPoolCreate,
		ReadContext:   resourceDataPoolRead,
		UpdateContext: resourceDataPoolUpdate,
		DeleteContext: resourceDataPoolDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"unique_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The Data Pool name",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The Data Pool description",
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"account": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"environment": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Environment where belong the Data Source",
			},
			"datasource": {
				Type:     schema.TypeString,
				Required: true,
			},
			"table": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"timestamp": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceDataPoolCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(graphql.Client)

	var diags diag.Diagnostics

	response, err := pc.CreateDataPool(ctx, c, pc.CreateDataPoolInput{
		UniqueName:  d.Get("unique_name").(string),
		Description: d.Get("description").(string),
		DataSource: pc.IdOrUniqueName{
			Id: d.Get("datasource").(string),
		},
		Table: d.Get("table").(string),
		Timestamp: pc.DimensionInput{
			ColumnName: d.Get("timestamp").(string),
		},
	})

	if err != nil {
		return diag.FromErr(err)
	}

	switch r := response.GetCreateDataPool().(type) {
	case *pc.CreateDataPoolCreateDataPoolDataPoolResponse:
		d.SetId(r.DataPool.Id)

		timeout := d.Timeout(schema.TimeoutCreate)

		err = waitForDataPoolLive(ctx, c, d.Id(), timeout)
		if err != nil {
			return diag.FromErr(err)
		}

		resourceDataPoolRead(ctx, d, meta)
	case *pc.CreateDataPoolCreateDataPoolFailureResponse:
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to create DataPool",
		})
	}

	return diags
}

func resourceDataPoolRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(graphql.Client)

	var diags diag.Diagnostics

	response, err := pc.DataPool(ctx, c, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(response.DataPool.Id)
	if err := d.Set("unique_name", response.DataPool.UniqueName); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("description", response.DataPool.Description); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("status", response.DataPool.Status); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("environment", response.DataPool.Environment.Id); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("account", response.DataPool.Account.Id); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("datasource", response.DataPool.DataSource.Id); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("table", response.DataPool.Table); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("timestamp", response.DataPool.Timestamp.ColumnName); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceDataPoolUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(graphql.Client)

	if d.HasChanges("unique_name", "description") {
		input := pc.ModifyDataPoolInput{
			IdOrUniqueName: pc.IdOrUniqueName{
				Id: d.Id(),
			},
			UniqueName:  d.Get("unique_name").(string),
			Description: d.Get("description").(string),
		}

		_, err := pc.ModifyDataPool(ctx, c, input)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceDataPoolRead(ctx, d, m)
}

func resourceDataPoolDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(graphql.Client)

	var diags diag.Diagnostics

	_, err := pc.DeleteDataPool(ctx, c, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	timeout := d.Timeout(schema.TimeoutDelete)
	err = waitForDataPoolDeletion(ctx, c, d.Id(), timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}

func waitForDataPoolLive(ctx context.Context, client graphql.Client, id string, timeout time.Duration) error {
	createStateConf := &resource.StateChangeConf{
		Pending: []string{
			string(pc.DataPoolStatusCreated),
			string(pc.DataPoolStatusPending),
		},
		Target: []string{
			string(pc.DataPoolStatusLive),
		},
		Refresh: func() (interface{}, string, error) {
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

func waitForDataPoolDeletion(ctx context.Context, client graphql.Client, id string, timeout time.Duration) error {
	ticketInterval := 10 // 10s
	timeoutSeconds := int(timeout.Seconds())
	n := 0

	ticker := time.NewTicker(time.Duration(ticketInterval) * time.Second)
	for range ticker.C {
		if n*ticketInterval > timeoutSeconds {
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
