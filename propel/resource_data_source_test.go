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

	pc "github.com/propeldata/terraform-provider/propel_client"
)

func TestAccPropelDataSourceBasic(t *testing.T) {
	ctx := map[string]interface{}{
		"resource_name":       "new",
		"unique_name":         acctest.RandString(10),
		"snowflake_account":   getTestSnowflakeAccountFromEnv(t),
		"snowflake_database":  "CLUSTER_TESTS",
		"snowflake_warehouse": getTestSnowflakeWarehouseFromEnv(t),
		"snowflake_schema":    "CLUSTER_TESTS",
		"snowflake_role":      getTestSnowflakeRoleFromEnv(t),
		"snowflake_username":  getTestSnowflakeUsernameFromEnv(t),
		"snowflake_password":  getTestSnowflakePasswordFromEnv(t),
	}

	ctxInvalid := map[string]interface{}{
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
				Config: testAccCheckPropelDataSourceConfigBasic(ctx),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPropelDataSourceExists("propel_data_source.new"),
					resource.TestCheckResourceAttr("propel_data_source.new", "description", ""),
					resource.TestCheckResourceAttr("propel_data_source.new", "type", "Snowflake"),
					resource.TestCheckResourceAttr("propel_data_source.new", "status", "CONNECTED"),
				),
			},
			{
				Config:      testAccCheckPropelDataSourceConfigBasic(ctxInvalid),
				ExpectError: regexp.MustCompile(`DataSource in BROKEN status`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPropelDataSourceExists("propel_data_source.foo"),
					resource.TestCheckResourceAttr("propel_data_source.new", "type", "Snowflake"),
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

		connection_settings {
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
			return errors.New("no DataSourceID set")
		}

		return nil
	}
}
