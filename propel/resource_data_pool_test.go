package propel

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	pc "github.com/propeldata/terraform-provider/propel_client"
)

func TestAccPropelDataPoolBasic(t *testing.T) {
	ctx := map[string]interface{}{
		"snowflake_account":   getTestSnowflakeAccountFromEnv(t),
		"snowflake_database":  "CLUSTER_TESTS",
		"snowflake_warehouse": getTestSnowflakeWarehouseFromEnv(t),
		"snowflake_schema":    "CLUSTER_TESTS",
		"snowflake_role":      getTestSnowflakeRoleFromEnv(t),
		"snowflake_username":  getTestSnowflakeUsernameFromEnv(t),
		"snowflake_password":  getTestSnowflakePasswordFromEnv(t),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckPropelDataPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPropelDataPoolConfigBasic(ctx),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPropelDataPoolExists("propel_data_pool.bar"),
					resource.TestCheckResourceAttr("propel_data_pool.bar", "table", "CLUSTER_TEST_TABLE_1"),
				),
			},
		},
	})
}

func testAccCheckPropelDataPoolConfigBasic(ctx map[string]interface{}) string {
	return Nprintf(`
	resource "propel_data_source" "foo" {
		unique_name = "test"

		connection_settings {
			account = "%{snowflake_account}"
			database = "%{snowflake_database}"
			warehouse = "%{snowflake_warehouse}"
			schema = "%{snowflake_schema}"
			role = "%{snowflake_role}"
			username = "%{snowflake_username}"
			password = "%{snowflake_password}"
		}
	}

	resource "propel_data_pool" "bar" {
		unique_name = "test"
		table = "CLUSTER_TEST_TABLE_1"
		timestamp = "timestamp_tz"
		data_source = "${propel_data_source.foo.id}"
	}`, ctx)
}

func testAccCheckPropelDataPoolDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(graphql.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "propel_data_pool" {
			continue
		}

		dataPoolID := rs.Primary.ID

		_, err := pc.DeleteDataPool(context.Background(), c, dataPoolID)
		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckPropelDataPoolExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no DataPoolID set")
		}

		return nil
	}
}
