package propel

import (
	"context"
	"time"

	cms "terraform-provider-hashicups/cms_graphql_client"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDataSource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDataSourceCreate,
		ReadContext:   resourceDataSourceRead,
		UpdateContext: resourceDataSourceUpdate,
		DeleteContext: resourceDataSourceDelete,
		Schema: map[string]*schema.Schema{
			"unique_name": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The DataSource name",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The DataSource description",
			},
			"account": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Account",
			},
			"database": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Database",
			},
			"warehouse": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "WareHouse",
			},
			"schema": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Schema",
			},
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Username",
			},
			"password": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "Password",
			},
			"role": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Role",
			},
			"created_at": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of when the DataSource was created",
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
		},
	}
}

func resourceDataSourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(graphql.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	input := cms.CreateSnowflakeDataSourceInput{
		UniqueName:  d.Get("unique_name").(string),
		Description: d.Get("description").(string),
		ConnectionSettings: cms.SnowflakeConnectionSettingsInput{
			Account:   d.Get("account").(string),
			Database:  d.Get("database").(string),
			Warehouse: d.Get("warehouse").(string),
			Schema:    d.Get("schema").(string),
			Username:  d.Get("username").(string),
			Password:  d.Get("password").(string),
			Role:      d.Get("role").(string),
		},
	}

	response, err := cms.CreateSnowflakeDataSource(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	switch resource := response.GetCreateSnowflakeDataSource().(type) {
	case *cms.CreateSnowflakeDataSourceCreateSnowflakeDataSourceDataSourceResponse:
		d.SetId(resource.DataSource.Id)

		resourceDataSourceRead(ctx, d, meta)
	case *cms.CreateSnowflakeDataSourceCreateSnowflakeDataSourceFailureResponse:
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to create DataSource",
		})
	}

	return diags
}

func resourceDataSourceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(graphql.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	response, err := cms.DataSource(ctx, c, d.Get("id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	dataSource := flattenDataSource(response.DataSource)
	if err := d.Set("dataSource", dataSource); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceDataSourceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//c := m.(*graphql.Client)

	return resourceDataSourceRead(ctx, d, m)
}

func resourceDataSourceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(graphql.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	dataSourceId := d.Id()

	_, err := cms.DeleteDataSource(ctx, c, dataSourceId)
	if err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}

func flattenDataSource(dataSource cms.DataSourceDataSource) []interface{} {
	d := make(map[string]interface{})
	d["id"] = dataSource.Id
	d["uniqueName"] = dataSource.UniqueName
	d["description"] = dataSource.Description
	d["account"] = dataSource.Account
	d["createdAt"] = dataSource.CreatedAt
	d["createdBy"] = dataSource.CreatedBy

	return []interface{}{d}
}
