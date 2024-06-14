package internal

import (
	"context"
	"errors"
	"fmt"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	pc "github.com/propeldata/terraform-provider-propel/propel_client"
)

func ClickHouseDataSourceSchema() *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeList,
		Optional:      true,
		ConflictsWith: []string{"snowflake_connection_settings", "http_connection_settings", "s3_connection_settings", "webhook_connection_settings", "kafka_connection_settings"},
		MaxItems:      1,
		Elem: &schema.Resource{
			Description: "The connection settings for a ClickHouse Data Source.",
			Schema: map[string]*schema.Schema{
				"url": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The URL where the ClickHouse host is listening to HTTP[S] connections.",
				},
				"user": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The user for authenticating against the ClickHouse host.",
				},
				"password": {
					Type:        schema.TypeString,
					Required:    true,
					Sensitive:   true,
					Description: "The password for the provided user.",
				},
				"database": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Which database to connect to.",
				},
				"readonly": {
					Type:        schema.TypeBool,
					Computed:    true,
					Description: "Whether the user has readonly permissions or not for querying ClickHouse.",
				},
			},
		},
	}
}

func ClickHouseDataSourceCreate(ctx context.Context, d *schema.ResourceData, c graphql.Client) (string, error) {
	input := &pc.CreateClickHouseDataSourceInput{}

	if v, ok := d.GetOk("unique_name"); ok && v.(string) != "" {
		uniqueName := v.(string)
		input.UniqueName = &uniqueName
	}

	if v, ok := d.GetOk("description"); ok && v.(string) != "" {
		description := v.(string)
		input.Description = &description
	}

	if v, ok := d.GetOk("clickhouse_connection_settings.0"); ok {
		connectionSettings := v.(map[string]any)
		input.ConnectionSettings = &pc.ClickHouseConnectionSettingsInput{
			Url:      connectionSettings["url"].(string),
			Database: connectionSettings["database"].(string),
			User:     connectionSettings["user"].(string),
			Password: connectionSettings["password"].(string),
		}
	}

	response, err := pc.CreateClickHouseDataSource(ctx, c, input)
	if err != nil {
		return "", fmt.Errorf("failed to create ClickHouse Data Source: %w", err)
	}

	return response.CreateClickHouseDataSource.DataSource.Id, nil
}

func ClickHouseDataSourceUpdate(ctx context.Context, d *schema.ResourceData, c graphql.Client) error {
	id := d.Id()
	input := &pc.ModifyClickHouseDataSourceInput{
		IdOrUniqueName: &pc.IdOrUniqueName{Id: &id},
	}

	if d.HasChanges("unique_name", "description") {
		uniqueName := d.Get("unique_name").(string)
		description := d.Get("description").(string)

		input.UniqueName = &uniqueName
		input.Description = &description
	}

	if d.HasChanges("clickhouse_connection_settings") {
		connectionSettings := d.Get("clickhouse_connection_settings.0").(map[string]any)
		partialInput := &pc.PartialClickHouseConnectionSettingsInput{}

		if v, ok := connectionSettings["url"]; ok && v.(string) != "" {
			chURL := v.(string)
			partialInput.Url = &chURL
		}

		if v, ok := connectionSettings["database"]; ok && v.(string) != "" {
			database := v.(string)
			partialInput.Database = &database
		}

		if v, ok := connectionSettings["user"]; ok && v.(string) != "" {
			user := v.(string)
			partialInput.User = &user
		}

		if v, ok := connectionSettings["password"]; ok && v.(string) != "" {
			password := v.(string)
			partialInput.Password = &password
		}

		input.ConnectionSettings = partialInput
	}

	if _, err := pc.ModifyClickHouseDataSource(ctx, c, input); err != nil {
		return fmt.Errorf("failed to modify ClickHouse Data Source: %w", err)
	}

	return nil
}

func HandleClickHouseConnectionSettings(response *pc.DataSourceResponse, d *schema.ResourceData) error {
	if _, exists := d.GetOk("clickhouse_connection_settings.0"); !exists {
		return nil
	}

	switch s := response.DataSource.GetConnectionSettings().(type) {
	case *pc.DataSourceDataConnectionSettingsClickHouseConnectionSettings:
		settings := map[string]any{
			"url":      s.GetUrl(),
			"database": s.GetDatabase(),
			"user":     s.GetUser(),
			"password": s.GetPassword(),
			"readonly": s.GetReadonly(),
		}

		if err := d.Set("clickhouse_connection_settings", []map[string]any{settings}); err != nil {
			return err
		}
	default:
		return errors.New("missing ClickHouseConnectionSettings")
	}

	return nil
}
