package propel

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/propeldata/terraform-provider-propel/propel/internal"
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
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
				Description: "The Data Source that the Data Pool belongs to.",
			},
			"table": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
				Description: "The name of the Data Pool's table.",
			},
			"column": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
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
						"clickhouse_type": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "The ClickHouse type to use when `type` is set to `CLICKHOUSE`.",
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
				Deprecated:  "Use Data Pool Access Policies instead. This attribute will be removed in the next major version of the provider.",
			},
			"timestamp": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Data Pool's timestamp column.",
			},
			"unique_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The Data Pool's unique ID column. Propel uses the primary timestamp and a unique ID to compose a primary key for determining whether records should be inserted, deleted, or updated within the Data Pool. Only for Snowflake Data Pools.",
				Deprecated:  "Use Table Settings to define the primary key. This attribute will be removed in the next major version of the provider.",
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
				Default:     false,
				Description: "Whether the Data Pool has access control enabled or not. If the Data Pool has access control enabled, Applications must be assigned Data Pool Access Policies in order to query the Data Pool and its Metrics.",
			},
			"table_settings": internal.TableSettingsSchema(),
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

		columnClickHouseType := column["clickhouse_type"].(string)
		if columnClickHouseType != "" {
			columns[i].ClickHouseType = &columnClickHouseType
		}
	}

	return columns
}

func resourceDataPoolCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(graphql.Client)

	accessControlEnabled := d.Get("access_control_enabled").(bool)

	columns := make([]*pc.DataPoolColumnInput, 0)
	if def, ok := d.Get("column").([]any); ok && len(def) > 0 {
		columns = expandDataPoolColumns(def)
	}

	input := &pc.CreateDataPoolInputV2{
		Columns:              columns,
		AccessControlEnabled: &accessControlEnabled,
	}

	if t, exists := d.GetOk("unique_name"); exists && t.(string) != "" {
		uniqueName := t.(string)
		input.UniqueName = &uniqueName
	}

	if t, exists := d.GetOk("description"); exists && t.(string) != "" {
		description := t.(string)
		input.Description = &description
	}

	if t, exists := d.GetOk("data_source"); exists && t.(string) != "" {
		dataSourceId := t.(string)
		input.DataSource = &dataSourceId
	}

	if t, exists := d.GetOk("table"); exists && t.(string) != "" {
		table := t.(string)
		input.Table = &table
	}

	if v, exists := d.GetOk("tenant_id"); exists && v.(string) != "" {
		input.Tenant = &pc.TenantInput{
			ColumnName: v.(string),
		}
	}

	if v, exists := d.GetOk("timestamp"); exists && v.(string) != "" {
		input.Timestamp = &pc.TimestampInput{
			ColumnName: d.Get("timestamp").(string),
		}
	}

	if v, exists := d.GetOk("unique_id"); exists && v.(string) != "" {
		input.UniqueId = &pc.UniqueIdInput{
			ColumnName: v.(string),
		}
	}

	if v, exists := d.GetOk("table_settings.0"); exists {
		s, err := internal.BuildTableSettingsInput(v.(map[string]any))
		if err != nil {
			return diag.FromErr(err)
		}

		input.TableSettings = s
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

	if err := internal.WaitForDataPoolLive(ctx, c, d.Id(), timeout); err != nil {
		return diag.FromErr(err)
	}

	return resourceDataPoolRead(ctx, d, meta)
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

	if response.DataPool.DataSource != nil {
		if err := d.Set("data_source", response.DataPool.DataSource.Id); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("table", response.DataPool.Table); err != nil {
		return diag.FromErr(err)
	}

	if response.DataPool.Timestamp != nil {
		if err := d.Set("timestamp", response.DataPool.Timestamp.ColumnName); err != nil {
			return diag.FromErr(err)
		}
	}

	if response.DataPool.Tenant != nil {
		if err := d.Set("tenant_id", response.DataPool.Tenant.ColumnName); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("access_control_enabled", response.DataPool.AccessControlEnabled); err != nil {
		return diag.FromErr(err)
	}

	if response.DataPool.UniqueId != nil {
		if err := d.Set("unique_id", response.DataPool.UniqueId.ColumnName); err != nil {
			return diag.FromErr(err)
		}
	}

	if response.DataPool.TableSettings != nil {
		if err := d.Set("table_settings", []map[string]any{internal.ParseTableSettings(response.DataPool.TableSettings.TableSettingsData)}); err != nil {
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

	columnNodes := response.DataPool.Columns.Nodes
	if len(columnNodes) > 0 {
		columns := make([]any, 0, len(columnNodes))

		for _, node := range columnNodes {
			columns = append(columns, map[string]any{
				"name":            node.GetColumnName(),
				"type":            node.GetType(),
				"clickhouse_type": node.GetClickHouseType(),
				"nullable":        node.GetIsNullable(),
			})
		}

		if err := d.Set("column", (any)(columns)); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceDataPoolUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(graphql.Client)
	id := d.Id()
	input := &pc.ModifyDataPoolInput{
		IdOrUniqueName: &pc.IdOrUniqueName{Id: &id},
	}

	if d.HasChanges("unique_name", "description", "access_control_enabled", "timestamp") {
		uniqueName := d.Get("unique_name").(string)
		description := d.Get("description").(string)
		accessControlEnabled := d.Get("access_control_enabled").(bool)

		input.UniqueName = &uniqueName
		input.Description = &description
		input.AccessControlEnabled = &accessControlEnabled
		input.Timestamp = &pc.TimestampInput{ColumnName: d.Get("timestamp").(string)}
	}

	if d.HasChange("syncing") {
		if _, exists := d.GetOk("syncing"); exists {
			syncing := d.Get("syncing").([]any)[0].(map[string]any)

			input.Syncing = &pc.DataPoolSyncingInput{
				Interval: pc.DataPoolSyncInterval(syncing["interval"].(string)),
			}
		}
	}

	_, err := pc.ModifyDataPool(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("column") {
		oldItem, newItem := d.GetChange("column")
		oldDef, oldOk := oldItem.([]any)
		newDef, newOk := newItem.([]any)

		if !oldOk || !newOk {
			diag.FromErr(errors.New("invalid column format"))
		}

		newColumns, err := getNewDataPoolColumns(oldDef, newDef)
		if err != nil {
			return diag.FromErr(err)
		}

		if err = addNewDataPoolColumns(ctx, d, c, id, newColumns); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceDataPoolRead(ctx, d, m)
}

func getNewDataPoolColumns(oldItemDef []any, newItemDef []any) (map[string]pc.DataPoolColumnInput, error) {
	newColumns := map[string]pc.DataPoolColumnInput{}

	for _, rawColumn := range newItemDef {
		column := rawColumn.(map[string]any)
		columnInput := pc.DataPoolColumnInput{
			ColumnName: column["name"].(string),
			Type:       pc.ColumnType(column["type"].(string)),
			IsNullable: column["nullable"].(bool),
		}

		columnClickHouseType := column["clickhouse_type"].(string)
		if columnClickHouseType != "" {
			columnInput.ClickHouseType = &columnClickHouseType
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

		columnClickHouseType := column["clickhouse_type"].(string)
		if columnClickHouseType != "" {
			columnInput.ClickHouseType = &columnClickHouseType
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

func addNewDataPoolColumns(ctx context.Context, d *schema.ResourceData, c graphql.Client, dataPoolId string, newColumns map[string]pc.DataPoolColumnInput) error {
	for _, newColumn := range newColumns {
		if !newColumn.IsNullable {
			return fmt.Errorf(`new column "%s" must be nullable`, newColumn.ColumnName)
		}

		response, err := pc.CreateAddColumnToDataPoolJob(ctx, c, &pc.CreateAddColumnToDataPoolJobInput{
			DataPool:             dataPoolId,
			ColumnName:           newColumn.ColumnName,
			ColumnType:           newColumn.Type,
			ColumnClickHouseType: newColumn.ClickHouseType,
		})
		if err != nil {
			return err
		}

		timeout := d.Timeout(schema.TimeoutUpdate)

		if err = internal.WaitForAddColumnJobSucceeded(ctx, c, response.CreateAddColumnToDataPoolJob.Job.Id, timeout); err != nil {
			return err
		}
	}

	return nil
}

func resourceDataPoolDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(graphql.Client)

	var diags diag.Diagnostics

	if _, err := pc.DeleteDataPool(ctx, c, d.Id()); err != nil {
		return diag.FromErr(err)
	}

	timeout := d.Timeout(schema.TimeoutDelete)
	if err := internal.WaitForDataPoolDeletion(ctx, c, d.Id(), timeout); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
