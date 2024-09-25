package propel

import (
	"context"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/propeldata/terraform-provider-propel/propel/internal/utils"
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
				ForceNew:    true,
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
				Elem:        &schema.Schema{Type: schema.TypeString},
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
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The operation to perform when comparing the column and filter values.",
							ValidateFunc: utils.IsValidOperator,
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
				Description: `The list of applications to which the Access Policy is assigned.`,
				Elem:        &schema.Schema{Type: schema.TypeString},
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

	columns := make([]string, 0)
	if def, ok := d.GetOk("columns"); ok {
		for _, col := range def.([]any) {
			columns = append(columns, col.(string))
		}
	}

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

	if def, ok := d.GetOk("applications"); ok {
		for _, app := range def.([]any) {
			_, err = pc.AssignDataPoolAccessPolicy(ctx, c, app.(string), response.CreateDataPoolAccessPolicy.DataPoolAccessPolicy.Id)
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

	apps := make([]string, 0)
	if response.DataPoolAccessPolicy.Applications != nil {
		for _, node := range response.DataPoolAccessPolicy.Applications.Nodes {
			apps = append(apps, node.Id)
		}
	}

	if err := d.Set("applications", apps); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDataPoolAccessPolicyUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(graphql.Client)

	var diags diag.Diagnostics

	if d.HasChanges("unique_name", "description", "data_pool", "columns", "row") {
		id := d.Id()
		uniqueName := d.Get("unique_name").(string)
		description := d.Get("description").(string)

		columns := make([]string, 0)
		if def, ok := d.GetOk("columns"); ok {
			for _, col := range def.([]any) {
				columns = append(columns, col.(string))
			}
		}

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
		oldApplications, newApplications := oldItem.([]any), newItem.([]any)
		oldMap, newMap := map[string]bool{}, map[string]bool{}

		for _, oldApp := range oldApplications {
			oldMap[oldApp.(string)] = true
		}

		for _, newApp := range newApplications {
			if !oldMap[newApp.(string)] {
				_, err := pc.AssignDataPoolAccessPolicy(ctx, c, newApp.(string), id)
				if err != nil {
					return diag.FromErr(err)
				}
			}

			newMap[newApp.(string)] = true
		}

		for oldApp := range oldMap {
			if !newMap[oldApp] {
				_, err := pc.UnAssignDataPoolAccessPolicy(ctx, c, id, oldApp)
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

	if def, ok := d.GetOk("applications"); ok {
		applications := def.([]any)
		for _, app := range applications {
			_, err := pc.UnAssignDataPoolAccessPolicy(ctx, c, d.Id(), app.(string))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	_, err := pc.DeleteDataPoolAccessPolicy(ctx, c, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
