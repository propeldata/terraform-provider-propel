package utils

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/meta"

	"github.com/propeldata/terraform-provider-propel/version"
)

// GetUserAgent augments the default user agent with provider details
func GetUserAgent(clientUserAgent string) string {
	return fmt.Sprintf("terraform-provider-propel/%s (terraform %s) %s",
		version.ProviderVersion,
		meta.SDKVersionString(),
		clientUserAgent)
}
