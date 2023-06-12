package propel

import (
	"context"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	pc "github.com/propeldata/terraform-provider-propel/propel_client"
)

func resourceMetric() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMetricCreate,
		ReadContext:   resourceMetricRead,
		UpdateContext: resourceMetricUpdate,
		DeleteContext: resourceMetricDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Provides a Propel Metric resource. This can be used to create and manage Propel Metrics.",
		Schema: map[string]*schema.Schema{
			"unique_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The Metric's name.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The Metric's description.",
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"SUM",
					"COUNT",
					"COUNT_DISTINCT",
					"AVERAGE",
					"MIN",
					"MAX",
				}, false),
				Description: "The Metric type. The different Metric types determine how the values are calculated.",
			},
			"measure": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
				Description: "The Dimension to be summed, taken the minimum of, taken the maximum of, averaged, etc. Only valid for SUM, MIN, MAX and AVERAGE Metrics.",
			},
			"data_pool": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Data Pool that powers this Metric.",
			},
			"filter": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Metric Filters allow defining a Metric with a subset of records from the given Data Pool. If no Metric Filters are present, all records will be included. To filter at query time, add Dimensions and use the `filters` property on the `timeSeriesInput`, `counterInput`, or `leaderboardInput` objects. There is no need to add `filters` to be able to filter at query time.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"column": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the column to filter on.",
						},
						"operator": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The operation to perform when comparing the column and filter values.",
							ValidateFunc: validation.StringInSlice([]string{
								"EQUALS",
								"NOT_EQUALS",
								"GREATER_THAN",
								"GREATER_THAN_OR_EQUAL_TO",
								"LESS_THAN",
								"LESS_THAN_OR_EQUAL_TO",
							}, false),
						},
						"value": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The value to compare the column to.",
						},
					},
				},
			},
			"dimensions": {
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Description: "The Metric's Dimensions. These Dimensions are available to Query Filters.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"dimension": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "The Dimension where the count distinct operation is going to be performed. Only valid for COUNT_DISTINCT Metrics.",
			},
		},
	}
}

func resourceMetricCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(graphql.Client)

	var diags diag.Diagnostics

	filters := make([]*pc.FilterInput, 0)
	if def, ok := d.Get("filter").([]interface{}); ok && len(def) > 0 {
		filters = expandMetricFilters(def)
	}

	dimensions := make([]*pc.DimensionInput, 0)
	if def, ok := d.GetOk("dimensions"); ok {
		dimensions = expandMetricDimensions(def.(*schema.Set).List())
	}

	dataPool := d.Get("data_pool").(string)
	uniqueName := d.Get("unique_name").(string)
	description := d.Get("description").(string)

	switch d.Get("type") {
	case "SUM":
		input := &pc.CreateSumMetricInput{
			DataPool:    dataPool,
			UniqueName:  &uniqueName,
			Description: &description,
			Filters:     filters,
			Dimensions:  dimensions,
			Measure: &pc.DimensionInput{
				ColumnName: d.Get("measure").(string),
			},
		}

		response, err := pc.CreateSumMetric(ctx, c, input)
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(response.GetCreateSumMetric().Metric.Id)
	case "COUNT":
		input := &pc.CreateCountMetricInput{
			DataPool:    dataPool,
			UniqueName:  &uniqueName,
			Description: &description,
			Filters:     filters,
			Dimensions:  dimensions,
		}

		response, err := pc.CreateCountMetric(ctx, c, input)
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(response.GetCreateCountMetric().Metric.Id)
	case "COUNT_DISTINCT":
		input := &pc.CreateCountDistinctMetricInput{
			DataPool:    dataPool,
			UniqueName:  &uniqueName,
			Description: &description,
			Filters:     filters,
			Dimensions:  dimensions,
			Dimension: &pc.DimensionInput{
				ColumnName: d.Get("dimension").(string),
			},
		}

		response, err := pc.CreateCountDistinctMetric(ctx, c, input)
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(response.GetCreateCountDistinctMetric().Metric.Id)
	case "AVERAGE":
		input := &pc.CreateAverageMetricInput{
			DataPool:    dataPool,
			UniqueName:  &uniqueName,
			Description: &description,
			Filters:     filters,
			Dimensions:  dimensions,
			Measure: &pc.DimensionInput{
				ColumnName: d.Get("measure").(string),
			},
		}

		response, err := pc.CreateAverageMetric(ctx, c, input)
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(response.GetCreateAverageMetric().Metric.Id)
	case "MAX":
		input := &pc.CreateMaxMetricInput{
			DataPool:    dataPool,
			UniqueName:  &uniqueName,
			Description: &description,
			Filters:     filters,
			Dimensions:  dimensions,
			Measure: &pc.DimensionInput{
				ColumnName: d.Get("measure").(string),
			},
		}

		response, err := pc.CreateMaxMetric(ctx, c, input)
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(response.GetCreateMaxMetric().Metric.Id)
	case "MIN":
		input := &pc.CreateMinMetricInput{
			DataPool:    dataPool,
			UniqueName:  &uniqueName,
			Description: &description,
			Filters:     filters,
			Dimensions:  dimensions,
			Measure: &pc.DimensionInput{
				ColumnName: d.Get("measure").(string),
			},
		}

		response, err := pc.CreateMinMetric(ctx, c, input)
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(response.GetCreateMinMetric().Metric.Id)
	}

	return diags
}

func resourceMetricRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(graphql.Client)

	var diags diag.Diagnostics

	response, err := pc.Metric(ctx, c, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(response.Metric.Id)
	if err := d.Set("unique_name", response.Metric.UniqueName); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("description", response.Metric.Description); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("data_pool", response.Metric.DataPool.Id); err != nil {
		return diag.FromErr(err)
	}

	dimensions := make([]string, 0)
	for _, dimension := range response.Metric.Dimensions {
		dimensions = append(dimensions, dimension.ColumnName)
	}

	if err := d.Set("dimensions", dimensions); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("type", response.Metric.Type); err != nil {
		return diag.FromErr(err)
	}

	filters := make([]map[string]interface{}, 0)

	switch s := response.Metric.Settings.(type) {
	case *pc.MetricDataSettingsCountMetricSettings:
		for _, f := range s.Filters {
			filter := map[string]interface{}{
				"column":   f.Column,
				"operator": f.Operator,
				"value":    f.Value,
			}

			filters = append(filters, filter)
		}
	case *pc.MetricDataSettingsSumMetricSettings:
		if err := d.Set("measure", s.Measure.ColumnName); err != nil {
			return diag.FromErr(err)
		}

		for _, f := range s.Filters {
			filter := map[string]interface{}{
				"column":   f.Column,
				"operator": f.Operator,
				"value":    f.Value,
			}

			filters = append(filters, filter)
		}
	case *pc.MetricDataSettingsCountDistinctMetricSettings:
		if err := d.Set("dimension", s.Dimension.ColumnName); err != nil {
			return diag.FromErr(err)
		}

		for _, f := range s.Filters {
			filter := map[string]interface{}{
				"column":   f.Column,
				"operator": f.Operator,
				"value":    f.Value,
			}

			filters = append(filters, filter)
		}
	case *pc.MetricDataSettingsAverageMetricSettings:
		if err := d.Set("measure", s.Measure.ColumnName); err != nil {
			return diag.FromErr(err)
		}

		for _, f := range s.Filters {
			filter := map[string]interface{}{
				"column":   f.Column,
				"operator": f.Operator,
				"value":    f.Value,
			}

			filters = append(filters, filter)
		}
	case *pc.MetricDataSettingsMinMetricSettings:
		if err := d.Set("measure", s.Measure.ColumnName); err != nil {
			return diag.FromErr(err)
		}

		for _, f := range s.Filters {
			filter := map[string]interface{}{
				"column":   f.Column,
				"operator": f.Operator,
				"value":    f.Value,
			}

			filters = append(filters, filter)
		}
	case *pc.MetricDataSettingsMaxMetricSettings:
		if err := d.Set("measure", s.Measure.ColumnName); err != nil {
			return diag.FromErr(err)
		}

		for _, f := range s.Filters {
			filter := map[string]interface{}{
				"column":   f.Column,
				"operator": f.Operator,
				"value":    f.Value,
			}

			filters = append(filters, filter)
		}
	}

	if err := d.Set("filter", filters); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceMetricUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(graphql.Client)

	if d.HasChanges("unique_name", "description", "dimensions", "filter") {
		uniqueName := d.Get("unique_name").(string)
		description := d.Get("description").(string)

		filters := make([]*pc.FilterInput, 0)
		if def, ok := d.Get("filter").([]interface{}); ok && len(def) > 0 {
			filters = expandMetricFilters(def)
		}

		dimensions := make([]*pc.DimensionInput, 0)
		if def, ok := d.GetOk("dimensions"); ok {
			dimensions = expandMetricDimensions(def.(*schema.Set).List())
		}

		modifyMetric := &pc.ModifyMetricInput{
			Metric:      d.Id(),
			UniqueName:  &uniqueName,
			Description: &description,
			Filters:     filters,
			Dimensions:  dimensions,
		}

		_, err := pc.ModifyMetric(ctx, c, modifyMetric)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceMetricRead(ctx, d, m)
}

func resourceMetricDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(graphql.Client)

	_, err := pc.DeleteMetric(ctx, c, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

func expandMetricFilters(def []interface{}) []*pc.FilterInput {
	filters := make([]*pc.FilterInput, 0, len(def))

	for _, rawFilter := range def {
		filter := rawFilter.(map[string]interface{})

		var operator pc.FilterOperator

		switch filter["operator"].(string) {
		case "EQUALS":
			operator = pc.FilterOperatorEquals
		case "NOT_EQUALS":
			operator = pc.FilterOperatorNotEquals
		case "GREATER_THAN":
			operator = pc.FilterOperatorGreaterThan
		case "GREATER_THAN_OR_EQUAL_TO":
			operator = pc.FilterOperatorGreaterThanOrEqualTo
		case "LESS_THAN":
			operator = pc.FilterOperatorLessThan
		case "LESS_THAN_OR_EQUAL_TO":
			operator = pc.FilterOperatorLessThanOrEqualTo
		}

		f := &pc.FilterInput{
			Column:   filter["column"].(string),
			Operator: operator,
			Value:    filter["value"].(string),
		}

		filters = append(filters, f)
	}

	return filters
}

func expandMetricDimensions(def []interface{}) []*pc.DimensionInput {
	dimensions := make([]*pc.DimensionInput, 0, len(def))

	for _, rawDimension := range def {
		dimension := rawDimension.(string)

		d := &pc.DimensionInput{
			ColumnName: dimension,
		}
		dimensions = append(dimensions, d)
	}

	return dimensions
}
