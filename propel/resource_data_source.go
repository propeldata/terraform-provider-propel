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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/propeldata/terraform-provider-propel/propel/internal/utils"
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
		Description:   "Provides a Propel Data Source resource. This can be used to create and manage Propel Data Sources.",
		Schema: map[string]*schema.Schema{
			"unique_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The Data Source's name.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The Data Source's description.",
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"Snowflake",
					"S3",
					"Http",
					"Webhook",
				}, true),
				Description: "The Data Source's type. Depending on this, you will need to specify one of `http_connection_settings`, `s3_connection_settings`, `webhook_connection_settings` or `snowflake_connection_settings`.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Data Source's status.",
			},
			"account": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Account that the Data Source belongs to.",
			},
			"environment": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Environment that the Data Source belongs to",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of when the Data Source was created.",
			},
			"modified_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of when the Data Source was modified.",
			},
			"created_by": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The user who created the Data Source.",
			},
			"modified_by": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The user who modified the Data Source.",
			},
			"snowflake_connection_settings": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{"http_connection_settings", "s3_connection_settings", "webhook_connection_settings"},
				MaxItems:      1,
				Description:   "Snowflake connection settings. Specify these for Snowflake Data Sources.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The Snowflake account. Only include the part before the \"snowflakecomputing.com\" part of your Snowflake URL (make sure you are in classic console, not Snowsight). For AWS-based accounts, this looks like \"znXXXXX.us-east-2.aws\". For Google Cloud-based accounts, this looks like \"ffXXXXX.us-central1.gcp\".",
						},
						"database": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The Snowflake database name.",
						},
						"warehouse": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The Snowflake warehouse name. It should be \"PROPELLING\" if you used the default name in the setup script.",
						},
						"schema": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The Snowflake schema.",
						},
						"role": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The Snowflake role. It should be \"PROPELLER\" if you used the default name in the setup script.",
						},
						"username": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The Snowflake username. It should be \"PROPEL\" if you used the default name in the setup script.",
						},
						"password": {
							Type:        schema.TypeString,
							Required:    true,
							Sensitive:   true,
							Description: "The Snowflake password.",
						},
					},
				},
			},
			"http_connection_settings": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{"snowflake_connection_settings", "s3_connection_settings", "webhook_connection_settings"},
				MaxItems:      1,
				Elem: &schema.Resource{
					Description: "HTTP connection settings. Specify these for HTTP Data Sources.",
					Schema: map[string]*schema.Schema{
						"basic_auth": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "The HTTP Basic authentication settings for uploading new data.\n\nIf this parameter is not provided, anyone with the URL to your tables will be able to upload data. While it's OK to test without HTTP Basic authentication, we recommend enabling it.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"username": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The username for HTTP Basic authentication that must be included in the Authorization header when uploading new data.",
									},
									"password": {
										Type:        schema.TypeString,
										Required:    true,
										Sensitive:   true,
										Description: "The password for HTTP Basic authentication that must be included in the Authorization header when uploading new data.",
									},
								},
							},
						},
					},
				},
			},
			"s3_connection_settings": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{"snowflake_connection_settings", "http_connection_settings", "webhook_connection_settings"},
				MaxItems:      1,
				Elem: &schema.Resource{
					Description: "The connection settings for an S3 Data Source. These include the S3 bucket name, the AWS access key ID, the AWS secret access key, and the tables (along with their paths).",
					Schema: map[string]*schema.Schema{
						"bucket": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the S3 bucket.",
						},
						"aws_access_key_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The AWS access key ID for an IAM user with sufficient access to the S3 bucket.",
						},
						"aws_secret_access_key": {
							Type:        schema.TypeString,
							Required:    true,
							Sensitive:   true,
							Description: "The AWS secret access key for an IAM user with sufficient access to the S3 bucket.",
						},
					},
				},
			},
			"webhook_connection_settings": {
				Type:          schema.TypeList,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"snowflake_connection_settings", "http_connection_settings", "s3_connection_settings"},
				MaxItems:      1,
				Elem: &schema.Resource{
					Description: "Webhook connection settings. Specify these for Webhook Data Sources.",
					Schema: map[string]*schema.Schema{
						"basic_auth": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "The HTTP basic authentication settings for the Webhook Data Source URL. If this parameter is not provided, anyone with the webhook URL will be able to send events. While it's OK to test without HTTP Basic authentication, we recommend enabling it.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"username": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Username for HTTP Basic authentication that must be included in the Authorization header when uploading new data.",
									},
									"password": {
										Type:        schema.TypeString,
										Required:    true,
										Sensitive:   true,
										Description: "Password for HTTP Basic authentication that must be included in the Authorization header when uploading new data.",
									},
								},
							},
						},
						"column": {
							Type:        schema.TypeList,
							Optional:    true,
							ForceNew:    true,
							Description: "The additional column for the Webhook Data Source table.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The column name.",
									},
									"json_property": {
										Type:     schema.TypeString,
										Required: true,
										Description: `The JSON property that the column will be derived from. For example, if you POST a JSON event like this:
														{ "greeting": { "message": "hello, world" } }
													Then you can use the JSON property "greeting.message" to extract "hello, world" to a column.`,
									},
									"type": {
										Type:         schema.TypeString,
										Required:     true,
										Description:  "The column type.",
										ValidateFunc: utils.IsValidColumnType,
									},
									"nullable": {
										Type:        schema.TypeBool,
										Required:    true,
										Description: "Whether the column's type is nullable or not.",
									},
								},
							},
						},
						"tenant": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "The tenant ID column, if configured.",
						},
						"timestamp": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "The primary timestamp column.",
						},
						"unique_id": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "The unique ID column. Propel uses the primary timestamp and a unique ID to compose a primary key for determining whether records should be inserted, deleted, or updated.",
						},
						"webhook_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The Webhook URL for posting JSON events.",
						},
						"data_pool_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The Webhook Data Pool ID.",
						},
					},
				},
			},
			"table": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Description: "Specify an HTTP or S3 Data Source's tables with this. You do not need to use this for Snowflake Data Sources, since Snowflake Data Sources' tables are automatically introspected. You do not need to use this for Webhook Data Sources, since table is automatically created.",
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The table's ID.",
						},
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the table.",
						},
						"path": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The path to the table's files in S3.",
						},
						"column": {
							Type:        schema.TypeList,
							Required:    true,
							ForceNew:    false,
							Description: "Specify a table's columns.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The column name.",
									},
									"type": {
										Type:         schema.TypeString,
										Required:     true,
										Description:  "The column type.",
										ValidateFunc: utils.IsValidColumnType,
									},
									"nullable": {
										Type:        schema.TypeBool,
										Required:    true,
										Description: "Whether the column's type is nullable or not.",
									},
								},
							},
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

func resourceDataSourceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// TODO(mroberts): The Propel GraphQL API should eventually return this uppercase.
	dataSourceType := d.Get("type").(string)
	switch strings.ToUpper(dataSourceType) {
	case "SNOWFLAKE":
		return resourceSnowflakeDataSourceCreate(ctx, d, meta)
	case "HTTP":
		return resourceHttpDataSourceCreate(ctx, d, meta)
	case "S3":
		return resourceS3DataSourceCreate(ctx, d, meta)
	case "WEBHOOK":
		return resourceWebhookDataSourceCreate(ctx, d, meta)
	default:
		return diag.Errorf("Unsupported Data Source type \"%v\"", dataSourceType)
	}
}

func resourceSnowflakeDataSourceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(graphql.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	connectionSettings := d.Get("snowflake_connection_settings").([]any)[0].(map[string]any)

	uniqueName := d.Get("unique_name").(string)
	description := d.Get("description").(string)
	input := &pc.CreateSnowflakeDataSourceInput{
		UniqueName:  &uniqueName,
		Description: &description,
		ConnectionSettings: &pc.SnowflakeConnectionSettingsInput{
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

	switch r := (*response.GetCreateSnowflakeDataSource()).(type) {
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

func resourceHttpDataSourceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(graphql.Client)

	var basicAuth *pc.HttpBasicAuthInput
	if d.Get("http_connection_settings") != nil && len(d.Get("http_connection_settings").([]any)) > 0 {
		cs := d.Get("http_connection_settings").([]any)[0].(map[string]any)

		if def, ok := cs["basic_auth"]; ok {
			basicAuth = expandBasicAuth(def.([]any))
		}
	}

	tables := make([]*pc.HttpDataSourceTableInput, 0)
	if def, ok := d.Get("table").([]any); ok && len(def) > 0 {
		tables = expandHttpTables(def)
	}

	uniqueName := d.Get("unique_name").(string)
	description := d.Get("description").(string)
	input := &pc.CreateHttpDataSourceInput{
		UniqueName:  &uniqueName,
		Description: &description,
		ConnectionSettings: &pc.HttpConnectionSettingsInput{
			BasicAuth: basicAuth,
			Tables:    tables,
		},
	}

	response, err := pc.CreateHttpDataSource(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	r := response.CreateHttpDataSource
	d.SetId(r.DataSource.Id)

	timeout := d.Timeout(schema.TimeoutCreate)

	if err = waitForDataSourceConnected(ctx, c, d.Id(), timeout); err != nil {
		return diag.FromErr(err)
	}

	return resourceDataSourceRead(ctx, d, meta)
}

func resourceS3DataSourceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(graphql.Client)

	tables := make([]*pc.S3DataSourceTableInput, 0)
	if def, ok := d.Get("table").([]any); ok && len(def) > 0 {
		tables = expandS3Tables(def)
	}

	connectionSettings := d.Get("s3_connection_settings").([]any)[0].(map[string]any)

	uniqueName := d.Get("unique_name").(string)
	description := d.Get("description").(string)
	input := &pc.CreateS3DataSourceInput{
		UniqueName:  &uniqueName,
		Description: &description,
		ConnectionSettings: &pc.S3ConnectionSettingsInput{
			Bucket:             connectionSettings["bucket"].(string),
			AwsAccessKeyId:     connectionSettings["aws_access_key_id"].(string),
			AwsSecretAccessKey: connectionSettings["aws_secret_access_key"].(string),
			Tables:             tables,
		},
	}

	response, err := pc.CreateS3DataSource(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	r := response.CreateS3DataSource
	d.SetId(r.DataSource.Id)

	timeout := d.Timeout(schema.TimeoutCreate)

	err = waitForDataSourceConnected(ctx, c, d.Id(), timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceDataSourceRead(ctx, d, meta)
}

func resourceWebhookDataSourceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(graphql.Client)

	connectionSettings := &pc.WebhookConnectionSettingsInput{}
	if d.Get("webhook_connection_settings") != nil && len(d.Get("webhook_connection_settings").([]any)) > 0 {
		cs := d.Get("webhook_connection_settings").([]any)[0].(map[string]any)

		if def, ok := cs["basic_auth"]; ok {
			connectionSettings.BasicAuth = expandBasicAuth(def.([]any))
		}

		columns := make([]*pc.WebhookDataSourceColumnInput, 0)
		if def, ok := cs["column"].([]any); ok && len(def) > 0 {
			columns = expandWebhookColumns(def)
		}

		connectionSettings.Timestamp = cs["timestamp"].(string)
		connectionSettings.Columns = columns

		if t, ok := cs["tenant"]; ok {
			tenant := t.(string)
			connectionSettings.Tenant = &tenant
		}

		u, ok := cs["unique_id"]
		if ok && u.(string) != "" {
			uniqueID := u.(string)
			connectionSettings.UniqueId = &uniqueID
		}
	}

	uniqueName := d.Get("unique_name").(string)
	description := d.Get("description").(string)

	if connectionSettings.UniqueId != nil {
		fmt.Println("UniqueID>>", *connectionSettings.UniqueId, "<<end")
	}

	fmt.Println(*connectionSettings)

	input := &pc.CreateWebhookDataSourceInput{
		UniqueName:         &uniqueName,
		Description:        &description,
		ConnectionSettings: connectionSettings,
	}

	response, err := pc.CreateWebhookDataSource(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	r := response.CreateWebhookDataSource
	d.SetId(r.DataSource.Id)

	timeout := d.Timeout(schema.TimeoutCreate)

	if err = waitForDataSourceConnected(ctx, c, d.Id(), timeout); err != nil {
		return diag.FromErr(err)
	}

	return resourceDataSourceRead(ctx, d, meta)
}

func resourceDataSourceRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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

	if err := d.Set("status", response.DataSource.GetStatus()); err != nil {
		return diag.FromErr(err)
	}

	// TODO(mroberts): The Propel GraphQL API should eventually return this uppercase.
	dataSourceType := string(response.DataSource.Type)
	switch strings.ToUpper(dataSourceType) {
	case "SNOWFLAKE":
		return handleSnowflakeConnectionSettings(response, d)
	case "HTTP":
		if diags := handleHttpTables(response, d); diags != nil {
			return diags
		}
		return handleHttpConnectionSettings(response, d)
	case "S3":
		if diags := handleS3Tables(response, d); diags != nil {
			return diags
		}
		return handleS3ConnectionSettings(response, d)
	case "WEBHOOK":
		return handleWebhookConnectionSettings(response, d)
	default:
		return diag.Errorf("Unsupported Data Source type \"%v\"", dataSourceType)
	}
}

func handleSnowflakeConnectionSettings(response *pc.DataSourceResponse, d *schema.ResourceData) diag.Diagnostics {
	cs := d.Get("snowflake_connection_settings").([]any)[0].(map[string]any)

	settings := map[string]any{
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

	if err := d.Set("snowflake_connection_settings", []map[string]any{settings}); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func handleHttpTables(response *pc.DataSourceResponse, d *schema.ResourceData) diag.Diagnostics {
	if response.DataSource.GetConnectionSettings().GetTypename() == nil {
		return nil
	}

	switch s := response.DataSource.GetConnectionSettings().(type) {
	case *pc.DataSourceDataConnectionSettingsHttpConnectionSettings:
		tables := make([]any, 0, len(s.Tables))

		for _, table := range s.Tables {
			columns := make([]any, 0, len(table.Columns))
			for _, column := range table.Columns {
				columns = append(columns, map[string]any{
					"name":     column.Name,
					"type":     column.Type,
					"nullable": column.Nullable,
				})
			}
			tables = append(tables, map[string]any{
				"id":     table.Id,
				"name":   table.Name,
				"column": columns,
			})
		}

		if err := d.Set("table", (any)(tables)); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func handleS3Tables(response *pc.DataSourceResponse, d *schema.ResourceData) diag.Diagnostics {
	if response.DataSource.GetConnectionSettings().GetTypename() == nil {
		return nil
	}

	switch s := response.DataSource.GetConnectionSettings().(type) {
	case *pc.DataSourceDataConnectionSettingsS3ConnectionSettings:
		tables := make([]any, 0, len(s.Tables))

		for _, table := range s.Tables {
			columns := make([]any, 0, len(table.Columns))
			for _, column := range table.Columns {
				columns = append(columns, map[string]any{
					"name":     column.Name,
					"type":     column.Type,
					"nullable": column.Nullable,
				})
			}
			tables = append(tables, map[string]any{
				"id":     table.Id,
				"name":   table.Name,
				"path":   table.Path,
				"column": columns,
			})
		}

		if err := d.Set("table", (any)(tables)); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func handleHttpConnectionSettings(response *pc.DataSourceResponse, d *schema.ResourceData) diag.Diagnostics {
	if d.Get("http_connection_settings") == nil || len(d.Get("http_connection_settings").([]any)) == 0 {
		return nil
	}

	switch s := response.DataSource.GetConnectionSettings().(type) {
	case *pc.DataSourceDataConnectionSettingsHttpConnectionSettings:
		settings := map[string]any{
			"basic_auth": nil,
		}

		if s.BasicAuth != nil {
			settings["basic_auth"] = []map[string]any{
				{
					"username": s.BasicAuth.GetUsername(),
					"password": s.BasicAuth.GetPassword(),
				},
			}
		}

		if err := d.Set("http_connection_settings", []map[string]any{settings}); err != nil {
			return diag.FromErr(err)
		}
	default:
		return diag.Errorf("Missing HttpConnectionSettings")
	}

	return nil
}

func handleS3ConnectionSettings(response *pc.DataSourceResponse, d *schema.ResourceData) diag.Diagnostics {
	cs := d.Get("s3_connection_settings").([]any)[0].(map[string]any)

	settings := map[string]any{
		"aws_secret_access_key": cs["aws_secret_access_key"],
	}

	switch s := response.DataSource.GetConnectionSettings().(type) {
	case *pc.DataSourceDataConnectionSettingsS3ConnectionSettings:
		settings["bucket"] = s.GetBucket()
		settings["aws_access_key_id"] = s.GetAwsAccessKeyId()

		if err := d.Set("s3_connection_settings", []map[string]any{settings}); err != nil {
			return diag.FromErr(err)
		}
	default:
		return diag.Errorf("Missing S3ConnectionSettings")
	}

	return nil
}

func handleWebhookConnectionSettings(response *pc.DataSourceResponse, d *schema.ResourceData) diag.Diagnostics {
	if d.Get("webhook_connection_settings") == nil || len(d.Get("webhook_connection_settings").([]any)) == 0 {
		return nil
	}

	switch s := response.DataSource.GetConnectionSettings().(type) {
	case *pc.DataSourceDataConnectionSettingsWebhookConnectionSettings:
		settings := map[string]any{
			"timestamp":   s.GetTimestamp(),
			"tenant":      s.GetTenant(),
			"unique_id":   s.GetUniqueId(),
			"webhook_url": s.GetWebhookUrl(),
		}

		if s.BasicAuth != nil {
			settings["basic_auth"] = []map[string]any{
				{
					"username": s.BasicAuth.GetUsername(),
					"password": s.BasicAuth.GetPassword(),
				},
			}
		}

		cols := make([]any, len(s.Columns))

		for i, column := range s.Columns {
			cols[i] = map[string]any{
				"name":          column.Name,
				"type":          column.Type,
				"nullable":      column.Nullable,
				"json_property": column.JsonProperty,
			}
		}

		settings["column"] = cols

		if len(response.DataSource.DataPools.GetNodes()) == 1 {
			settings["data_pool_id"] = response.DataSource.DataPools.Nodes[0].Id
		}

		if err := d.Set("webhook_connection_settings", []map[string]any{settings}); err != nil {
			return diag.FromErr(err)
		}
	default:
		return diag.Errorf("Missing WebhookConnectionSettings")
	}

	return nil
}

func resourceSnowflakeDataSourceUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(graphql.Client)

	if d.HasChanges("unique_name", "description", "snowflake_connection_settings") {
		id := d.Id()
		uniqueName := d.Get("unique_name").(string)
		description := d.Get("description").(string)

		connectionSettings := d.Get("snowflake_connection_settings").([]any)[0].(map[string]any)
		connectionSettingsInput := &pc.PartialSnowflakeConnectionSettingsInput{}

		if def, ok := connectionSettings["account"]; ok {
			account := def.(string)
			connectionSettingsInput.Account = &account
		}

		if def, ok := connectionSettings["database"]; ok {
			database := def.(string)
			connectionSettingsInput.Database = &database
		}

		if def, ok := connectionSettings["warehouse"]; ok {
			warehouse := def.(string)
			connectionSettingsInput.Warehouse = &warehouse
		}

		if def, ok := connectionSettings["schema"]; ok {
			schemaF := def.(string)
			connectionSettingsInput.Schema = &schemaF
		}

		if def, ok := connectionSettings["role"]; ok {
			role := def.(string)
			connectionSettingsInput.Role = &role
		}

		if def, ok := connectionSettings["username"]; ok {
			username := def.(string)
			connectionSettingsInput.Username = &username
		}

		if def, ok := connectionSettings["password"]; ok {
			password := def.(string)
			connectionSettingsInput.Password = &password
		}

		modifyDataSource := &pc.ModifySnowflakeDataSourceInput{
			IdOrUniqueName: &pc.IdOrUniqueName{
				Id: &id,
			},
			UniqueName:         &uniqueName,
			Description:        &description,
			ConnectionSettings: connectionSettingsInput,
		}

		_, err := pc.ModifySnowflakeDataSource(ctx, c, modifyDataSource)
		if err != nil {
			return diag.FromErr(err)
		}

		timeout := d.Timeout(schema.TimeoutCreate)

		if err := waitForDataSourceConnected(ctx, c, d.Id(), timeout); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceDataSourceRead(ctx, d, m)
}

func resourceHttpDataSourceUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(graphql.Client)

	if d.HasChanges("unique_name", "description", "table", "http_connection_settings") {
		id := d.Id()
		uniqueName := d.Get("unique_name").(string)
		description := d.Get("description").(string)

		var basicAuth *pc.HttpBasicAuthInput
		if d.Get("http_connection_settings") != nil && len(d.Get("http_connection_settings").([]any)) > 0 {
			cs := d.Get("http_connection_settings").([]any)[0].(map[string]any)

			if def, ok := cs["basic_auth"]; ok {
				basicAuth = expandBasicAuth(def.([]any))
			}
		}

		tables := make([]*pc.HttpDataSourceTableInput, 0)
		if _, ok := d.GetOk("table"); ok {
			tables = expandHttpTables(d.Get("table").([]any))
		}

		input := &pc.ModifyHttpDataSourceInput{
			IdOrUniqueName: &pc.IdOrUniqueName{
				Id: &id,
			},
			UniqueName:  &uniqueName,
			Description: &description,
			ConnectionSettings: &pc.PartialHttpConnectionSettingsInput{
				BasicAuth: basicAuth,
				Tables:    tables,
			},
		}

		if _, err := pc.ModifyHttpDataSource(ctx, c, input); err != nil {
			return diag.FromErr(err)
		}

		timeout := d.Timeout(schema.TimeoutCreate)

		if err := waitForDataSourceConnected(ctx, c, d.Id(), timeout); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceDataSourceRead(ctx, d, m)
}

func resourceS3DataSourceUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(graphql.Client)

	if d.HasChanges("unique_name", "description", "table", "s3_connection_settings") {
		id := d.Id()
		uniqueName := d.Get("unique_name").(string)
		description := d.Get("description").(string)

		tables := make([]*pc.S3DataSourceTableInput, 0)
		if _, ok := d.GetOk("table"); ok {
			tables = expandS3Tables(d.Get("table").([]any))
		}

		csPartialInput := &pc.PartialS3ConnectionSettingsInput{Tables: tables}

		connectionSettings := d.Get("s3_connection_settings").([]any)[0].(map[string]any)

		if def, ok := connectionSettings["bucket"]; ok {
			bucket := def.(string)
			csPartialInput.Bucket = &bucket
		}

		if def, ok := connectionSettings["aws_access_key_id"]; ok {
			accessKeyID := def.(string)
			csPartialInput.AwsAccessKeyId = &accessKeyID
		}

		if def, ok := connectionSettings["aws_access_key_id"]; ok {
			accessKeyID := def.(string)
			csPartialInput.AwsAccessKeyId = &accessKeyID
		}

		if def, ok := connectionSettings["aws_secret_access_key"]; ok {
			secretAccessKey := def.(string)
			csPartialInput.AwsSecretAccessKey = &secretAccessKey
		}

		input := &pc.ModifyS3DataSourceInput{
			IdOrUniqueName: &pc.IdOrUniqueName{
				Id: &id,
			},
			UniqueName:         &uniqueName,
			Description:        &description,
			ConnectionSettings: csPartialInput,
		}

		if _, err := pc.ModifyS3DataSource(ctx, c, input); err != nil {
			return diag.FromErr(err)
		}

		timeout := d.Timeout(schema.TimeoutCreate)

		if err := waitForDataSourceConnected(ctx, c, d.Id(), timeout); err != nil {
			return diag.FromErr(err)
		}
	}
	return resourceDataSourceRead(ctx, d, m)
}

func resourceWebhookDataSourceUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(graphql.Client)

	if d.HasChanges("unique_name", "description", "webhook_connection_settings") {
		id := d.Id()
		uniqueName := d.Get("unique_name").(string)
		description := d.Get("description").(string)

		var basicAuth *pc.HttpBasicAuthInput
		if d.Get("webhook_connection_settings") != nil && len(d.Get("webhook_connection_settings").([]any)) > 0 {
			cs := d.Get("webhook_connection_settings").([]any)[0].(map[string]any)

			if def, ok := cs["basic_auth"]; ok {
				basicAuth = expandBasicAuth(def.([]any))
			}
		}

		input := &pc.ModifyWebhookDataSourceInput{
			IdOrUniqueName: &pc.IdOrUniqueName{
				Id: &id,
			},
			UniqueName:  &uniqueName,
			Description: &description,
			ConnectionSettings: &pc.PartialWebhookConnectionSettingsInput{
				BasicAuth: basicAuth,
			},
		}

		if _, err := pc.ModifyWebhookDataSource(ctx, c, input); err != nil {
			return diag.FromErr(err)
		}

		timeout := d.Timeout(schema.TimeoutCreate)

		if err := waitForDataSourceConnected(ctx, c, d.Id(), timeout); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceDataSourceRead(ctx, d, m)
}

func resourceDataSourceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {

	dataSourceType := d.Get("type").(string)
	switch strings.ToUpper(dataSourceType) {
	case "SNOWFLAKE":
		return resourceSnowflakeDataSourceUpdate(ctx, d, meta)
	case "HTTP":
		return resourceHttpDataSourceUpdate(ctx, d, meta)
	case "S3":
		return resourceS3DataSourceUpdate(ctx, d, meta)
	case "WEBHOOK":
		return resourceWebhookDataSourceUpdate(ctx, d, meta)
	default:
		return diag.Errorf("Unsupported Data Source type \"%v\"", dataSourceType)
	}

}

func resourceDataSourceDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(graphql.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Deletes default Data Pool first
	if strings.ToUpper(d.Get("type").(string)) == "WEBHOOK" {
		cs := d.Get("webhook_connection_settings").([]any)[0].(map[string]any)

		dataPoolID := cs["data_pool_id"].(string)

		_, err := pc.DeleteDataPool(ctx, c, dataPoolID)
		if err != nil {
			return diag.FromErr(err)
		}

		timeout := d.Timeout(schema.TimeoutDelete)
		if err := waitForDataPoolDeletion(ctx, c, dataPoolID, timeout); err != nil {
			return diag.FromErr(err)
		}
	}

	_, err := pc.DeleteDataSource(ctx, c, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	timeout := d.Timeout(schema.TimeoutDelete)
	if err = waitForDataSourceDeletion(ctx, c, d.Id(), timeout); err != nil {
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
		Refresh: func() (any, string, error) {
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

func waitForDataSourceDeletion(ctx context.Context, client graphql.Client, id string, timeout time.Duration) error {
	tickerInterval := 10 // 10s
	timeoutSeconds := int(timeout.Seconds())
	n := 0

	ticker := time.NewTicker(time.Duration(tickerInterval) * time.Second)
	for range ticker.C {
		if n*tickerInterval > timeoutSeconds {
			ticker.Stop()
			break
		}

		_, err := pc.DataSource(ctx, client, id)
		if err != nil {
			ticker.Stop()

			if strings.Contains(err.Error(), "not found") {
				return nil
			}

			return fmt.Errorf("error trying to fetch Data Source: %s", err)
		}

		n++
	}
	return nil
}

func expandHttpTables(def []any) []*pc.HttpDataSourceTableInput {
	tables := make([]*pc.HttpDataSourceTableInput, 0, len(def))

	for _, rawTable := range def {
		table := rawTable.(map[string]any)

		columns := expandHttpColumns(table["column"].([]any))

		tables = append(tables, &pc.HttpDataSourceTableInput{
			Name:    table["name"].(string),
			Columns: columns,
		})
	}

	return tables
}

func expandHttpColumns(def []any) []*pc.HttpDataSourceColumnInput {
	columns := make([]*pc.HttpDataSourceColumnInput, len(def))

	for i, rawColumn := range def {
		column := rawColumn.(map[string]any)

		columns[i] = &pc.HttpDataSourceColumnInput{
			Name:     column["name"].(string),
			Type:     pc.ColumnType(column["type"].(string)),
			Nullable: column["nullable"].(bool),
		}
	}

	return columns
}

func expandS3Tables(def []any) []*pc.S3DataSourceTableInput {
	tables := make([]*pc.S3DataSourceTableInput, len(def))

	for i, rawTable := range def {
		table := rawTable.(map[string]any)

		columns := expandS3Columns(table["column"].([]any))

		path := table["path"].(string)
		tables[i] = &pc.S3DataSourceTableInput{
			Name:    table["name"].(string),
			Path:    &path,
			Columns: columns,
		}
	}

	return tables
}

func expandS3Columns(def []any) []*pc.S3DataSourceColumnInput {
	columns := make([]*pc.S3DataSourceColumnInput, len(def))

	for i, rawColumn := range def {
		column := rawColumn.(map[string]any)

		columns[i] = &pc.S3DataSourceColumnInput{
			Name:     column["name"].(string),
			Type:     pc.ColumnType(column["type"].(string)),
			Nullable: column["nullable"].(bool),
		}
	}

	return columns
}

func expandBasicAuth(def []any) *pc.HttpBasicAuthInput {
	basicAuth := def[0].(map[string]any)

	return &pc.HttpBasicAuthInput{
		Username: basicAuth["username"].(string),
		Password: basicAuth["password"].(string),
	}
}

func expandWebhookColumns(def []any) []*pc.WebhookDataSourceColumnInput {
	columns := make([]*pc.WebhookDataSourceColumnInput, len(def))

	for i, rawColumn := range def {
		column := rawColumn.(map[string]any)

		columns[i] = &pc.WebhookDataSourceColumnInput{
			Name:         column["name"].(string),
			Type:         pc.ColumnType(column["type"].(string)),
			Nullable:     column["nullable"].(bool),
			JsonProperty: column["json_property"].(string),
		}
	}

	return columns
}
