package propel

import (
	"context"

	cms "terraform-provider-hashicups/cms_graphql_client"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDataPool() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDataPoolCreate,
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"unique_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The DataPool name",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The DataPool description",
			},
			"data_source": {
				Type:     schema.TypeString,
				Required: true,
			},
			"table": {
				Type:     schema.TypeString,
				Required: true,
			},
			"timestamp": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceDataPoolCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(graphql.Client)

	var diags diag.Diagnostics

	response, err := cms.CreateDataPool(ctx, c, cms.CreateDataPoolInput{
		UniqueName:  "",
		Description: "",
		DataSource: cms.IdOrUniqueName{
			Id: d.Get("data_source").(string),
		},
		Table: d.Get("table").(string),
		Timestamp: cms.DimensionInput{
			ColumnName: d.Get("timestamp").(string),
		},
	})

	if err != nil {
		return diag.FromErr(err)
	}

	switch resource := response.GetCreateDataPool().(type) {
	case *cms.CreateDataPoolCreateDataPoolDataPoolResponse:
		d.SetId(resource.DataPool.Id)

		resourceDataPoolRead(ctx, d, meta)
	case *cms.CreateDataPoolCreateDataPoolFailureResponse:
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to create DataPool",
		})
	}

	return diags
}

func resourceDataPoolRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(graphql.Client)

	var diags diag.Diagnostics

	response, err := cms.DataPool(ctx, c, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(response.DataPool.Id)
	if err := d.Set("unique_name", response.DataPool.UniqueName); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
