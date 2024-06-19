package propel

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	pc "github.com/propeldata/terraform-provider-propel/propel_client"
)

func TestAccPropelDataSourceBasic(t *testing.T) {
	httpCtx := map[string]any{
		"resource_name": "new",
		"unique_name":   acctest.RandString(10),
	}

	s3CtxInvalid := map[string]any{
		"resource_name": "fizz",
		"unique_name":   acctest.RandString(10),
	}

	webhookCtx := map[string]any{
		"resource_name": "webhook",
		"unique_name":   acctest.RandString(10),
	}

	snowflakeCtxInvalid := map[string]any{
		"resource_name":       "foo",
		"unique_name":         acctest.RandString(10),
		"snowflake_account":   "invalid-account",
		"snowflake_database":  "invalid-database",
		"snowflake_warehouse": "invalid-warehouse",
		"snowflake_schema":    "invalid-schema",
		"snowflake_role":      "invalid-role",
		"snowflake_username":  "invalid-username",
		"snowflake_password":  "invalid-password",
	}

	kafkaCtxInvalid := map[string]any{
		"resource_name":          "kafka",
		"unique_name":            acctest.RandString(10),
		"kafka_auth":             "PLAIN",
		"kafka_user":             "invalid-user",
		"kafka_password":         "invalid-password",
		"kafka_bootstrap_server": "192.168.90.84:9092",
	}

	clickHouseCtxInvalid := map[string]any{
		"resource_name":       "clickhouse",
		"unique_name":         acctest.RandString(10),
		"clickhouse_url":      "http://192.168.90.84:8123",
		"clickhouse_database": "invalid-database",
		"clickhouse_user":     "invalid-user",
		"clickhouse_password": "invalid-password",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckPropelDataSourceDestroy,
		Steps: []resource.TestStep{
			// should create the Data Source
			{
				Config: testAccCheckPropelDataSourceConfigBasic(httpCtx),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPropelDataSourceExists("propel_data_source.new"),
					resource.TestCheckResourceAttr("propel_data_source.new", "description", ""),
					resource.TestCheckResourceAttr("propel_data_source.new", "type", "Http"),
					resource.TestCheckResourceAttr("propel_data_source.new", "status", "CONNECTED"),
					resource.TestCheckResourceAttr("propel_data_source.new", "table.0.column.#", "1"),
				),
			},
			// should apply an update to the data source table schema
			{
				Config: testAccUpdatePropelDataSourceConfigBasic(httpCtx),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPropelDataSourceExists("propel_data_source.new"),
					resource.TestCheckResourceAttr("propel_data_source.new", "table.0.column.#", "2"),
				),
			},
			{
				Config:      testAccCheckPropelDataSourceS3ConfigBroken(s3CtxInvalid),
				ExpectError: regexp.MustCompile(`unexpected state 'BROKEN'`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPropelDataSourceExists("propel_data_source.fizz"),
					resource.TestCheckResourceAttr("propel_data_source.fizz", "type", "S3"),
					resource.TestCheckResourceAttr("propel_data_source.fizz", "status", "BROKEN"),
					resource.TestCheckResourceAttr("propel_data_source.fizz", "s3_connection_settings.0.aws_access_key_id", "whatever"),
				),
			},
			{
				Config:      testAccCheckPropelDataSourceSnowflakeConfigBroken(snowflakeCtxInvalid),
				ExpectError: regexp.MustCompile(`unexpected state 'BROKEN'`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPropelDataSourceExists("propel_data_source.foo"),
					resource.TestCheckResourceAttr("propel_data_source.foo", "type", "Snowflake"),
					resource.TestCheckResourceAttr("propel_data_source.foo", "status", "BROKEN"),
					resource.TestCheckResourceAttr("propel_data_source.foo", "snowflake_connection_settings.0.database", "invalid-database"),
				),
			},
			// should create Webhook data source
			{
				Config: testAccWebhookDataSourceBasic(webhookCtx),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPropelDataSourceExists("propel_data_source.webhook"),
					resource.TestCheckResourceAttr("propel_data_source.webhook", "type", "Webhook"),
					resource.TestCheckResourceAttr("propel_data_source.webhook", "status", "CONNECTED"),
					resource.TestCheckResourceAttr("propel_data_source.webhook", "webhook_connection_settings.0.timestamp", "timestamp_tz"),
				),
			},
			// should add a column to the Webhook data pool
			{
				Config: testAccUpdateWebhookDataSourceBasic(webhookCtx),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPropelDataSourceExists("propel_data_source.webhook"),
					resource.TestCheckResourceAttr("propel_data_source.webhook", "type", "Webhook"),
					resource.TestCheckResourceAttr("propel_data_source.webhook", "status", "CONNECTED"),
					resource.TestCheckResourceAttr("propel_data_source.webhook", "webhook_connection_settings.0.column.3.name", "new"),
				),
			},
			// should create Kafka Data Source
			{
				Config:      testAccCheckPropelDataSourceKafkaConfigBroken(kafkaCtxInvalid),
				ExpectError: regexp.MustCompile(`unexpected state 'BROKEN'`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPropelDataSourceExists("propel_data_source.kafka"),
					resource.TestCheckResourceAttr("propel_data_source.kafka", "type", "Kafka"),
					resource.TestCheckResourceAttr("propel_data_source.kafka", "status", "BROKEN"),
					resource.TestCheckResourceAttr("propel_data_source.kafka", "kafka_connection_settings.0.auth", "PLAIN"),
				),
			},
			// should create ClickHouse Data Source
			{
				Config:      testAccCheckPropelDataSourceClickHouseConfigBroken(clickHouseCtxInvalid),
				ExpectError: regexp.MustCompile(`unexpected state 'BROKEN'`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPropelDataSourceExists("propel_data_source.clickhouse"),
					resource.TestCheckResourceAttr("propel_data_source.clickhouse", "type", "ClickHouse"),
					resource.TestCheckResourceAttr("propel_data_source.clickhouse", "status", "BROKEN"),
					resource.TestCheckResourceAttr("propel_data_source.clickhouse", "clickhouse_connection_settings.0.database", "invalid-database"),
				),
			},
		},
	})
}

func testAccCheckPropelDataSourceConfigBasic(ctx map[string]any) string {
	return Nprintf(`
	resource "propel_data_source" "%{resource_name}" {
		unique_name = "%{unique_name}"
		type = "HTTP"

		table {
			name = "CLUSTER_TEST_TABLE_1"

			column {
				name = "timestamp_tz"
				type = "TIMESTAMP"
				nullable = false
			}
		}
	}`, ctx)
}

func testAccUpdatePropelDataSourceConfigBasic(ctx map[string]any) string {
	return Nprintf(`
	resource "propel_data_source" "%{resource_name}" {
		unique_name = "%{unique_name}"
		type = "HTTP"

		table {
			name = "CLUSTER_TEST_TABLE_1"

			column {
				name = "timestamp_tz"
				type = "TIMESTAMP"
				nullable = false
			}

			column {
				name = "id"
				type = "STRING"
				nullable = false
			}
		}
	}`, ctx)
}

func testAccCheckPropelDataSourceS3ConfigBroken(ctx map[string]any) string {
	return Nprintf(`
	resource "propel_data_source" "%{resource_name}" {
		unique_name = "%{unique_name}"
		type = "S3"

		s3_connection_settings {
			bucket = "whatever"
			aws_access_key_id = "whatever"
			aws_secret_access_key = "whatever"
		}

		table {
			name = "CLUSTER_TEST_TABLE_1"
			path = "foo/*.parquet"

			column {
				name = "timestamp_tz"
				type = "TIMESTAMP"
				nullable = false
			}
		}
	}`, ctx)
}

func testAccCheckPropelDataSourceClickHouseConfigBroken(ctx map[string]any) string {
	return Nprintf(`
	resource "propel_data_source" "%{resource_name}" {
		unique_name = "%{unique_name}"
		type = "CLICKHOUSE"

		clickhouse_connection_settings {
			url = "%{clickhouse_url}"
			database = "%{clickhouse_database}"
			user = "%{clickhouse_user}"
			password = "%{clickhouse_password}"
		}
	}`, ctx)
}

func testAccCheckPropelDataSourceKafkaConfigBroken(ctx map[string]any) string {
	return Nprintf(`
	resource "propel_data_source" "%{resource_name}" {
		unique_name = "%{unique_name}"
		type = "KAFKA"

		kafka_connection_settings {
			auth = "%{kafka_auth}"
			user = "%{kafka_user}"
			password = "%{kafka_password}"
			bootstrap_servers = ["%{kafka_bootstrap_server}"]
		}
	}`, ctx)
}

func testAccCheckPropelDataSourceSnowflakeConfigBroken(ctx map[string]any) string {
	return Nprintf(`
	resource "propel_data_source" "%{resource_name}" {
		unique_name = "%{unique_name}"
		type = "SNOWFLAKE"

		snowflake_connection_settings {
			account = "%{snowflake_account}"
			database = "%{snowflake_database}"
			warehouse = "%{snowflake_warehouse}"
			schema = "%{snowflake_schema}"
			role = "%{snowflake_role}"
			username = "%{snowflake_username}"
			password = "%{snowflake_password}"
		}
	}`, ctx)
}

func testAccWebhookDataSourceBasic(ctx map[string]any) string {
	return Nprintf(`
	resource "propel_data_source" "%{resource_name}" {
		unique_name = "%{unique_name}"
		type = "WEBHOOK"

		webhook_connection_settings {
			timestamp = "timestamp_tz"

			column {
				name = "id"
				type = "STRING"
				nullable = false
				json_property = "id"
			}

			column {
				name = "customer_id"
				type = "STRING"
				nullable = false
				json_property = "customer_id"
			}

			column {
				name = "timestamp_tz"
				type = "TIMESTAMP"
				nullable = false
				json_property = "timestamp_tz"
			}

			basic_auth {
				username = "foo"
				password = "bar"
			}
	
			unique_id = "id"
			tenant = "customer_id"
		}
		
	}`, ctx)
}

func testAccUpdateWebhookDataSourceBasic(ctx map[string]any) string {
	return Nprintf(`
	resource "propel_data_source" "%{resource_name}" {
		unique_name = "%{unique_name}"
		type = "WEBHOOK"

		webhook_connection_settings {
			timestamp = "timestamp_tz"

			column {
				name = "id"
				type = "STRING"
				nullable = false
				json_property = "id"
			}

			column {
				name = "customer_id"
				type = "STRING"
				nullable = false
				json_property = "customer_id"
			}

			column {
				name = "timestamp_tz"
				type = "TIMESTAMP"
				nullable = false
				json_property = "timestamp_tz"
			}

			column {
				name = "new"
				type = "STRING"
				nullable = true
				json_property = "new_column"
			}

			basic_auth {
				username = "foo"
				password = "bar"
			}

			access_control_enabled = true
	
			unique_id = "id"
			tenant = "customer_id"
		}
		
	}`, ctx)
}

func testAccCheckPropelDataSourceDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(graphql.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "propel_data_source" {
			continue
		}

		dataSourceID := rs.Primary.ID

		_, err := pc.DeleteDataSource(context.Background(), c, dataSourceID)
		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckPropelDataSourceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no Data Source ID set")
		}

		return nil
	}
}
