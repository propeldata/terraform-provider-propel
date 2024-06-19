package propel

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/propeldata/terraform-provider-propel/propel/internal"
	"github.com/propeldata/terraform-provider-propel/propel/internal/utils"
	pc "github.com/propeldata/terraform-provider-propel/propel_client"
)

func resourceDataSource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDataSourceCreate,
		ReadContext:   resourceDataSourceRead,
		UpdateContext: resourceDataSourceUpdate,
		DeleteContext: resourceDataSourceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Description:   "Provides a Propel Data Source resource. This can be used to create and manage Propel Data Sources.",
		Schema: map[string]*schema.Schema{
			"unique_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The Data Source's name.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The Data Source's description.",
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"SNOWFLAKE",
					"S3",
					"HTTP",
					"WEBHOOK",
					"KAFKA",
					"CLICKHOUSE",
				}, true),
				Description: "The Data Source's type. Depending on this, you will need to specify one of `snowflake_connection_settings`, `s3_connection_settings`, `http_connection_settings`, `webhook_connection_settings`, `kafka_connection_settings` or `clickhouse_connection_settings`. The valid values are `SNOWFLAKE`, `S3`, `HTTP`, `WEBHOOK`, `KAFKA` and `CLICKHOUSE`",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Data Source's status.",
			},
			"account": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Account that the Data Source belongs to.",
			},
			"environment": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Environment that the Data Source belongs to",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of when the Data Source was created.",
			},
			"modified_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of when the Data Source was modified.",
			},
			"created_by": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The user who created the Data Source.",
			},
			"modified_by": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The user who modified the Data Source.",
			},
			"snowflake_connection_settings":  internal.SnowflakeDataSourceSchema(),
			"http_connection_settings":       internal.HttpDataSourceSchema(),
			"s3_connection_settings":         internal.S3DataSourceSchema(),
			"webhook_connection_settings":    internal.WebhookDataSourceSchema(),
			"kafka_connection_settings":      internal.KafkaDataSourceSchema(),
			"clickhouse_connection_settings": internal.ClickHouseDataSourceSchema(),
			"table": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Description: "Specify an HTTP or S3 Data Source's tables with this. You do not need to use this for Snowflake Data Sources, since Snowflake Data Sources' tables are automatically introspected. You do not need to use this for Webhook Data Sources, since table is automatically created.",
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The table's ID.",
						},
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the table.",
						},
						"path": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The path to the table's files in S3.",
						},
						"column": {
							Type:        schema.TypeList,
							Required:    true,
							ForceNew:    false,
							Description: "Specify a table's columns.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The column name.",
									},
									"type": {
										Type:         schema.TypeString,
										Required:     true,
										Description:  "The column type.",
										ValidateFunc: utils.IsValidColumnType,
									},
									"nullable": {
										Type:        schema.TypeBool,
										Required:    true,
										Description: "Whether the column's type is nullable or not.",
									},
								},
							},
						},
					},
				},
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
	}
}

func resourceDataSourceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var id string
	var err error

	c := meta.(graphql.Client)
	dataSourceType := d.Get("type").(string)

	switch strings.ToUpper(dataSourceType) {
	case "SNOWFLAKE":
		id, err = internal.SnowflakeDataSourceCreate(ctx, d, c)
	case "HTTP":
		id, err = internal.HttpDataSourceCreate(ctx, d, c)
	case "S3":
		id, err = internal.S3DataSourceCreate(ctx, d, c)
	case "WEBHOOK":
		id, err = internal.WebhookDataSourceCreate(ctx, d, c)
	case "KAFKA":
		id, err = internal.KafkaDataSourceCreate(ctx, d, c)
	case "CLICKHOUSE":
		id, err = internal.ClickHouseDataSourceCreate(ctx, d, c)
	default:
		err = fmt.Errorf("unsupported Data Source type \"%v\"", dataSourceType)
	}

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)

	timeout := d.Timeout(schema.TimeoutCreate)
	if err := waitForDataSourceConnected(ctx, c, id, timeout); err != nil {
		return diag.FromErr(err)
	}

	return resourceDataSourceRead(ctx, d, meta)
}

func resourceDataSourceRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(graphql.Client)

	response, err := pc.DataSource(ctx, c, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(response.DataSource.Id)

	if err := d.Set("unique_name", response.DataSource.GetUniqueName()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("description", response.DataSource.GetDescription()); err != nil {
		return diag.FromErr(err)
	}

	dataSourceType := string(response.DataSource.GetType())
	if err := d.Set("type", strings.ToUpper(dataSourceType)); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("created_at", response.DataSource.GetCreatedAt().String()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("created_by", response.DataSource.GetCreatedBy()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("modified_at", response.DataSource.GetModifiedAt().String()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("modified_by", response.DataSource.GetModifiedBy()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("environment", response.DataSource.GetEnvironment().Id); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("account", response.DataSource.GetAccount().Id); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("status", response.DataSource.GetStatus()); err != nil {
		return diag.FromErr(err)
	}

	switch strings.ToUpper(dataSourceType) {
	case "SNOWFLAKE":
		err = internal.HandleSnowflakeConnectionSettings(response, d)
	case "HTTP":
		err = internal.HandleHttpConnectionSettings(response, d)
	case "S3":
		err = internal.HandleS3ConnectionSettings(response, d)
	case "WEBHOOK":
		err = internal.HandleWebhookConnectionSettings(response, d)
	case "KAFKA":
		err = internal.HandleKafkaConnectionSettings(response, d)
	case "CLICKHOUSE":
		err = internal.HandleClickHouseConnectionSettings(response, d)
	default:
		err = fmt.Errorf("unsupported Data Source type \"%v\"", dataSourceType)
	}

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDataSourceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var err error
	c := meta.(graphql.Client)

	dataSourceType := d.Get("type").(string)
	switch strings.ToUpper(dataSourceType) {
	case "SNOWFLAKE":
		err = internal.SnowflakeDataSourceUpdate(ctx, d, c)
	case "HTTP":
		err = internal.HttpDataSourceUpdate(ctx, d, c)
	case "S3":
		err = internal.S3DataSourceUpdate(ctx, d, c)
	case "WEBHOOK":
		err = internal.WebhookDataSourceUpdate(ctx, d, c)
	case "KAFKA":
		err = internal.KafkaDataSourceUpdate(ctx, d, c)
	case "CLICKHOUSE":
		err = internal.ClickHouseDataSourceUpdate(ctx, d, c)
	default:
		err = fmt.Errorf("unsupported Data Source type \"%v\"", dataSourceType)
	}

	if err != nil {
		return diag.FromErr(err)
	}

	timeout := d.Timeout(schema.TimeoutCreate)

	if err := waitForDataSourceConnected(ctx, c, d.Id(), timeout); err != nil {
		return diag.FromErr(err)
	}
	return resourceDataSourceRead(ctx, d, meta)

}

func resourceDataSourceDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(graphql.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Deletes default Data Pool first
	if strings.ToUpper(d.Get("type").(string)) == "WEBHOOK" {
		if err := internal.WebhookDataSourceDelete(ctx, d, c); err != nil {
			return diag.FromErr(err)
		}
	}

	_, err := pc.DeleteDataSource(ctx, c, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	timeout := d.Timeout(schema.TimeoutDelete)
	if err = waitForDataSourceDeletion(ctx, c, d.Id(), timeout); err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}

func waitForDataSourceConnected(ctx context.Context, client graphql.Client, id string, timeout time.Duration) error {
	createStateConf := &retry.StateChangeConf{
		Pending: []string{
			string(pc.DataSourceStatusCreated),
			string(pc.DataSourceStatusConnecting),
		},
		Target: []string{
			string(pc.DataSourceStatusConnected),
		},
		Refresh: func() (any, string, error) {
			resp, err := pc.DataSource(ctx, client, id)
			if err != nil {
				return nil, "", fmt.Errorf("error trying to read Data Source status: %s", err)
			}

			return resp, string(resp.DataSource.Status), nil
		},
		Timeout:                   timeout - time.Minute,
		Delay:                     10 * time.Second,
		MinTimeout:                5 * time.Second,
		ContinuousTargetOccurence: 3,
	}

	_, err := createStateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf("error waiting for Data Source to be CONNECTED: %s", err)
	}

	return nil
}

func waitForDataSourceDeletion(ctx context.Context, client graphql.Client, id string, timeout time.Duration) error {
	tickerInterval := 10 // 10s
	timeoutSeconds := int(timeout.Seconds())
	n := 0

	ticker := time.NewTicker(time.Duration(tickerInterval) * time.Second)
	for range ticker.C {
		if n*tickerInterval > timeoutSeconds {
			ticker.Stop()
			break
		}

		_, err := pc.DataSource(ctx, client, id)
		if err != nil {
			ticker.Stop()

			if strings.Contains(err.Error(), "not found") {
				return nil
			}

			return fmt.Errorf("error trying to fetch Data Source: %s", err)
		}

		n++
	}
	return nil
}
