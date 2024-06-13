package internal

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	pc "github.com/propeldata/terraform-provider-propel/propel_client"
)

func TableSettingsSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Computed:    true,
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
								Type:        schema.TypeString,
								Optional:    true,
								ForceNew:    true,
								Description: "The ClickHouse table engine.",
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
								ForceNew:    true,
								Description: "The `ver` parameter to the ReplacingMergeTree table engine.",
							},
							"columns": {
								Type:        schema.TypeSet,
								Optional:    true,
								ForceNew:    true,
								Description: "The columns argument for the SummingMergeTree table engine.",
								Elem:        &schema.Schema{Type: schema.TypeString},
							},
						},
					},
				},
				"partition_by": {
					Type:        schema.TypeSet,
					Optional:    true,
					ForceNew:    true,
					Description: "The PARTITION BY clause for the Data Pool's table. This field is optional. A default will be chosen based on the Data Pool's `timestamp` and `uniqueId` values, if specified.",
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"primary_key": {
					Type:        schema.TypeSet,
					Optional:    true,
					ForceNew:    true,
					Description: "The PRIMARY KEY clause for the Data Pool's table. This field is optional. A default will be chosen based on the Data Pool's `timestamp` and `uniqueId` values, if specified.",
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"order_by": {
					Type:        schema.TypeSet,
					Optional:    true,
					ForceNew:    true,
					Description: "The ORDER BY clause for the Data Pool's table. This field is optional. A default will be chosen based on the Data Pool's `timestamp` and `uniqueId` values, if specified.",
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
			},
		},
	}
}

func BuildTableSettingsInput(settings map[string]any) (*pc.TableSettingsInput, error) {
	tableSettingsInput := &pc.TableSettingsInput{}

	if t, ok := settings["engine"]; ok && len(t.([]any)) == 1 {
		tableSettingsInput.Engine = &pc.TableEngineInput{}
		engine := settings["engine"].([]any)[0].(map[string]any)
		engineType := pc.TableEngineType(engine["type"].(string))

		switch engine["type"].(string) {
		case "MERGE_TREE":
			if v, ok := engine["ver"]; ok && v.(string) != "" {
				return nil, fmt.Errorf("%q field should not be set for MERGE_TREE engine", "ver")
			}

			if v, ok := engine["columns"]; ok && len(v.(*schema.Set).List()) > 0 {
				return nil, fmt.Errorf("%q field should not be set for MERGE_TREE engine", "columns")
			}

			tableSettingsInput.Engine.MergeTree = &pc.MergeTreeTableEngineInput{Type: &engineType}
		case "REPLACING_MERGE_TREE":
			if v, ok := engine["columns"]; ok && len(v.(*schema.Set).List()) > 0 {
				return nil, fmt.Errorf("%q field should not be set for REPLACING_MERGE_TREE engine", "columns")
			}

			tableSettingsInput.Engine.ReplacingMergeTree = &pc.ReplacingMergeTreeTableEngineInput{Type: &engineType}

			if v, ok := engine["ver"]; ok && v.(string) != "" {
				ver := engine["ver"].(string)
				tableSettingsInput.Engine.ReplacingMergeTree.Ver = &ver
			}
		case "SUMMING_MERGE_TREE":
			if v, ok := engine["ver"]; ok && v.(string) != "" {
				return nil, fmt.Errorf("%q field should not be set for SUMMING_MERGE_TREE engine", "ver")
			}

			tableSettingsInput.Engine.SummingMergeTree = &pc.SummingMergeTreeTableEngineInput{Type: &engineType}

			if v, ok := engine["columns"]; ok && len(v.(*schema.Set).List()) > 0 {
				columns := make([]string, 0)
				for _, col := range engine["columns"].(*schema.Set).List() {
					columns = append(columns, col.(string))
				}

				tableSettingsInput.Engine.SummingMergeTree.Columns = columns
			}
		case "AGGREGATING_MERGE_TREE":
			if v, ok := engine["ver"]; ok && v.(string) != "" {
				return nil, fmt.Errorf("%q field should not be set for AGGREGATING_MERGE_TREE engine", "ver")
			}

			if v, ok := engine["columns"]; ok && len(v.(*schema.Set).List()) > 0 {
				return nil, fmt.Errorf("%q field should not be set for AGGREGATING_MERGE_TREE engine", "columns")
			}

			tableSettingsInput.Engine.AggregatingMergeTree = &pc.AggregatingMergeTreeTableEngineInput{Type: &engineType}
		}
	}

	if v, ok := settings["partition_by"]; ok && len(v.(*schema.Set).List()) > 0 {
		partitions := make([]string, 0)
		for _, part := range settings["partition_by"].(*schema.Set).List() {
			partitions = append(partitions, part.(string))
		}

		tableSettingsInput.PartitionBy = partitions
	}

	if v, ok := settings["primary_key"]; ok && len(v.(*schema.Set).List()) > 0 {
		primaryKeys := make([]string, 0)
		for _, k := range settings["primary_key"].(*schema.Set).List() {
			primaryKeys = append(primaryKeys, k.(string))
		}

		tableSettingsInput.PrimaryKey = primaryKeys
	}

	if v, ok := settings["order_by"]; ok && len(v.(*schema.Set).List()) > 0 {
		orderBy := make([]string, 0)
		for _, k := range settings["order_by"].(*schema.Set).List() {
			orderBy = append(orderBy, k.(string))
		}

		tableSettingsInput.OrderBy = orderBy
	}

	return tableSettingsInput, nil
}

func ParseTableSettings(settingsData pc.TableSettingsData) map[string]any {
	settings := map[string]any{
		"partition_by": settingsData.GetPartitionBy(),
		"primary_key":  settingsData.GetPrimaryKey(),
		"order_by":     settingsData.GetOrderBy(),
	}

	if settingsData.GetEngine() != nil {
		switch e := (*settingsData.GetEngine()).(type) {
		case *pc.TableSettingsDataEngineMergeTreeTableEngine:
			settings["engine"] = []map[string]any{
				{
					"type": pc.TableEngineTypeMergeTree,
				},
			}
		case *pc.TableSettingsDataEngineReplacingMergeTreeTableEngine:
			settings["engine"] = []map[string]any{
				{
					"type": pc.TableEngineTypeReplacingMergeTree,
					"ver":  e.GetVer(),
				},
			}
		case *pc.TableSettingsDataEngineSummingMergeTreeTableEngine:
			settings["engine"] = []map[string]any{
				{
					"type":    pc.TableEngineTypeSummingMergeTree,
					"columns": e.GetColumns(),
				},
			}
		case *pc.TableSettingsDataEngineAggregatingMergeTreeTableEngine:
			settings["engine"] = []map[string]any{
				{
					"type": pc.TableEngineTypeAggregatingMergeTree,
				},
			}
		}
	}

	return settings
}
