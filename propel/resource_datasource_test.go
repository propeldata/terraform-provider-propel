package propel

import (
	"context"
	"fmt"
	"log"
	"testing"

	cms "github.com/propeldata/terraform-provider/cms_graphql_client"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestCreateDataSource(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPropelDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPropelDataSource(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPropelDataSourceExists("propel_datasource.new"),
				),
			},
		},
	})
}

func testAccCheckPropelDataSource() string {
	return fmt.Sprintf(`
	resource "propel_datasource" "new" {
		database = test
	}`)
}

func testAccCheckPropelDataSourceDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(graphql.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "propel_datasource" {
			continue
		}

		dataSourceID := rs.Primary.ID

		_, err := cms.DeleteDataSource(context.Background(), c, dataSourceID)
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
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DataSourceID set")
		}

		return nil
	}
}
