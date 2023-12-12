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
	"github.com/stretchr/testify/assert"

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

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckPropelDataSourceDestroy,
		Steps: []resource.TestStep{
			// should create the data source
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
				),
			},
			{
				Config:      testAccCheckPropelDataSourceSnowflakeConfigBroken(snowflakeCtxInvalid),
				ExpectError: regexp.MustCompile(`unexpected state 'BROKEN'`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPropelDataSourceExists("propel_data_source.foo"),
					resource.TestCheckResourceAttr("propel_data_source.foo", "type", "Snowflake"),
					resource.TestCheckResourceAttr("propel_data_source.foo", "status", "BROKEN"),
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
		},
	})
}

func testAccCheckPropelDataSourceConfigBasic(ctx map[string]any) string {
	return Nprintf(`
	resource "propel_data_source" "%{resource_name}" {
		unique_name = "%{unique_name}"
		type = "Http"

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
		type = "Http"

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

func testAccCheckPropelDataSourceSnowflakeConfigBroken(ctx map[string]any) string {
	return Nprintf(`
	resource "propel_data_source" "%{resource_name}" {
		unique_name = "%{unique_name}"
		type = "Snowflake"

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
		type = "Webhook"

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
		type = "Webhook"

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

func Test_getNewDataSourceColumns(t *testing.T) {
	tests := []struct {
		name               string
		oldItemDef         []any
		newItemDef         []any
		expectedNewColumns map[string]pc.WebhookDataSourceColumnInput
		expectedError      string
	}{
		{
			name: "Successful new columns",
			oldItemDef: []any{
				map[string]any{"name": "COLUMN_A", "type": "STRING", "nullable": false, "json_property": "column_a"},
				map[string]any{"name": "COLUMN_B", "type": "FLOAT", "nullable": false, "json_property": "column_b"},
			},
			newItemDef: []any{
				map[string]any{"name": "COLUMN_A", "type": "STRING", "nullable": false, "json_property": "column_a"},
				map[string]any{"name": "COLUMN_B", "type": "FLOAT", "nullable": false, "json_property": "column_b"},
				map[string]any{"name": "COLUMN_C", "type": "INT64", "nullable": false, "json_property": "column_c"},
				map[string]any{"name": "COLUMN_D", "type": "TIMESTAMP", "nullable": false, "json_property": "column_d"},
			},
			expectedNewColumns: map[string]pc.WebhookDataSourceColumnInput{
				"COLUMN_C": {Name: "COLUMN_C", Type: "INT64", Nullable: false, JsonProperty: "column_c"},
				"COLUMN_D": {Name: "COLUMN_D", Type: "TIMESTAMP", Nullable: false, JsonProperty: "column_d"},
			},
			expectedError: "",
		},
		{
			name: "Repeated column names",
			oldItemDef: []any{
				map[string]any{"name": "COLUMN_A", "type": "STRING", "nullable": false, "json_property": "column_a"},
			},
			newItemDef: []any{
				map[string]any{"name": "COLUMN_B", "type": "FLOAT", "nullable": false, "json_property": "column_b"},
				map[string]any{"name": "COLUMN_B", "type": "TIMESTAMP", "nullable": false, "json_property": "column_sb"},
			},
			expectedNewColumns: map[string]pc.WebhookDataSourceColumnInput{},
			expectedError:      `column "COLUMN_B" already exists`,
		},
		{
			name: "Unsupported column deletion",
			oldItemDef: []any{
				map[string]any{"name": "COLUMN_A", "type": "STRING", "nullable": false, "json_property": "column_a"},
				map[string]any{"name": "COLUMN_B", "type": "FLOAT", "nullable": false, "json_property": "column_b"},
			},
			newItemDef: []any{
				map[string]any{"name": "COLUMN_A", "type": "STRING", "nullable": false, "json_property": "column_a"},
			},
			expectedNewColumns: map[string]pc.WebhookDataSourceColumnInput{},
			expectedError:      `column "COLUMN_B" was removed, column deletions are not supported`,
		},
		{
			name: "Unsupported column update",
			oldItemDef: []any{
				map[string]any{"name": "COLUMN_A", "type": "STRING", "nullable": false, "json_property": "column_a"},
				map[string]any{"name": "COLUMN_B", "type": "FLOAT", "nullable": false, "json_property": "column_b"},
			},
			newItemDef: []any{
				map[string]any{"name": "COLUMN_A", "type": "STRING", "nullable": false, "json_property": "column_z"},
			},
			expectedNewColumns: map[string]pc.WebhookDataSourceColumnInput{},
			expectedError:      `column "COLUMN_A" was modified, column updates are not supported`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(st *testing.T) {
			a := assert.New(st)

			result, err := getNewDataSourceColumns(tt.oldItemDef, tt.newItemDef)
			if tt.expectedError != "" {
				a.Error(err)
				a.EqualError(err, tt.expectedError)
				return
			}

			a.NoError(err)
			a.Equal(tt.expectedNewColumns, result)
		})
	}
}
