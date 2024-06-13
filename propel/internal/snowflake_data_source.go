package internal

import (
	"context"
	"errors"
	"fmt"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	pc "github.com/propeldata/terraform-provider-propel/propel_client"
)

func SnowflakeDataSourceSchema() *schema.Schema {
	return &schema.Schema{

		Type:          schema.TypeList,
		Optional:      true,
		ConflictsWith: []string{"http_connection_settings", "s3_connection_settings", "webhook_connection_settings", "kafka_connection_settings", "clickhouse_connection_settings"},
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
	}
}

func SnowflakeDataSourceCreate(ctx context.Context, d *schema.ResourceData, c graphql.Client) (string, error) {
	uniqueName := d.Get("unique_name").(string)
	description := d.Get("description").(string)
	connectionSettings := d.Get("snowflake_connection_settings.0").(map[string]any)

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
		return "", err
	}

	switch r := (*response.GetCreateSnowflakeDataSource()).(type) {
	case *pc.CreateSnowflakeDataSourceCreateSnowflakeDataSourceDataSourceResponse:
		return r.DataSource.Id, nil
	case *pc.CreateSnowflakeDataSourceCreateSnowflakeDataSourceFailureResponse:
		return "", fmt.Errorf("failed to create Snowflake Data Source: %s", r.GetError().GetMessage())
	default:
		return "", errors.New("received an unexpected response when creating Snowflake Data Source")
	}
}

func SnowflakeDataSourceUpdate(ctx context.Context, d *schema.ResourceData, c graphql.Client) error {
	id := d.Id()
	input := &pc.ModifySnowflakeDataSourceInput{
		IdOrUniqueName: &pc.IdOrUniqueName{Id: &id},
	}

	if d.HasChanges("unique_name", "description") {
		uniqueName := d.Get("unique_name").(string)
		description := d.Get("description").(string)

		input.UniqueName = &uniqueName
		input.Description = &description
	}

	if d.HasChanges("snowflake_connection_settings") {
		connectionSettings := d.Get("snowflake_connection_settings.0").(map[string]any)
		partialInput := &pc.PartialSnowflakeConnectionSettingsInput{}

		if def, ok := connectionSettings["account"]; ok && def.(string) != "" {
			account := def.(string)
			partialInput.Account = &account
		}

		if def, ok := connectionSettings["database"]; ok && def.(string) != "" {
			database := def.(string)
			partialInput.Database = &database
		}

		if def, ok := connectionSettings["warehouse"]; ok && def.(string) != "" {
			warehouse := def.(string)
			partialInput.Warehouse = &warehouse
		}

		if def, ok := connectionSettings["schema"]; ok && def.(string) != "" {
			schemaF := def.(string)
			partialInput.Schema = &schemaF
		}

		if def, ok := connectionSettings["role"]; ok && def.(string) != "" {
			role := def.(string)
			partialInput.Role = &role
		}

		if def, ok := connectionSettings["username"]; ok && def.(string) != "" {
			username := def.(string)
			partialInput.Username = &username
		}

		if def, ok := connectionSettings["password"]; ok && def.(string) != "" {
			password := def.(string)
			partialInput.Password = &password
		}

		input.ConnectionSettings = partialInput
	}

	_, err := pc.ModifySnowflakeDataSource(ctx, c, input)
	return err
}

func HandleSnowflakeConnectionSettings(response *pc.DataSourceResponse, d *schema.ResourceData) error {
	cs := d.Get("snowflake_connection_settings.0").(map[string]any)
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

		if err := d.Set("snowflake_connection_settings", []map[string]any{settings}); err != nil {
			return err
		}
	default:
		return errors.New("missing SnowflakeConnectionSettings")
	}

	return nil
}
