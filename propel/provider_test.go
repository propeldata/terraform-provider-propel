package propel

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	testAccProvider          *schema.Provider
	testAccProviderFactories map[string]func() (*schema.Provider, error)
)

func init() {
	testAccProvider = Provider()
	testAccProviderFactories = map[string]func() (*schema.Provider, error){
		"propel": func() (*schema.Provider, error) {
			return testAccProvider, nil
		},
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("PROPEL_CLIENT_ID"); v == "" {
		t.Fatal("PROPEL_CLIENT_ID must be set for acceptance tests")
	}

	if v := os.Getenv("PROPEL_CLIENT_SECRET"); v == "" {
		t.Fatal("PROPEL_CLIENT_SECRET must be set for acceptance tests")
	}

	err := testAccProvider.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
	if err != nil {
		t.Fatal(err)
	}
}

func skipIfEnvNotSet(t *testing.T, env string) {
	if t == nil {
		log.Println("[DEBUG] Not running inside of test")
		return
	}

	if os.Getenv(env) == "" {
		log.Printf("[DEBUG] Warning - environment variable %s is not set - skipping test %s", env, t.Name())
		t.Skipf("Environment variable %s is not set", env)
	}
}

// This is a Printf sibling (Nprintf; Named Printf), which handles strings like
// Nprintf("Hello %{target}!", map[string]any{"target":"world"}) == "Hello world!".
func Nprintf(format string, params map[string]any) string {
	for key, val := range params {
		format = strings.Replace(format, "%{"+key+"}", fmt.Sprintf("%v", val), -1)
	}
	return format
}

func testAccCheckPropelResourceExists(resourceId string, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceId]

		if !ok {
			return fmt.Errorf("not found: %s", resourceId)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no %s ID set", resourceName)
		}

		return nil
	}
}
