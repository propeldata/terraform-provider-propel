package propel

import (
	"context"
	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	pc "github.com/propeldata/terraform-provider-propel/propel_client"
	"testing"
)

func TestAccPropelApplication(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckPropelApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPropelApplicationConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPropelResourceExists("propel_application.test", "Application"),
					resource.TestCheckResourceAttr("propel_application.test", "propeller", "P1_SMALL"),
				),
			},
		},
	})
}

func testAccCheckPropelApplicationConfig() string {
	return `
		resource "propel_application" "test" {
			scopes = ["METRIC_QUERY"]
			propeller = "P1_SMALL"
		}
    `
}

func testAccCheckPropelApplicationDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(graphql.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "propel_application" {
			continue
		}

		applicationID := rs.Primary.ID

		if _, err := pc.DeleteApplication(context.Background(), c, applicationID); err != nil {
			return err
		}
	}

	return nil
}
