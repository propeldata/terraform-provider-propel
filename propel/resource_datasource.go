package propel

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	pc "github.com/propeldata/terraform-provider/graphql_client"
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
				Description: "The DataSource name",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The DataSource description",
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"account": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"environment": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Environment where belong the DataSource",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of when the DataSource was created",
			},
			"modified_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of when the DataSource was modified",
			},
			"created_by": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The user who created the DataSource",
			},
			"modified_by": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The user who modified the DataSource",
			},
			"connection_settings": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
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
			"checks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"error": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"checked_at": {
							Type:     schema.TypeString,
							Computed: true,
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
	c := meta.(graphql.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	connectionSettings := d.Get("connection_settings").([]interface{})[0].(map[string]interface{})

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

	switch resource := response.GetCreateSnowflakeDataSource().(type) {
	case *pc.CreateSnowflakeDataSourceCreateSnowflakeDataSourceDataSourceResponse:
		d.SetId(resource.DataSource.Id)

		timeout := d.Timeout(schema.TimeoutCreate)

		err = waitForDataSourceConnected(ctx, c, d.Id(), timeout)
		if err != nil {
			return diag.FromErr(err)
		}

		return resourceDataSourceRead(ctx, d, meta)
	case *pc.CreateSnowflakeDataSourceCreateSnowflakeDataSourceFailureResponse:
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

	response, err := pc.DataSource(ctx, c, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(response.DataSource.Id)
	if err := d.Set("unique_name", response.DataSource.UniqueName); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("description", response.DataSource.Description); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("created_at", response.DataSource.CreatedAt.String()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("created_by", response.DataSource.CreatedBy); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("modified_at", response.DataSource.ModifiedAt.String()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("modified_by", response.DataSource.ModifiedBy); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("environment", response.DataSource.Environment.Id); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("account", response.DataSource.Account.Id); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("type", response.DataSource.Type); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("status", response.DataSource.Status); err != nil {
		return diag.FromErr(err)
	}

	// NOTE: Hack to parse connection settings
	settingsJSON, err := json.Marshal(response.DataSource.ConnectionSettings)
	if err != nil {
		return diag.FromErr(err)
	}

	var settingsRaw map[string]string

	err = json.Unmarshal(settingsJSON, &settingsRaw)
	if err != nil {
		return diag.FromErr(err)
	}

	settings := []map[string]interface{}{
		{
			"account":   settingsRaw["account"],
			"database":  settingsRaw["database"],
			"warehouse": settingsRaw["warehouse"],
			"schema":    settingsRaw["schema"],
			"role":      settingsRaw["role"],
			"username":  settingsRaw["username"],
		},
	}

	if err := d.Set("connection_settings", settings); err != nil {
		return diag.FromErr(err)
	}

	checks := make([]map[string]interface{}, 0)
	for _, c := range response.DataSource.Checks {
		check := make(map[string]interface{}, 0)
		check["name"] = c.Name
		check["description"] = c.Description
		check["status"] = c.Status
		check["error"] = c.Error.GetMessage()
		check["checked_at"] = c.CheckedAt.String()

		checks = append(checks, check)
	}

	if err := d.Set("checks", checks); err != nil {
		return diag.FromErr(err)
	}
	return diags
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
				return 0, "", fmt.Errorf("error trying to read DataSource status: %s", err)
			}

			if resp.DataSource.Status == pc.DataSourceStatusBroken {
				return 0, string(resp.DataSource.Status), fmt.Errorf("DataSource in BROKEN status")
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
		return fmt.Errorf("error waiting for DataSource to be CONNECTED: %s", err)
	}

	return nil
}
