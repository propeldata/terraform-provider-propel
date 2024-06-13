package internal

import (
	"context"
	"errors"
	"fmt"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/propeldata/terraform-provider-propel/propel/internal/utils"
	pc "github.com/propeldata/terraform-provider-propel/propel_client"
)

func WebhookDataSourceSchema() *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeList,
		Optional:      true,
		ConflictsWith: []string{"snowflake_connection_settings", "http_connection_settings", "s3_connection_settings", "kafka_connection_settings", "clickhouse_connection_settings"},
		MaxItems:      1,
		Elem: &schema.Resource{
			Description: "Webhook connection settings. Specify these for Webhook Data Sources.",
			Schema: map[string]*schema.Schema{
				"basic_auth": basicAuthSchema(),
				"column": {
					Type:        schema.TypeList,
					Optional:    true,
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
								Description: `
The JSON property that the column will be derived from. For example, if you POST a JSON event like this: 

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
				"access_control_enabled": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Whether the resulting Data Pool has access control enabled or not. If the Data Pool has access control enabled, Applications must be assigned Data Pool Access Policies in order to query the Data Pool and its Metrics.",
				},
				"tenant": {
					Type:        schema.TypeString,
					Optional:    true,
					ForceNew:    true,
					Description: "The tenant ID column, if configured.",
					Deprecated:  "Remove this attribute's configuration as it's no longer in use and the attribute will be removed in the next major version of the provider.",
				},
				"timestamp": {
					Type:        schema.TypeString,
					Optional:    true,
					ForceNew:    true,
					Description: "The primary timestamp column.",
				},
				"unique_id": {
					Type:        schema.TypeString,
					Optional:    true,
					ForceNew:    true,
					Description: "The unique ID column. Propel uses the primary timestamp and a unique ID to compose a primary key for determining whether records should be inserted, deleted, or updated.",
					Deprecated:  "Will be removed; use Table Settings to define the primary key.",
				},
				"table_settings": TableSettingsSchema(),
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
	}
}

func WebhookDataSourceCreate(ctx context.Context, d *schema.ResourceData, c graphql.Client) (string, error) {
	uniqueName := d.Get("unique_name").(string)
	description := d.Get("description").(string)

	input := &pc.CreateWebhookDataSourceInput{
		UniqueName:  &uniqueName,
		Description: &description,
	}

	cs := d.Get("webhook_connection_settings.0").(map[string]any)
	connectionSettingsInput := &pc.WebhookConnectionSettingsInput{}

	if def, ok := cs["basic_auth"]; ok && len(def.([]any)) > 0 {
		connectionSettingsInput.BasicAuth = expandBasicAuth(def.([]any))
	}

	if def, ok := cs["column"].([]any); ok && len(def) > 0 {
		connectionSettingsInput.Columns = expandWebhookColumns(def)
	}

	if t, ok := cs["timestamp"]; ok && t.(string) != "" {
		timestamp := t.(string)
		connectionSettingsInput.Timestamp = &timestamp
	}

	if t, ok := cs["tenant"]; ok && t.(string) != "" {
		tenant := t.(string)
		connectionSettingsInput.Tenant = &tenant
	}

	if u, ok := cs["unique_id"]; ok && u.(string) != "" {
		uniqueID := u.(string)
		connectionSettingsInput.UniqueId = &uniqueID
	}

	if v, ok := cs["access_control_enabled"]; ok && v.(bool) {
		accessControl := v.(bool)
		connectionSettingsInput.AccessControlEnabled = &accessControl
	}

	if v, exists := cs["table_settings"]; exists && len(v.([]any)) > 0 {
		settings := v.([]any)[0].(map[string]any)

		ts, err := BuildTableSettingsInput(settings)
		if err != nil {
			return "", err
		}

		connectionSettingsInput.TableSettings = ts
	}

	response, err := pc.CreateWebhookDataSource(ctx, c, input)
	if err != nil {
		return "", err
	}

	return response.CreateWebhookDataSource.DataSource.Id, nil
}

func WebhookDataSourceUpdate(ctx context.Context, d *schema.ResourceData, c graphql.Client) error {
	id := d.Id()
	input := &pc.ModifyWebhookDataSourceInput{
		IdOrUniqueName: &pc.IdOrUniqueName{Id: &id},
	}

	if d.HasChange("webhook_connection_settings") {
		oldItem, newItem := d.GetChange("webhook_connection_settings")
		oldDef, oldOk := oldItem.([]any)
		newDef, newOk := newItem.([]any)

		if !oldOk || !newOk || len(newDef) < 1 {
			return errors.New("invalid Webhook Connection Settings format")
		}

		oldConnectionSettings, newConnectionSettings := oldDef[0].(map[string]any), newDef[0].(map[string]any)

		oldColumnItem, okOldColumn := oldConnectionSettings["column"]
		newColumnItem, okNewColumn := newConnectionSettings["column"]
		if !okNewColumn || !okOldColumn {
			return errors.New("invalid Webhook columns")
		}

		dataPoolId := oldConnectionSettings["data_pool_id"].(string)

		newColumns, err := newWebhookColumns(oldColumnItem.([]any), newColumnItem.([]any))
		if err != nil {
			return err
		}

		if len(newColumns) > 0 {
			if err := addWebhookColumns(ctx, d, c, dataPoolId, newColumns); err != nil {
				return err
			}
		}

		if newConnectionSettings["access_control_enabled"].(bool) != oldConnectionSettings["access_control_enabled"].(bool) {
			accessControlEnabled := newConnectionSettings["access_control_enabled"].(bool)
			if _, err = pc.ModifyDataPool(ctx, c, &pc.ModifyDataPoolInput{
				IdOrUniqueName:       &pc.IdOrUniqueName{Id: &dataPoolId},
				AccessControlEnabled: &accessControlEnabled,
			}); err != nil {
				return err
			}
		}

		if basicAuthDef, ok := newConnectionSettings["basic_auth"]; ok {
			basicAuthEnabled := len(basicAuthDef.([]any)) > 0

			input.ConnectionSettings = &pc.PartialWebhookConnectionSettingsInput{
				BasicAuthEnabled: &basicAuthEnabled,
			}

			if basicAuthEnabled {
				input.ConnectionSettings.BasicAuth = expandBasicAuth(basicAuthDef.([]any))
			}
		}
	}

	if d.HasChanges("unique_name", "description") {
		uniqueName := d.Get("unique_name").(string)
		description := d.Get("description").(string)

		input.UniqueName = &uniqueName
		input.Description = &description
	}

	_, err := pc.ModifyWebhookDataSource(ctx, c, input)
	return err
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

func newWebhookColumns(oldItemDef []any, newItemDef []any) (map[string]pc.WebhookDataSourceColumnInput, error) {
	newColumns := map[string]pc.WebhookDataSourceColumnInput{}

	for _, rawColumn := range newItemDef {
		column := rawColumn.(map[string]any)
		columnInput := pc.WebhookDataSourceColumnInput{
			Name:         column["name"].(string),
			Type:         pc.ColumnType(column["type"].(string)),
			Nullable:     column["nullable"].(bool),
			JsonProperty: column["json_property"].(string),
		}

		if _, ok := newColumns[columnInput.Name]; ok {
			return nil, fmt.Errorf(`column "%s" already exists`, columnInput.Name)
		}

		newColumns[columnInput.Name] = columnInput
	}

	for _, rawColumn := range oldItemDef {
		column := rawColumn.(map[string]any)
		columnInput := pc.WebhookDataSourceColumnInput{
			Name:         column["name"].(string),
			Type:         pc.ColumnType(column["type"].(string)),
			Nullable:     column["nullable"].(bool),
			JsonProperty: column["json_property"].(string),
		}

		newColumnInput, ok := newColumns[columnInput.Name]
		if !ok {
			return nil, fmt.Errorf(`column "%s" was removed, column deletions are not supported`, columnInput.Name)
		}

		if columnInput.Type != newColumnInput.Type || columnInput.Nullable != newColumnInput.Nullable || columnInput.JsonProperty != newColumnInput.JsonProperty {
			return nil, fmt.Errorf(`column "%s" was modified, column updates are not supported`, columnInput.Name)
		}

		delete(newColumns, columnInput.Name)
	}

	return newColumns, nil
}

func addWebhookColumns(ctx context.Context, d *schema.ResourceData, c graphql.Client, dataPoolId string, newColumns map[string]pc.WebhookDataSourceColumnInput) error {
	for _, newColumn := range newColumns {
		if !newColumn.Nullable {
			return fmt.Errorf(`new column "%s" must be nullable`, newColumn.Name)
		}

		jobResponse, err := pc.CreateAddColumnToDataPoolJob(ctx, c, &pc.CreateAddColumnToDataPoolJobInput{
			DataPool:     dataPoolId,
			ColumnName:   newColumn.Name,
			ColumnType:   newColumn.Type,
			JsonProperty: &newColumn.JsonProperty,
		})
		if err != nil {
			return err
		}

		timeout := d.Timeout(schema.TimeoutUpdate)

		if err = WaitForAddColumnJobSucceeded(ctx, c, jobResponse.CreateAddColumnToDataPoolJob.Job.Id, timeout); err != nil {
			return err
		}
	}

	return nil
}
