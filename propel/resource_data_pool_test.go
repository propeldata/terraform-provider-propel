package propel

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	pc "github.com/propeldata/terraform-provider-propel/propel_client"
)

func TestAccPropelDataPoolBasic(t *testing.T) {
	ctx := map[string]interface{}{}

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
		unique_name = "terraform-test-3"
		type = "Http"

		http_connection_settings {
			basic_auth {
				username = "foo"
				password = "bar"
			}
		}

		table {
			name = "CLUSTER_TEST_TABLE_1"

			column {
				name = "timestamp_tz"
				type = "TIMESTAMP"
				nullable = false
			}
		}
	}

	resource "propel_data_pool" "bar" {
		unique_name = "terraform-test-3"
		table = "${propel_data_source.foo.table[0].name}"

		column {
			name = "timestamp_tz"
			type = "TIMESTAMP"
			nullable = false
		}
		timestamp = "${propel_data_source.foo.table[0].column[0].name}"
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
			return errors.New("no Data Pool ID set")
		}

		return nil
	}
}
