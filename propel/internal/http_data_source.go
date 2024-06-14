package internal

import (
	"context"
	"errors"
	"fmt"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	pc "github.com/propeldata/terraform-provider-propel/propel_client"
)

func HttpDataSourceSchema() *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeList,
		Optional:      true,
		ConflictsWith: []string{"snowflake_connection_settings", "s3_connection_settings", "webhook_connection_settings", "kafka_connection_settings", "clickhouse_connection_settings"},
		MaxItems:      1,
		Elem: &schema.Resource{
			Description: "HTTP connection settings. Specify these for HTTP Data Sources.",
			Schema: map[string]*schema.Schema{
				"basic_auth": basicAuthSchema(),
			},
		},
	}
}

func HttpDataSourceCreate(ctx context.Context, d *schema.ResourceData, c graphql.Client) (string, error) {
	input := &pc.CreateHttpDataSourceInput{}

	if v, ok := d.GetOk("unique_name"); ok && v.(string) != "" {
		uniqueName := v.(string)
		input.UniqueName = &uniqueName
	}

	if v, ok := d.GetOk("description"); ok && v.(string) != "" {
		description := v.(string)
		input.Description = &description
	}

	if v, ok := d.GetOk("http_connection_settings.0"); ok {
		connectionSettings := v.(map[string]any)

		var basicAuth *pc.HttpBasicAuthInput
		if def, ok := connectionSettings["basic_auth"]; ok && len(def.([]any)) > 0 {
			basicAuth = expandBasicAuth(def.([]any))
		}

		tables := make([]*pc.HttpDataSourceTableInput, 0)
		if def, ok := d.Get("table").([]any); ok && len(def) > 0 {
			tables = expandHttpTables(def)
		}

		input.ConnectionSettings = &pc.HttpConnectionSettingsInput{
			BasicAuth: basicAuth,
			Tables:    tables,
		}
	}

	response, err := pc.CreateHttpDataSource(ctx, c, input)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP Data Source: %w", err)
	}

	return response.CreateHttpDataSource.DataSource.Id, nil
}

func HttpDataSourceUpdate(ctx context.Context, d *schema.ResourceData, c graphql.Client) error {
	id := d.Id()
	input := &pc.ModifyHttpDataSourceInput{
		IdOrUniqueName: &pc.IdOrUniqueName{Id: &id},
	}

	if d.HasChanges("unique_name", "description") {
		uniqueName := d.Get("unique_name").(string)
		description := d.Get("description").(string)

		input.UniqueName = &uniqueName
		input.Description = &description
	}

	if d.HasChanges("table", "http_connection_settings") {
		connectionSettings := d.Get("http_connection_settings.0").(map[string]any)

		var basicAuth *pc.HttpBasicAuthInput
		if def, ok := connectionSettings["basic_auth"]; ok {
			basicAuth = expandBasicAuth(def.([]any))
		}

		tables := make([]*pc.HttpDataSourceTableInput, 0)
		if def, ok := d.GetOk("table"); ok {
			tables = expandHttpTables(def.([]any))
		}

		input.ConnectionSettings = &pc.PartialHttpConnectionSettingsInput{
			BasicAuth: basicAuth,
			Tables:    tables,
		}
	}

	if _, err := pc.ModifyHttpDataSource(ctx, c, input); err != nil {
		return fmt.Errorf("failed to modify HTTP Data Source: %w", err)
	}

	return nil
}

func HandleHttpConnectionSettings(response *pc.DataSourceResponse, d *schema.ResourceData) error {
	if _, exists := d.GetOk("http_connection_settings.0"); !exists {
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
			return err
		}

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

		if err := d.Set("table", tables); err != nil {
			return err
		}
	default:
		return errors.New("missing HttpConnectionSettings")
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
