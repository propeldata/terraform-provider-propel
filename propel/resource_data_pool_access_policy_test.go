package propel

import (
	"context"
	"testing"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	pc "github.com/propeldata/terraform-provider-propel/propel_client"
)

func TestAccPropelDataPoolAccessPolicyBasic(t *testing.T) {
	ctx := map[string]any{}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckPropelDataPoolAccessPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPropelDataPoolAccessPolicyConfigBasic(ctx),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPropelDataPoolExists("propel_data_pool_access_policy.baz"),
					resource.TestCheckResourceAttr("propel_data_pool.bar", "access_control_enabled", "true"),
					resource.TestCheckResourceAttr("propel_data_pool_access_policy.baz", "unique_name", "terraform-test-5"),
					resource.TestCheckResourceAttr("propel_data_pool_access_policy.baz", "description", "This is an example of a Data Pool Access Policy"),
				),
			},
		},
	})
}

func testAccCheckPropelDataPoolAccessPolicyConfigBasic(ctx map[string]any) string {
	// language=hcl-terraform
	return Nprintf(`
	resource "propel_data_source" "foo" {
		unique_name = "terraform-test-5"
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

			column {
				name = "account_id"
				type = "STRING"
				nullable = false
			}
		}
	}

	resource "propel_data_pool" "bar" {
		unique_name = "terraform-test-5"
		table = "${propel_data_source.foo.table[0].name}"

		column {
			name = "timestamp_tz"
			type = "TIMESTAMP"
			nullable = false
		}
		column {
			name = "account_id"
			type = "STRING"
			nullable = false
		}
		timestamp = "${propel_data_source.foo.table[0].column[0].name}"
		data_source = "${propel_data_source.foo.id}"
		access_control_enabled = true
	}
	
	resource "propel_data_pool_access_policy" "baz" {
		unique_name = "terraform-test-5"
		description = "This is an example of a Data Pool Access Policy"
		data_pool   = propel_data_pool.bar.id
		
		columns = ["*"]

		row {
		    column   = "account_id"
			operator = "IS_NOT_NULL"
		}

		applications = ["APP01FXSFB8AJJBNKCAQTFV3J29G4"]
	}`, ctx)
}

func testAccCheckPropelDataPoolAccessPolicyDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(graphql.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "propel_data_pool_access_policy" {
			continue
		}

		dataPoolAccessPolicyID := rs.Primary.ID

		_, err := pc.DeleteDataPoolAccessPolicy(context.Background(), c, dataPoolAccessPolicyID)
		if err != nil {
			return err
		}
	}

	return nil
}
