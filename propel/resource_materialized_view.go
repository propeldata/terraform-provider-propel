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
				Description:   "If specified, the Materialized View will target an existing Data Pool. Ensure the Data Pool's schema is compatible with your Materialized View's SQL statement.",
				ConflictsWith: []string{"new_data_pool"},
				MaxItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The ID of the Data Pool",
						},
					},
				},
			},
			"new_data_pool": {
				Type:          schema.TypeList,
				Optional:      true,
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
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{},
													},
												},
												"columns": {
													Type:        schema.TypeSet,
													Optional:    true,
													Description: "The columns argument for the SummingMergeTree table engine.",
													Elem:        &schema.Schema{Type: schema.TypeString},
												},
												"merge_tree": {
													Type:          schema.TypeList,
													Optional:      true,
													Description:   "",
													ConflictsWith: []string{"replacing_merge_tree", "summing_merge_tree", "aggregating_merge_tree"},
													MaxItems:      1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{},
													},
												},
												"replacing_merge_tree": {
													Type:          schema.TypeList,
													Optional:      true,
													Description:   "",
													ConflictsWith: []string{"merge_tree", "summing_merge_tree", "aggregating_merge_tree"},
													MaxItems:      1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"ver": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "The column with the version number",
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{},
																},
															},
														},
													},
												},
												"summing_merge_tree": {
													Type:          schema.TypeList,
													Optional:      true,
													Description:   "",
													ConflictsWith: []string{"merge_tree", "replacing_merge_tree", "aggregating_merge_tree"},
													MaxItems:      1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"columns": {
																Type:        schema.TypeSet,
																Optional:    true,
																Description: "The columns argument for the SummingMergeTree table engine",
																Elem:        &schema.Schema{Type: schema.TypeString},
															},
														},
													},
												},
												"aggregating_merge_tree": {
													Type:          schema.TypeList,
													Optional:      true,
													Description:   "",
													ConflictsWith: []string{"replacing_merge_tree", "summing_merge_tree", ""},
													MaxItems:      1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{},
													},
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
				Description: "Whether historical data should be backfilled or not",
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
		},
	}
}

func resourceMaterializedViewCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(graphql.Client)

	var diags diag.Diagnostics

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

	return diags
}
