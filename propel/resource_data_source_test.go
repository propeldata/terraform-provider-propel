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
	httpCtx := map[string]interface{}{
		"resource_name": "new",
		"unique_name":   acctest.RandString(10),
	}

	s3CtxInvalid := map[string]interface{}{
		"resource_name": "fizz",
		"unique_name":   acctest.RandString(10),
	}

	snowflakeCtxInvalid := map[string]interface{}{
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
			{
				Config: testAccCheckPropelDataSourceConfigBasic(httpCtx),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPropelDataSourceExists("propel_data_source.new"),
					resource.TestCheckResourceAttr("propel_data_source.new", "description", ""),
					resource.TestCheckResourceAttr("propel_data_source.new", "type", "Http"),
					resource.TestCheckResourceAttr("propel_data_source.new", "status", "CONNECTED"),
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
		},
	})
}

func testAccCheckPropelDataSourceConfigBasic(ctx map[string]interface{}) string {
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

func testAccCheckPropelDataSourceS3ConfigBroken(ctx map[string]interface{}) string {
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

func testAccCheckPropelDataSourceSnowflakeConfigBroken(ctx map[string]interface{}) string {
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
