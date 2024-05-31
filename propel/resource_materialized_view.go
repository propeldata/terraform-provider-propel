package propel

import (
	"context"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

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
						"unique_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: ".Optionally specify the Data Pool's unique ID. This will influence the Data Pool's engine settings.",
						},
						"access_control_enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Enables or disables access control for the Data Pool. If the Data Pool has access control enabled, Applications must be assigned Data Pool Access Policies in order to query the Data Pool and its Metrics.",
						},
						"table_settings": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Override the Data Pool's table settings. These describe how the Data Pool's table is created in ClickHouse, and a default will be chosen based on the Data Pool's `timestamp` and `uniqueId` values, if any. You can override these defaults in order to specify a custom table engine, custom ORDER BY, etc.",
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"engine": {
										Type:        schema.TypeList,
										Optional:    true,
										Description: "The ClickHouse table engine for the Data Pool's table. This field is optional. A default will be chosen based on the Data Pool's `timestamp` and `uniqueId` values, if specified.",
										MaxItems:    1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"type": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.StringInSlice([]string{
														"MERGE_TREE",
														"REPLACING_MERGE_TREE",
														"SUMMING_MERGE_TREE",
														"AGGREGATING_MERGE_TREE",
													}, true),
												},
												"ver": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "The `ver` parameter to the ReplacingMergeTree table engine.",
												},
												"columns": {
													Type:        schema.TypeSet,
													Optional:    true,
													Description: "The columns argument for the SummingMergeTree table engine.",
													Elem:        &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"partition_by": {
										Type:        schema.TypeSet,
										Optional:    true,
										Description: "The PARTITION BY clause for the Data Pool's table. This field is optional. A default will be chosen based on the Data Pool's `timestamp` and `uniqueId` values, if specified.",
										Elem:        &schema.Schema{Type: schema.TypeString},
									},
									"primary_key": {
										Type:        schema.TypeSet,
										Optional:    true,
										Description: "The PRIMARY KEY clause for the Data Pool's table. This field is optional. A default will be chosen based on the Data Pool's `timestamp` and `uniqueId` values, if specified.",
										Elem:        &schema.Schema{Type: schema.TypeString},
									},
									"order_by": {
										Type:        schema.TypeSet,
										Optional:    true,
										Description: "The ORDER BY clause for the Data Pool's table. This field is optional. A default will be chosen based on the Data Pool's `timestamp` and `uniqueId` values, if specified.",
										Elem:        &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
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

	if _, exists := d.GetOk("new_data_pool.0"); exists {
		destination.NewDataPool = &pc.CreateMaterializedViewDestinationNewDataPoolInput{}
		attrs := d.Get("new_data_pool").([]any)[0].(map[string]any)

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
			destination.NewDataPool.TableSettings = &pc.TableSettingsInput{}
			settings := attrs["table_settings"].([]any)[0].(map[string]any)

			if t, ok := settings["engine"]; ok && len(t.([]any)) == 1 {
				destination.NewDataPool.TableSettings.Engine = &pc.TableEngineInput{}
				engine := settings["engine"].([]any)[0].(map[string]any)
				engineType := pc.TableEngineType(engine["type"].(string))

				switch engine["type"].(string) {
				case "MERGE_TREE":
					destination.NewDataPool.TableSettings.Engine.MergeTree = &pc.MergeTreeTableEngineInput{Type: &engineType}
				case "REPLACING_MERGE_TREE":
					destination.NewDataPool.TableSettings.Engine.ReplacingMergeTree = &pc.ReplacingMergeTreeTableEngineInput{Type: &engineType}

					if v, ok := engine["ver"]; ok && v.(string) != "" {
						ver := engine["ver"].(string)
						destination.NewDataPool.TableSettings.Engine.ReplacingMergeTree.Ver = &ver
					}
				case "SUMMING_MERGE_TREE":
					destination.NewDataPool.TableSettings.Engine.SummingMergeTree = &pc.SummingMergeTreeTableEngineInput{Type: &engineType}

					if v, ok := engine["columns"]; ok && len(v.(*schema.Set).List()) > 0 {
						columns := make([]string, 0)
						for _, col := range engine["columns"].(*schema.Set).List() {
							columns = append(columns, col.(string))
						}

						destination.NewDataPool.TableSettings.Engine.SummingMergeTree.Columns = columns
					}
				case "AGGREGATING_MERGE_TREE":
					destination.NewDataPool.TableSettings.Engine.AggregatingMergeTree = &pc.AggregatingMergeTreeTableEngineInput{Type: &engineType}
				}
			}

			if v, ok := settings["partition_by"]; ok && len(v.(*schema.Set).List()) > 0 {
				partitions := make([]string, 0)
				for _, part := range settings["partition_by"].(*schema.Set).List() {
					partitions = append(partitions, part.(string))
				}

				destination.NewDataPool.TableSettings.PartitionBy = partitions
			}

			if v, ok := settings["primary_key"]; ok && len(v.(*schema.Set).List()) > 0 {
				primaryKeys := make([]string, 0)
				for _, k := range settings["primary_key"].(*schema.Set).List() {
					primaryKeys = append(primaryKeys, k.(string))
				}

				destination.NewDataPool.TableSettings.PrimaryKey = primaryKeys
			}

			if v, ok := settings["order_by"]; ok && len(v.(*schema.Set).List()) > 0 {
				orderBy := make([]string, 0)
				for _, k := range settings["order_by"].(*schema.Set).List() {
					orderBy = append(orderBy, k.(string))
				}

				destination.NewDataPool.TableSettings.OrderBy = orderBy
			}
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
