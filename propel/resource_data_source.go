package propel

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	pc "github.com/propeldata/terraform-provider-propel/propel_client"
)

func resourceDataSource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDataSourceCreate,
		ReadContext:   resourceDataSourceRead,
		UpdateContext: resourceDataSourceUpdate,
		DeleteContext: resourceDataSourceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"unique_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The Data Source name",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The Data Source description",
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"account": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Account that the Data Source belongs to",
			},
			"environment": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Environment that the Data Source belongs to",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of when the Data Source was created",
			},
			"modified_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of when the Data Source was modified",
			},
			"created_by": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The user who created the Data Source",
			},
			"modified_by": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The user who modified the Data Source",
			},
			"snowflake_connection_settings": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{"http_connection_settings"},
				MaxItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account": {
							Type:     schema.TypeString,
							Required: true,
						},
						"database": {
							Type:     schema.TypeString,
							Required: true,
						},
						"warehouse": {
							Type:     schema.TypeString,
							Required: true,
						},
						"schema": {
							Type:     schema.TypeString,
							Required: true,
						},
						"role": {
							Type:     schema.TypeString,
							Required: true,
						},
						"username": {
							Type:     schema.TypeString,
							Required: true,
						},
						"password": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
					},
				},
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
	}
}

func resourceDataSourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// TODO(mroberts): The Propel GraphQL API should eventually return this uppercase.
	dataSourceType := d.Get("type").(string)
	switch strings.ToUpper(dataSourceType) {
	case "SNOWFLAKE":
		return resourceSnowflakeDataSourceCreate(ctx, d, meta)
	default:
		return diag.Errorf("Unsupported Data Source type \"%v\"", dataSourceType)
	}
}

func resourceSnowflakeDataSourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(graphql.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	connectionSettings := d.Get("snowflake_connection_settings").([]interface{})[0].(map[string]interface{})

	input := pc.CreateSnowflakeDataSourceInput{
		UniqueName:  d.Get("unique_name").(string),
		Description: d.Get("description").(string),
		ConnectionSettings: pc.SnowflakeConnectionSettingsInput{
			Account:   connectionSettings["account"].(string),
			Database:  connectionSettings["database"].(string),
			Warehouse: connectionSettings["warehouse"].(string),
			Schema:    connectionSettings["schema"].(string),
			Role:      connectionSettings["role"].(string),
			Username:  connectionSettings["username"].(string),
			Password:  connectionSettings["password"].(string),
		},
	}

	response, err := pc.CreateSnowflakeDataSource(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	switch r := response.GetCreateSnowflakeDataSource().(type) {
	case *pc.CreateSnowflakeDataSourceCreateSnowflakeDataSourceDataSourceResponse:
		d.SetId(r.DataSource.Id)

		timeout := d.Timeout(schema.TimeoutCreate)

		err = waitForDataSourceConnected(ctx, c, d.Id(), timeout)
		if err != nil {
			return diag.FromErr(err)
		}

		return resourceDataSourceRead(ctx, d, meta)
	case *pc.CreateSnowflakeDataSourceCreateSnowflakeDataSourceFailureResponse:
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to create Data Source",
		})
	}

	return diags
}

func resourceDataSourceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(graphql.Client)

	response, err := pc.DataSource(ctx, c, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(response.DataSource.Id)
	if err := d.Set("unique_name", response.DataSource.GetUniqueName()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("description", response.DataSource.GetDescription()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("created_at", response.DataSource.GetCreatedAt().String()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("created_by", response.DataSource.GetCreatedBy()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("modified_at", response.DataSource.GetModifiedAt().String()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("modified_by", response.DataSource.GetModifiedBy()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("environment", response.DataSource.GetEnvironment().Id); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("account", response.DataSource.GetAccount().Id); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("type", response.DataSource.GetType()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("status", response.DataSource.GetStatus()); err != nil {
		return diag.FromErr(err)
	}

	// TODO(mroberts): The Propel GraphQL API should eventually return this uppercase.
	dataSourceType := string(response.DataSource.Type)
	switch strings.ToUpper(dataSourceType) {
	case "SNOWFLAKE":
		return handleSnowflakeConnectionSettings(response, d)
	default:
		return diag.Errorf("Unsupported Data Source type \"%v\"", dataSourceType)
	}
}

func handleSnowflakeConnectionSettings(response *pc.DataSourceResponse, d *schema.ResourceData) diag.Diagnostics {
	cs := d.Get("snowflake_connection_settings").([]interface{})[0].(map[string]interface{})

	settings := map[string]interface{}{
		"password": cs["password"],
	}

	switch s := response.DataSource.GetConnectionSettings().(type) {
	case *pc.DataSourceDataConnectionSettingsSnowflakeConnectionSettings:
		settings["account"] = s.GetAccount()
		settings["database"] = s.GetDatabase()
		settings["warehouse"] = s.GetWarehouse()
		settings["schema"] = s.GetSchema()
		settings["role"] = s.GetRole()
		settings["username"] = s.GetUsername()
	default:
		return diag.Errorf("Missing SnowflakeConnectionSettings")
	}

	if err := d.Set("snowflake_connection_settings", []map[string]interface{}{settings}); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDataSourceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(graphql.Client)

	if d.HasChanges("unique_name", "description") {
		modifyDataSource := pc.ModifySnowflakeDataSourceInput{
			IdOrUniqueName: pc.IdOrUniqueName{
				Id: d.Id(),
			},
			UniqueName:  d.Get("unique_name").(string),
			Description: d.Get("description").(string),
		}

		_, err := pc.ModifySnowflakeDataSource(ctx, c, modifyDataSource)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceDataSourceRead(ctx, d, m)
}

func resourceDataSourceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(graphql.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	_, err := pc.DeleteDataSource(ctx, c, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}

func waitForDataSourceConnected(ctx context.Context, client graphql.Client, id string, timeout time.Duration) error {
	createStateConf := &resource.StateChangeConf{
		Pending: []string{
			string(pc.DataSourceStatusCreated),
			string(pc.DataSourceStatusConnecting),
		},
		Target: []string{
			string(pc.DataSourceStatusConnected),
		},
		Refresh: func() (interface{}, string, error) {
			resp, err := pc.DataSource(ctx, client, id)
			if err != nil {
				return nil, "", fmt.Errorf("error trying to read Data Source status: %s", err)
			}

			return resp, string(resp.DataSource.Status), nil
		},
		Timeout:                   timeout - time.Minute,
		Delay:                     10 * time.Second,
		MinTimeout:                5 * time.Second,
		ContinuousTargetOccurence: 3,
	}

	_, err := createStateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf("error waiting for Data Source to be CONNECTED: %s", err)
	}

	return nil
}
