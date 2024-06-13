package propel

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/propeldata/terraform-provider-propel/propel/internal"
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
					"Kafka",
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
			"snowflake_connection_settings": internal.SnowflakeDataSourceSchema(),
			"http_connection_settings":      internal.HttpDataSourceSchema(),
			"s3_connection_settings": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{"snowflake_connection_settings", "http_connection_settings", "webhook_connection_settings", "kafka_connection_settings", "clickhouse_connection_settings"},
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
			"webhook_connection_settings": internal.WebhookDataSourceSchema(),
			"kafka_connection_settings": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{"snowflake_connection_settings", "http_connection_settings", "s3_connection_settings", "webhook_connection_settings", "clickhouse_connection_settings"},
				MaxItems:      1,
				Elem: &schema.Resource{
					Description: "The connection settings for a Kafka Data Source.",
					Schema: map[string]*schema.Schema{
						"auth": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The type of authentication to use. Can be SCRAM-SHA-256, SCRAM-SHA-512, PLAIN or NONE.",
							ValidateFunc: validation.StringInSlice([]string{
								"SCRAM-SHA-256",
								"SCRAM-SHA-512",
								"PLAIN",
								"NONE",
							}, true),
						},
						"user": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The user for authenticating against the Kafka servers.",
						},
						"password": {
							Type:        schema.TypeString,
							Required:    true,
							Sensitive:   true,
							Description: "The password for the provided user.",
						},
						"tls": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Whether the the connection to the Kafka servers is encrypted or not.",
						},
						"bootstrap_servers": {
							Type:        schema.TypeSet,
							Required:    true,
							Description: "The bootstrap server(s) to connect to.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"clickhouse_connection_settings": internal.ClickHouseDataSourceSchema(),
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
	var id string
	var err error

	c := meta.(graphql.Client)
	dataSourceType := d.Get("type").(string)

	switch strings.ToUpper(dataSourceType) {
	case "SNOWFLAKE":
		id, err = internal.SnowflakeDataSourceCreate(ctx, d, c)
	case "HTTP":
		id, err = internal.HttpDataSourceCreate(ctx, d, c)
	case "S3":
		return resourceS3DataSourceCreate(ctx, d, meta)
	case "WEBHOOK":
		id, err = internal.WebhookDataSourceCreate(ctx, d, c)
	case "KAFKA":
		return resourceKafkaDataSourceCreate(ctx, d, meta)
	case "CLICKHOUSE":
		id, err = internal.ClickHouseDataSourceCreate(ctx, d, c)
	default:
		err = fmt.Errorf("unsupported Data Source type \"%v\"", dataSourceType)
	}

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)

	timeout := d.Timeout(schema.TimeoutCreate)
	if err := waitForDataSourceConnected(ctx, c, id, timeout); err != nil {
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

func resourceKafkaDataSourceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(graphql.Client)

	uniqueName := d.Get("unique_name").(string)
	description := d.Get("description").(string)
	connectionSettings := d.Get("kafka_connection_settings.0").(map[string]any)

	bootstrapServers := make([]string, 0)
	if v, exists := connectionSettings["bootstrap_servers"]; exists {
		for _, bServer := range v.(*schema.Set).List() {
			bootstrapServers = append(bootstrapServers, bServer.(string))
		}
	}

	input := &pc.CreateKafkaDataSourceInput{
		UniqueName:  &uniqueName,
		Description: &description,
		ConnectionSettings: &pc.KafkaConnectionSettingsInput{
			Auth:             connectionSettings["auth"].(string),
			User:             connectionSettings["user"].(string),
			Password:         connectionSettings["password"].(string),
			BootstrapServers: bootstrapServers,
		},
	}

	if v, exists := connectionSettings["tls"]; exists && v.(bool) {
		tls := v.(bool)
		input.ConnectionSettings.Tls = &tls
	}

	response, err := pc.CreateKafkaDataSource(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	r := response.CreateKafkaDataSource
	d.SetId(r.DataSource.Id)

	timeout := d.Timeout(schema.TimeoutCreate)

	err = waitForDataSourceConnected(ctx, c, d.Id(), timeout)
	if err != nil {
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

	if err := d.Set("type", utils.GetDataSourceType(response.DataSource.GetType())); err != nil {
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

	dataSourceType := string(response.DataSource.Type)
	switch strings.ToUpper(dataSourceType) {
	case "SNOWFLAKE":
		err = internal.HandleSnowflakeConnectionSettings(response, d)
	case "HTTP":
		err = internal.HandleHttpConnectionSettings(response, d)
	case "S3":
		if diags := handleS3Tables(response, d); diags != nil {
			return diags
		}
		return handleS3ConnectionSettings(response, d)
	case "WEBHOOK":
		return handleWebhookConnectionSettings(response, d)
	case "KAFKA":
		return handleKafkaConnectionSettings(response, d)
	case "CLICKHOUSE":
		err = internal.HandleClickHouseConnectionSettings(response, d)
	default:
		err = fmt.Errorf("unsupported Data Source type \"%v\"", dataSourceType)
	}

	if err != nil {
		return diag.FromErr(err)
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

func handleS3ConnectionSettings(response *pc.DataSourceResponse, d *schema.ResourceData) diag.Diagnostics {
	settings := make(map[string]any)

	csList := d.Get("s3_connection_settings").([]any)
	if len(csList) == 1 {
		cs := csList[0].(map[string]any)

		settings["aws_secret_access_key"] = cs["aws_secret_access_key"]
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
	var settings map[string]any

	switch s := response.DataSource.GetConnectionSettings().(type) {
	case *pc.DataSourceDataConnectionSettingsWebhookConnectionSettings:
		settings = map[string]any{
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
			settings["access_control_enabled"] = response.DataSource.DataPools.Nodes[0].AccessControlEnabled
		}

		if s.GetTableSettings() != nil {
			settings["table_settings"] = []map[string]any{internal.ParseTableSettings(s.GetTableSettings().TableSettingsData)}
		}

		if err := d.Set("webhook_connection_settings", []map[string]any{settings}); err != nil {
			return diag.FromErr(err)
		}
	default:
		return diag.Errorf("Missing WebhookConnectionSettings")
	}

	return nil
}

func handleKafkaConnectionSettings(response *pc.DataSourceResponse, d *schema.ResourceData) diag.Diagnostics {
	switch s := response.DataSource.GetConnectionSettings().(type) {
	case *pc.DataSourceDataConnectionSettingsKafkaConnectionSettings:
		settings := map[string]any{
			"auth":              s.GetAuth(),
			"user":              s.GetUser(),
			"password":          s.GetPassword(),
			"tls":               s.GetTls(),
			"bootstrap_servers": s.GetBootstrapServers(),
		}

		if err := d.Set("kafka_connection_settings", []map[string]any{settings}); err != nil {
			return diag.FromErr(err)
		}
	default:
		return diag.Errorf("Missing KafkaConnectionSettings")
	}

	return nil
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

func resourceKafkaDataSourceUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(graphql.Client)

	id := d.Id()
	input := &pc.ModifyKafkaDataSourceInput{
		IdOrUniqueName: &pc.IdOrUniqueName{Id: &id},
	}

	if d.HasChanges("unique_name", "description") {
		uniqueName := d.Get("unique_name").(string)
		description := d.Get("description").(string)

		input.UniqueName = &uniqueName
		input.Description = &description
	}

	if d.HasChanges("kafka_connection_settings") {
		connectionSettings := d.Get("kafka_connection_settings.0").(map[string]any)

		bootstrapServers := make([]string, 0)
		if v, exists := connectionSettings["bootstrap_servers"]; exists {
			for _, bServer := range v.(*schema.Set).List() {
				bootstrapServers = append(bootstrapServers, bServer.(string))
			}
		}

		csPartialInput := &pc.PartialKafkaConnectionSettingsInput{
			BootstrapServers: bootstrapServers,
		}

		if def, ok := connectionSettings["auth"]; ok {
			auth := def.(string)
			csPartialInput.Auth = &auth
		}

		if def, ok := connectionSettings["user"]; ok {
			user := def.(string)
			csPartialInput.User = &user
		}

		if def, ok := connectionSettings["password"]; ok {
			password := def.(string)
			csPartialInput.Password = &password
		}

		if def, ok := connectionSettings["tls"]; ok {
			tls := def.(bool)
			csPartialInput.Tls = &tls
		}

		input.ConnectionSettings = csPartialInput
	}

	if _, err := pc.ModifyKafkaDataSource(ctx, c, input); err != nil {
		return diag.FromErr(err)
	}

	timeout := d.Timeout(schema.TimeoutCreate)

	if err := waitForDataSourceConnected(ctx, c, d.Id(), timeout); err != nil {
		return diag.FromErr(err)
	}
	return resourceDataSourceRead(ctx, d, m)
}

func resourceDataSourceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var err error
	c := meta.(graphql.Client)

	dataSourceType := d.Get("type").(string)
	switch strings.ToUpper(dataSourceType) {
	case "SNOWFLAKE":
		err = internal.SnowflakeDataSourceUpdate(ctx, d, c)
	case "HTTP":
		err = internal.HttpDataSourceUpdate(ctx, d, c)
	case "S3":
		return resourceS3DataSourceUpdate(ctx, d, meta)
	case "WEBHOOK":
		err = internal.WebhookDataSourceUpdate(ctx, d, c)
	case "KAFKA":
		return resourceKafkaDataSourceUpdate(ctx, d, meta)
	case "CLICKHOUSE":
		err = internal.ClickHouseDataSourceUpdate(ctx, d, c)
	default:
		err = fmt.Errorf("unsupported Data Source type \"%v\"", dataSourceType)
	}

	if err != nil {
		return diag.FromErr(err)
	}

	timeout := d.Timeout(schema.TimeoutCreate)

	if err := waitForDataSourceConnected(ctx, c, d.Id(), timeout); err != nil {
		return diag.FromErr(err)
	}
	return resourceDataSourceRead(ctx, d, meta)

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
	createStateConf := &retry.StateChangeConf{
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
