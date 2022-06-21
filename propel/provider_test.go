package propel

import (
	"context"
	"fmt"
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

const (
	snowflakeTestAccountEnvVar   = "PROPEL_TEST_SNOWFLAKE_ACCOUNT"
	snowflakeTestWarehouseEnvVar = "PROPEL_TEST_SNOWFLAKE_WAREHOUSE"
	snowflakeTestRoleEnvVar      = "PROPEL_TEST_SNOWFLAKE_ROLE"
	snowflakeTestUsernameEnvVar  = "PROPEL_TEST_SNOWFLAKE_USERNAME"
	snowflakeTestPasswordEnvVar  = "PROPEL_TEST_SNOWFLAKE_PASSWORD"
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
	if err := os.Getenv("PROPEL_CLIENT_ID"); err == "" {
		t.Fatal("PROPEL_CLIENT_ID must be set for acceptance tests")
	}
	if err := os.Getenv("PROPEL_CLIENT_SECRET"); err == "" {
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
// Nprintf("Hello %{target}!", map[string]interface{}{"target":"world"}) == "Hello world!".
func Nprintf(format string, params map[string]interface{}) string {
	for key, val := range params {
		format = strings.Replace(format, "%{"+key+"}", fmt.Sprintf("%v", val), -1)
	}
	return format
}

func getTestSnowflakeAccountFromEnv(t *testing.T) string {
	skipIfEnvNotSet(t, snowflakeTestAccountEnvVar)

	if v := os.Getenv(snowflakeTestAccountEnvVar); v != "" {
		return v
	}

	return ""
}

func getTestSnowflakeWarehouseFromEnv(t *testing.T) string {
	skipIfEnvNotSet(t, snowflakeTestWarehouseEnvVar)

	if v := os.Getenv(snowflakeTestWarehouseEnvVar); v != "" {
		return v
	}

	return ""
}

func getTestSnowflakeRoleFromEnv(t *testing.T) string {
	skipIfEnvNotSet(t, snowflakeTestRoleEnvVar)

	if v := os.Getenv(snowflakeTestRoleEnvVar); v != "" {
		return v
	}

	return ""
}

func getTestSnowflakeUsernameFromEnv(t *testing.T) string {
	skipIfEnvNotSet(t, snowflakeTestUsernameEnvVar)

	if v := os.Getenv(snowflakeTestUsernameEnvVar); v != "" {
		return v
	}

	return ""
}

func getTestSnowflakePasswordFromEnv(t *testing.T) string {
	skipIfEnvNotSet(t, snowflakeTestPasswordEnvVar)

	if v := os.Getenv(snowflakeTestPasswordEnvVar); v != "" {
		return v
	}

	return ""
}
