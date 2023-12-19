package utils

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/meta"

	client "github.com/propeldata/terraform-provider-propel/propel_client"
	"github.com/propeldata/terraform-provider-propel/version"
)

var (
	dataSourceTypeMap = map[client.DataSourceType]string{
		client.DataSourceTypeSnowflake: "Snowflake",
		client.DataSourceTypeS3:        "S3",
		client.DataSourceTypeHttp:      "Http",
		client.DataSourceTypeWebhook:   "Webhook",
	}
)

// GetUserAgent augments the default user agent with provider details
func GetUserAgent(clientUserAgent string) string {
	return fmt.Sprintf("terraform-provider-propel/%s (terraform %s) %s",
		version.ProviderVersion,
		meta.SDKVersionString(),
		clientUserAgent)
}

func GetDataSourceType(dataSourceType client.DataSourceType) string {
	if dsType, ok := dataSourceTypeMap[dataSourceType]; ok {
		return dsType
	}

	return ""
}
