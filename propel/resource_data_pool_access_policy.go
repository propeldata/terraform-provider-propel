package propel

import (
	"context"
	"slices"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	pc "github.com/propeldata/terraform-provider-propel/propel_client"
)

func resourceDataPoolAccessPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDataPoolAccessPolicyCreate,
		ReadContext:   resourceDataPoolAccessPolicyRead,
		UpdateContext: resourceDataPoolAccessPolicyUpdate,
		DeleteContext: resourceDataPoolAccessPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Description:   "Provides a Propel Data Pool Access Policy resource. This can be used to create and manage Propel Data Pool Access Policies.",
		Schema: map[string]*schema.Schema{
			"unique_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The Data Pool Access Policy's name.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The Data Pool Access Policy's description.",
			},
			"data_pool": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Data Pool to which this Access Policy belongs.",
			},
			"account": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Account to which the Data Pool Access Policy belongs.",
			},
			"environment": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Environment to which the Data Pool Access Policy belongs.",
			},
			"columns": {
				Type:        schema.TypeList,
				Required:    true,
				Description: `The list of columns that the Access Policy makes available for querying. Set "*" to allow all columns.`,
				Elem:        schema.TypeString,
			},
			"row": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: `Row-level filters that the Access Policy applies before executing queries. Not setting any row filters means all rows can be queried.`,
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
								"IS_NULL",
								"IS_NOT_NULL",
							}, false),
						},
						"value": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The value to compare the column to.",
						},
						"and": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Additional filters to AND with this one. AND takes precedence over OR. It is defined as a JSON string value.",
							ValidateFunc: validation.StringIsJSON,
							StateFunc: func(v any) string {
								nJSON, _ := structure.NormalizeJsonString(v)
								return nJSON
							},
						},
						"or": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Additional filters to OR with this one. AND takes precedence over OR. It is defined as a JSON string value.",
							ValidateFunc: validation.StringIsJSON,
							StateFunc: func(v any) string {
								nJSON, _ := structure.NormalizeJsonString(v)
								return nJSON
							},
						},
					},
				},
			},
			"applications": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: `The list of columns that the Access Policy makes available for querying. Set "*" to allow all columns.`,
				Elem:        schema.TypeString,
			},
		},
	}
}

func resourceDataPoolAccessPolicyCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(graphql.Client)

	var diags diag.Diagnostics

	dataPoolId := d.Get("data_pool").(string)
	uniqueName := d.Get("unique_name").(string)
	description := d.Get("description").(string)
	columns := d.Get("columns").([]string)

	rows := make([]*pc.FilterInput, 0)
	if def, ok := d.Get("row").([]any); ok && len(def) > 0 {
		rows, diags = expandMetricFilters(def)
		if diags != nil {
			return diags
		}
	}

	createPolicyInput := &pc.CreateDataPoolAccessPolicyInput{
		UniqueName:  &uniqueName,
		Description: &description,
		DataPool:    dataPoolId,
		Columns:     columns,
		Rows:        rows,
	}

	response, err := pc.CreateDataPoolAccessPolicy(ctx, c, createPolicyInput)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(response.CreateDataPoolAccessPolicy.DataPoolAccessPolicy.Id)

	if _, exists := d.GetOk("applications"); exists {
		applications := d.Get("applications").([]string)
		for _, app := range applications {
			_, err = pc.AssignDataPoolAccessPolicy(ctx, c, app, response.CreateDataPoolAccessPolicy.DataPoolAccessPolicy.Id)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	resourceDataPoolAccessPolicyRead(ctx, d, meta)

	return diags
}

func resourceDataPoolAccessPolicyRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(graphql.Client)

	var diags diag.Diagnostics

	response, err := pc.DataPoolAccessPolicy(ctx, c, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(response.DataPoolAccessPolicy.Id)

	if err := d.Set("unique_name", response.DataPoolAccessPolicy.UniqueName); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("description", response.DataPoolAccessPolicy.Description); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("environment", response.DataPoolAccessPolicy.Environment.Id); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("account", response.DataPoolAccessPolicy.Account.Id); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("data_pool", response.DataPoolAccessPolicy.DataPool.Id); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("columns", response.DataPoolAccessPolicy.Columns); err != nil {
		return diag.FromErr(err)
	}

	rows, err := parseMetricFilters(response.DataPoolAccessPolicy.Rows)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("row", rows); err != nil {
		return diag.FromErr(err)
	}

	apps := make([]string, 0, len(response.DataPoolAccessPolicy.Applications.Nodes))
	for _, node := range response.DataPoolAccessPolicy.Applications.Nodes {
		apps = append(apps, node.Id)
	}

	if err := d.Set("applications", apps); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceDataPoolAccessPolicyUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(graphql.Client)

	var diags diag.Diagnostics

	if d.HasChanges("unique_name", "description", "data_pool", "columns", "row") {
		id := d.Id()
		uniqueName := d.Get("unique_name").(string)
		description := d.Get("description").(string)
		columns := d.Get("columns").([]string)

		rows := make([]*pc.FilterInput, 0)
		if def, ok := d.Get("row").([]any); ok && len(def) > 0 {
			rows, diags = expandMetricFilters(def)
			if diags != nil {
				return diags
			}
		}

		input := &pc.ModifyDataPoolAccessPolicyInput{
			Id:          id,
			UniqueName:  &uniqueName,
			Description: &description,
			Columns:     columns,
			Rows:        rows,
		}

		_, err := pc.ModifyDataPoolAccessPolicy(ctx, c, input)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChanges("applications") {
		id := d.Id()

		oldItem, newItem := d.GetChange("applications")
		oldApplications, newApplications := oldItem.([]string), newItem.([]string)

		// TODO: maybe make this a bit more efficient
		for _, oldApp := range oldApplications {
			if !slices.Contains(newApplications, oldApp) {
				_, err := pc.UnAssignDataPoolAccessPolicy(ctx, c, id, oldApp)
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}
		for _, newApp := range newApplications {
			if !slices.Contains(oldApplications, newApp) {
				_, err := pc.AssignDataPoolAccessPolicy(ctx, c, id, newApp)
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}
	}

	return resourceDataPoolAccessPolicyRead(ctx, d, m)
}

func resourceDataPoolAccessPolicyDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(graphql.Client)

	applications := d.Get("applications").([]string)
	for _, app := range applications {
		_, err := pc.UnAssignDataPoolAccessPolicy(ctx, c, d.Id(), app)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	_, err := pc.DeleteDataPoolAccessPolicy(ctx, c, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
