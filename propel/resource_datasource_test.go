package propel

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	pc "github.com/propeldata/terraform-provider/graphql_client"
)

func TestAccPropelDataSourceBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckPropelDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckPropelDataSourceConfigBasic(),
				ExpectError: regexp.MustCompile(`DataSource in BROKEN status`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPropelDataSourceExists("propel_datasource.new"),
					resource.TestCheckResourceAttr("propel_datasource.new", "unique_name", "test"),
					resource.TestCheckResourceAttr("propel_datasource.new", "status", "BROKEN"),
				),
			},
		},
	})
}

func testAccCheckPropelDataSourceConfigBasic() string {
	return fmt.Sprintf(`
	resource "propel_datasource" "new" {
		unique_name = "test"

		connection_settings {
			account = "snowflake-account"
			database = "snowflake-database"
			warehouse = "snowflake-warehouse"
			schema = "snowflake-schema"
			role = "snowflake-role"
			username = "snowflake-username"
			password = "snowflake-password"
		}
	}`)
}

func testAccCheckPropelDataSourceDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(graphql.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "propel_datasource" {
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
