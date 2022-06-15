package propel

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	pc "github.com/propeldata/terraform-provider/graphql_client"
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
				Description: "The DataPool name",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The DataPool description",
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

	switch resource := response.GetCreateDataPool().(type) {
	case *pc.CreateDataPoolCreateDataPoolDataPoolResponse:
		d.SetId(resource.DataPool.Id)

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

			return fmt.Errorf("error trying to fetch DataPool: %s", err)
		}

		n++
	}
	return nil
}
