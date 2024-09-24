package internal

import (
	"context"
	"errors"
	"fmt"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	pc "github.com/propeldata/terraform-provider-propel/propel_client"
)

func KafkaDataSourceSchema() *schema.Schema {
	return &schema.Schema{
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
					Type:        schema.TypeList,
					Required:    true,
					Description: "The bootstrap server(s) to connect to.",
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
			},
		},
	}
}

func KafkaDataSourceCreate(ctx context.Context, d *schema.ResourceData, c graphql.Client) (string, error) {
	input := &pc.CreateKafkaDataSourceInput{}

	if v, ok := d.GetOk("unique_name"); ok && v.(string) != "" {
		uniqueName := v.(string)
		input.UniqueName = &uniqueName
	}

	if v, ok := d.GetOk("description"); ok && v.(string) != "" {
		description := v.(string)
		input.Description = &description
	}

	if v, ok := d.GetOk("kafka_connection_settings.0"); ok {
		connectionSettings := v.(map[string]any)
		tls := false

		bootstrapServers := make([]string, 0)
		if s, exists := connectionSettings["bootstrap_servers"]; exists {
			for _, bServer := range s.([]any) {
				bootstrapServers = append(bootstrapServers, bServer.(string))
			}
		}

		if t, exists := connectionSettings["tls"]; exists && t.(bool) {
			tls = t.(bool)
		}

		input.ConnectionSettings = &pc.KafkaConnectionSettingsInput{
			Auth:             connectionSettings["auth"].(string),
			User:             connectionSettings["user"].(string),
			Password:         connectionSettings["password"].(string),
			BootstrapServers: bootstrapServers,
			Tls:              &tls,
		}
	}

	response, err := pc.CreateKafkaDataSource(ctx, c, input)
	if err != nil {
		return "", fmt.Errorf("failed to create Kafka Data Source: %w", err)
	}

	return response.CreateKafkaDataSource.DataSource.Id, nil
}

func KafkaDataSourceUpdate(ctx context.Context, d *schema.ResourceData, c graphql.Client) error {
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
			for _, bServer := range v.([]any) {
				bootstrapServers = append(bootstrapServers, bServer.(string))
			}
		}

		csPartialInput := &pc.PartialKafkaConnectionSettingsInput{
			BootstrapServers: bootstrapServers,
		}

		if def, ok := connectionSettings["auth"]; ok && def.(string) != "" {
			auth := def.(string)
			csPartialInput.Auth = &auth
		}

		if def, ok := connectionSettings["user"]; ok && def.(string) != "" {
			user := def.(string)
			csPartialInput.User = &user
		}

		if def, ok := connectionSettings["password"]; ok && def.(string) != "" {
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
		return fmt.Errorf("failed to modify Kafka Data Source: %w", err)
	}

	return nil
}

func HandleKafkaConnectionSettings(response *pc.DataSourceResponse, d *schema.ResourceData) error {
	if _, exists := d.GetOk("kafka_connection_settings.0"); !exists {
		return nil
	}

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
			return err
		}
	default:
		return errors.New("missing KafkaConnectionSettings")
	}

	return nil
}
