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

	"github.com/propeldata/terraform-provider-propel/propel/internal/utils"
	pc "github.com/propeldata/terraform-provider-propel/propel_client"
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
		Description:   "Provides a Propel Data Pool resource. This can be used to create and manage Propel Data Pools.",
		Schema: map[string]*schema.Schema{
			"unique_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The Data Pool's name.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The Data Pool's description.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Data Pool's status.",
			},
			"account": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Account that the Data Pool belongs to.",
			},
			"environment": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Environment that the Data Pool belongs to.",
			},
			"data_source": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Data Source that the Data Pool belongs to.",
			},
			"table": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the Data Pool's table.",
			},
			"column": {
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    false,
				Description: "The list of columns, their types and nullability.",
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
			"tenant_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The tenant ID for restricting access between customers.",
			},
			"timestamp": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The Data Pool's timestamp column.",
			},
			"unique_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The Data Pool's unique ID column. Propel uses the primary timestamp and a unique ID to compose a primary key for determining whether records should be inserted, deleted, or updated within the Data Pool. Only for Snowflake Data Pools.",
			},
			"syncing": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Description: "The Data Pool's syncing settings.",
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Indicates whether syncing is enabled or disabled.",
						},
						"interval": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The syncing interval.",
						},
						"last_synced_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The date and time of the most recent Sync in UTC.",
						},
					},
				},
			},
			"access_control_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether the Data Pool has access control enabled or not. If the Data Pool has access control enabled, Applications must be assigned Data Pool Access Policies in order to query the Data Pool and its Metrics.",
			},
		},
	}
}

func expandDataPoolColumns(def []any) []*pc.DataPoolColumnInput {
	columns := make([]*pc.DataPoolColumnInput, len(def))

	for i, rawColumn := range def {
		column := rawColumn.(map[string]any)

		columns[i] = &pc.DataPoolColumnInput{
			ColumnName: column["name"].(string),
			Type:       pc.ColumnType(column["type"].(string)),
			IsNullable: column["nullable"].(bool),
		}
	}

	return columns
}

func resourceDataPoolCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(graphql.Client)

	var diags diag.Diagnostics

	id := d.Get("data_source").(string)
	uniqueName := d.Get("unique_name").(string)
	description := d.Get("description").(string)
	accessControlEnabled := d.Get("access_control_enabled").(bool)

	columns := make([]*pc.DataPoolColumnInput, 0)
	if def, ok := d.Get("column").([]any); ok && len(def) > 0 {
		columns = expandDataPoolColumns(def)
	}

	input := &pc.CreateDataPoolInputV2{
		UniqueName:  &uniqueName,
		Description: &description,
		DataSource:  id,
		Table:       d.Get("table").(string),
		Timestamp: &pc.TimestampInput{
			ColumnName: d.Get("timestamp").(string),
		},
		Columns:              columns,
		AccessControlEnabled: &accessControlEnabled,
	}

	if _, exists := d.GetOk("tenant_id"); exists {
		input.Tenant = &pc.TenantInput{
			ColumnName: d.Get("tenant_id").(string),
		}
	}

	if _, exists := d.GetOk("unique_id"); exists {
		input.UniqueId = &pc.UniqueIdInput{
			ColumnName: d.Get("unique_id").(string),
		}
	}

	if _, exists := d.GetOk("syncing"); exists {
		syncing := d.Get("syncing").([]any)[0].(map[string]any)

		input.Syncing = &pc.DataPoolSyncingInput{
			Interval: pc.DataPoolSyncInterval(syncing["interval"].(string)),
		}
	}

	response, err := pc.CreateDataPool(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(response.CreateDataPoolV2.DataPool.Id)

	timeout := d.Timeout(schema.TimeoutCreate)

	err = waitForDataPoolLive(ctx, c, d.Id(), timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	resourceDataPoolRead(ctx, d, meta)

	return diags
}

func resourceDataPoolRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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

	if err := d.Set("data_source", response.DataPool.DataSource.Id); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("table", response.DataPool.Table); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("timestamp", response.DataPool.Timestamp.ColumnName); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("access_control_enabled", response.DataPool.AccessControlEnabled); err != nil {
		return diag.FromErr(err)
	}

	if response.DataPool.UniqueId != nil {
		if err := d.Set("unique_id", response.DataPool.UniqueId.ColumnName); err != nil {
			return diag.FromErr(err)
		}
	}

	syncing := map[string]any{
		"status":   response.DataPool.Syncing.GetStatus(),
		"interval": response.DataPool.Syncing.GetInterval(),
	}

	lastSyncedAt := response.DataPool.Syncing.GetLastSyncedAt()
	if lastSyncedAt != nil {
		syncing["last_synced_at"] = lastSyncedAt.Format(time.RFC3339)
	}

	if err := d.Set("syncing", []map[string]any{syncing}); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceDataPoolUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(graphql.Client)
	id := d.Id()

	if d.HasChanges("unique_name", "description", "syncing", "access_control_enabled") {
		uniqueName := d.Get("unique_name").(string)
		description := d.Get("description").(string)
		accessControlEnabled := d.Get("access_control_enabled").(bool)
		input := &pc.ModifyDataPoolInput{
			IdOrUniqueName: &pc.IdOrUniqueName{
				Id: &id,
			},
			UniqueName:           &uniqueName,
			Description:          &description,
			AccessControlEnabled: &accessControlEnabled,
		}

		if _, exists := d.GetOk("syncing"); exists {
			syncing := d.Get("syncing").([]any)[0].(map[string]any)

			input.Syncing = &pc.DataPoolSyncingInput{
				Interval: pc.DataPoolSyncInterval(syncing["interval"].(string)),
			}
		}

		_, err := pc.ModifyDataPool(ctx, c, input)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChanges("column") {
		oldItem, newItem := d.GetChange("column")
		oldDef, oldOk := oldItem.([]any)
		newDef, newOk := newItem.([]any)

		if oldOk && newOk {
			newColumns, err := getNewColumns(oldDef, newDef)
			if err != nil {
				return diag.FromErr(err)
			}

			for _, newColumn := range newColumns {
				if !newColumn.IsNullable {
					return diag.FromErr(fmt.Errorf(`new column "%s" must be nullable`, newColumn.ColumnName))
				}

				_, err := pc.CreateAddColumnToDataPoolJob(ctx, c, &pc.CreateAddColumnToDataPoolJobInput{
					DataPool:   id,
					ColumnName: newColumn.ColumnName,
					ColumnType: newColumn.Type,
				})
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}
	}

	return resourceDataPoolRead(ctx, d, m)
}

func getNewColumns(oldItemDef []any, newItemDef []any) (map[string]pc.DataPoolColumnInput, error) {
	newColumns := map[string]pc.DataPoolColumnInput{}

	for _, rawColumn := range newItemDef {
		column := rawColumn.(map[string]any)
		columnInput := pc.DataPoolColumnInput{
			ColumnName: column["name"].(string),
			Type:       pc.ColumnType(column["type"].(string)),
			IsNullable: column["nullable"].(bool),
		}

		if _, ok := newColumns[columnInput.ColumnName]; ok {
			return nil, fmt.Errorf(`column "%s" already exists`, columnInput.ColumnName)
		}

		newColumns[columnInput.ColumnName] = columnInput
	}

	for _, rawColumn := range oldItemDef {
		column := rawColumn.(map[string]any)
		columnInput := pc.DataPoolColumnInput{
			ColumnName: column["name"].(string),
			Type:       pc.ColumnType(column["type"].(string)),
			IsNullable: column["nullable"].(bool),
		}

		newColumnInput, ok := newColumns[columnInput.ColumnName]
		if !ok {
			return nil, fmt.Errorf(`column "%s" was removed, column deletions are not supported`, columnInput.ColumnName)
		}

		if columnInput.Type != newColumnInput.Type || columnInput.IsNullable != newColumnInput.IsNullable {
			return nil, fmt.Errorf(`column "%s" was modified, column updates are not supported`, columnInput.ColumnName)
		}

		delete(newColumns, columnInput.ColumnName)
	}

	return newColumns, nil
}

func resourceDataPoolDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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

func waitForDataPoolDeletion(ctx context.Context, client graphql.Client, id string, timeout time.Duration) error {
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
