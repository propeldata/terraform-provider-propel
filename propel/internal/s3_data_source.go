package internal

import (
	"context"
	"errors"
	"fmt"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	pc "github.com/propeldata/terraform-provider-propel/propel_client"
)

func S3DataSourceSchema() *schema.Schema {
	return &schema.Schema{
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
	}
}

func S3DataSourceCreate(ctx context.Context, d *schema.ResourceData, c graphql.Client) (string, error) {
	input := &pc.CreateS3DataSourceInput{}

	if v, ok := d.GetOk("unique_name"); ok && v.(string) != "" {
		uniqueName := v.(string)
		input.UniqueName = &uniqueName
	}

	if v, ok := d.GetOk("description"); ok && v.(string) != "" {
		description := v.(string)
		input.Description = &description
	}

	tables := make([]*pc.S3DataSourceTableInput, 0)
	if def, ok := d.Get("table").([]any); ok && len(def) > 0 {
		tables = expandS3Tables(def)
	}

	if v, ok := d.GetOk("s3_connection_settings.0"); ok {
		connectionSettings := v.(map[string]any)

		input.ConnectionSettings = &pc.S3ConnectionSettingsInput{
			Bucket:             connectionSettings["bucket"].(string),
			AwsAccessKeyId:     connectionSettings["aws_access_key_id"].(string),
			AwsSecretAccessKey: connectionSettings["aws_secret_access_key"].(string),
			Tables:             tables,
		}
	}

	response, err := pc.CreateS3DataSource(ctx, c, input)
	if err != nil {
		return "", fmt.Errorf("failed to create S3 Data Source: %w", err)
	}

	return response.CreateS3DataSource.DataSource.Id, nil
}

func S3DataSourceUpdate(ctx context.Context, d *schema.ResourceData, c graphql.Client) error {
	id := d.Id()
	input := &pc.ModifyS3DataSourceInput{
		IdOrUniqueName: &pc.IdOrUniqueName{Id: &id},
	}

	if d.HasChanges("unique_name", "description") {
		uniqueName := d.Get("unique_name").(string)
		description := d.Get("description").(string)

		input.UniqueName = &uniqueName
		input.Description = &description
	}

	if d.HasChanges("table", "s3_connection_settings") {
		connectionSettings := d.Get("s3_connection_settings.0").(map[string]any)

		tables := make([]*pc.S3DataSourceTableInput, 0)
		if _, ok := d.GetOk("table"); ok {
			tables = expandS3Tables(d.Get("table").([]any))
		}

		partialInput := &pc.PartialS3ConnectionSettingsInput{
			Tables: tables,
		}

		if def, ok := connectionSettings["bucket"]; ok && def.(string) != "" {
			bucket := def.(string)
			partialInput.Bucket = &bucket
		}

		if def, ok := connectionSettings["aws_access_key_id"]; ok && def.(string) != "" {
			accessKeyID := def.(string)
			partialInput.AwsAccessKeyId = &accessKeyID
		}

		if def, ok := connectionSettings["aws_access_key_id"]; ok && def.(string) != "" {
			accessKeyID := def.(string)
			partialInput.AwsAccessKeyId = &accessKeyID
		}

		if def, ok := connectionSettings["aws_secret_access_key"]; ok && def.(string) != "" {
			secretAccessKey := def.(string)
			partialInput.AwsSecretAccessKey = &secretAccessKey
		}

		input.ConnectionSettings = partialInput
	}

	if _, err := pc.ModifyS3DataSource(ctx, c, input); err != nil {
		return fmt.Errorf("failed to modify S3 Data Source: %w", err)
	}

	return nil
}

func HandleS3ConnectionSettings(response *pc.DataSourceResponse, d *schema.ResourceData) error {
	if _, exists := d.GetOk("s3_connection_settings.0"); !exists {
		return nil
	}

	cs := d.Get("s3_connection_settings.0").(map[string]any)
	settings := map[string]any{
		"aws_secret_access_key": cs["aws_secret_access_key"],
	}

	switch s := response.DataSource.GetConnectionSettings().(type) {
	case *pc.DataSourceDataConnectionSettingsS3ConnectionSettings:
		settings["bucket"] = s.GetBucket()
		settings["aws_access_key_id"] = s.GetAwsAccessKeyId()

		if err := d.Set("s3_connection_settings", []map[string]any{settings}); err != nil {
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
				"path":   table.Path,
				"column": columns,
			})
		}

		if err := d.Set("table", tables); err != nil {
			return err
		}
	default:
		return errors.New("missing S3ConnectionSettings")
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
