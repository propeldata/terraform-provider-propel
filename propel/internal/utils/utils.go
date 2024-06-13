package utils

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/meta"

	pc "github.com/propeldata/terraform-provider-propel/propel_client"
	"github.com/propeldata/terraform-provider-propel/version"
)

var (
	dataSourceTypeMap = map[pc.DataSourceType]string{
		pc.DataSourceTypeSnowflake:  "Snowflake",
		pc.DataSourceTypeS3:         "S3",
		pc.DataSourceTypeHttp:       "Http",
		pc.DataSourceTypeWebhook:    "Webhook",
		pc.DataSourceTypeKafka:      "Kafka",
		pc.DataSourceTypeClickhouse: "ClickHouse",
	}
)

// GetUserAgent augments the default user agent with provider details
func GetUserAgent(clientUserAgent string) string {
	return fmt.Sprintf("terraform-provider-propel/%s (terraform %s) %s",
		version.ProviderVersion,
		meta.SDKVersionString(),
		clientUserAgent)
}

func GetDataSourceType(dataSourceType pc.DataSourceType) string {
	if dsType, ok := dataSourceTypeMap[dataSourceType]; ok {
		return dsType
	}

	return ""
}
