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

func TestAccPropelMaterializedViewBasic(t *testing.T) {
	ctx := map[string]any{}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckPropelMaterializedViewDestroy,
		Steps: []resource.TestStep{
			// should create the Materialized View
			{
				Config: testAccCheckPropelMaterializedViewConfigBasic(ctx),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPropelMaterializedViewExists("propel_materialized_view.foo"),
					resource.TestCheckResourceAttr("propel_materialized_view.foo", "unique_name", "terraform-mv-1"),
				),
			},
		},
	})
}

func testAccCheckPropelMaterializedViewConfigBasic(ctx map[string]any) string {
	// language=hcl-terraform
	return Nprintf(`
		resource "propel_data_source" "terraform-mv-source-dp" {
			unique_name = "terraform-mv-source-dp"
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
		
				unique_id = "id"
			}
		}

		resource "propel_materialized_view" "foo" {
			unique_name = "terraform-mv-1"
			sql = "SELECT customer_id, timestamp_tz AS timestamp FROM \"terraform-mv-source-dp\""
			new_data_pool = {
				unique_name = "terraform-mv-data-pool"
				description = "terraform MV Data Pool"
				timestamp = "timestamp_tz"
				unique_id = "customer_id"
				access_control_enabled = true
				table_settings = {
					engine = {
						type = "SUMMING_MERGE_TREE"
					}
				}
			}
			backfill = false
		}
	`, ctx)
}

func testAccCheckPropelMaterializedViewDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(graphql.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "propel_materialized_view" {
			continue
		}

		materializedViewID := rs.Primary.ID

		_, err := pc.DeleteMaterializedView(context.Background(), c, materializedViewID)
		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckPropelMaterializedViewExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no Materialized View ID set")
		}

		return nil
	}
}
