package propel

import (
	"context"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/propeldata/terraform-provider-propel/propel/internal"
	pc "github.com/propeldata/terraform-provider-propel/propel_client"
)

func resourceMaterializedView() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMaterializedViewCreate,
		ReadContext:   resourceMaterializedViewRead,
		UpdateContext: resourceMaterializedViewUpdate,
		DeleteContext: resourceMaterializedViewDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Description:   "Provides a Propel Materialized View resource. This can be used to create and manage Propel Materialized Views.",
		Schema: map[string]*schema.Schema{
			"unique_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The Materialized View's name.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The Materialized View's description.",
			},
			"account": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Materialized View's Account.",
			},
			"environment": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Environment that the Materialized View belongs to.",
			},
			"sql": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The SQL that the Materialized View executes.",
			},
			"existing_data_pool": {
				Type:          schema.TypeList,
				Optional:      true,
				ForceNew:      true,
				Description:   "If specified, the Materialized View will target an existing Data Pool. Ensure the Data Pool's schema is compatible with your Materialized View's SQL statement.",
				ConflictsWith: []string{"new_data_pool"},
				MaxItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The ID of the Data Pool.",
						},
					},
				},
			},
			"new_data_pool": {
				Type:          schema.TypeList,
				Optional:      true,
				ForceNew:      true,
				Description:   "If specified, the Materialized View will create and target a new Data Pool. You can further customize the new Data Pool's engine settings.",
				ConflictsWith: []string{"existing_data_pool"},
				MaxItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"unique_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The Data Pool's unique name.",
						},
						"description": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The Data Pool's description.",
						},
						"timestamp": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Optionally specify the Data Pool's primary timestamp. This will influence the Data Pool's engine settings.",
						},
						"access_control_enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Enables or disables access control for the Data Pool. If the Data Pool has access control enabled, Applications must be assigned Data Pool Access Policies in order to query the Data Pool and its Metrics.",
						},
						"table_settings": internal.TableSettingsSchema(),
					},
				},
			},
			"backfill": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Description: "Whether historical data should be backfilled or not.",
			},
			"destination": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Materialized View's destination (AKA \"target\") Data Pool.",
			},
			"source": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Materialized View's source Data Pool.",
			},
			"others": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Other Data Pools queried by the Materialized View.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceMaterializedViewCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(graphql.Client)

	uniqueName := d.Get("unique_name").(string)
	description := d.Get("description").(string)
	backfill := d.Get("backfill").(bool)
	destination := &pc.CreateMaterializedViewDestinationInput{}

	if v, exists := d.GetOk("existing_data_pool.0.id"); exists && v.(string) != "" {
		id := d.Get("existing_data_pool.0.id").(string)
		destination.ExistingDataPool = &pc.DataPoolInput{Id: &id}
	}

	if v, exists := d.GetOk("new_data_pool.0"); exists {
		destination.NewDataPool = &pc.CreateMaterializedViewDestinationNewDataPoolInput{}
		attrs := v.(map[string]any)

		if v, ok := attrs["unique_name"]; ok && v.(string) != "" {
			dpUniqueName := attrs["unique_name"].(string)
			destination.NewDataPool.UniqueName = &dpUniqueName
		}

		if v, ok := attrs["description"]; ok && v.(string) != "" {
			dpDescription := attrs["description"].(string)
			destination.NewDataPool.Description = &dpDescription
		}

		if v, ok := attrs["timestamp"]; ok && v.(string) != "" {
			destination.NewDataPool.Timestamp = &pc.TimestampInput{ColumnName: attrs["timestamp"].(string)}
		}

		if v, ok := attrs["unique_id"]; ok && v.(string) != "" {
			destination.NewDataPool.UniqueId = &pc.UniqueIdInput{ColumnName: attrs["unique_id"].(string)}
		}

		if _, ok := attrs["access_control_enabled"]; ok {
			accessControl := attrs["access_control_enabled"].(bool)
			destination.NewDataPool.AccessControlEnabled = &accessControl
		}

		if t, ok := attrs["table_settings"]; ok && len(t.([]any)) == 1 {
			settings := t.([]any)[0].(map[string]any)
			destination.NewDataPool.TableSettings = internal.ParseTableSettingsInput(settings)
		}
	}

	input := &pc.CreateMaterializedViewInput{
		UniqueName:  &uniqueName,
		Description: &description,
		Sql:         d.Get("sql").(string),
		Destination: destination,
		BackfillOptions: &pc.BackfillOptionsInput{
			Backfill: &backfill,
		},
	}

	response, err := pc.CreateMaterializedView(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(response.CreateMaterializedView.MaterializedView.Id)

	resourceMaterializedViewRead(ctx, d, meta)

	return nil
}

func resourceMaterializedViewRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(graphql.Client)

	response, err := pc.MaterializedView(ctx, c, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(response.MaterializedView.Id)

	if err := d.Set("unique_name", response.MaterializedView.UniqueName); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("description", response.MaterializedView.Description); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("account", response.MaterializedView.Account.Id); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("environment", response.MaterializedView.Environment.Id); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("sql", response.MaterializedView.Sql); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("destination", response.MaterializedView.Destination.Id); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("source", response.MaterializedView.Source.Id); err != nil {
		return diag.FromErr(err)
	}

	if len(response.MaterializedView.Others) > 0 {
		others := make([]string, 0, len(response.MaterializedView.Others))

		for _, dp := range response.MaterializedView.Others {
			others = append(others, dp.Id)
		}

		if err := d.Set("others", others); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceMaterializedViewUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(graphql.Client)

	input := &pc.ModifyMaterializedViewInput{Id: d.Id()}

	if d.HasChanges("unique_name", "description") {
		uniqueName := d.Get("unique_name").(string)
		description := d.Get("description").(string)

		input.UniqueName = &uniqueName
		input.Description = &description
	}

	if _, err := pc.ModifyMaterializedView(ctx, c, input); err != nil {
		return diag.FromErr(err)
	}

	return resourceMaterializedViewRead(ctx, d, m)
}

func resourceMaterializedViewDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(graphql.Client)

	_, err := pc.DeleteMaterializedView(ctx, c, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
