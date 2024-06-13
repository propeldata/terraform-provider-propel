package internal

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	pc "github.com/propeldata/terraform-provider-propel/propel_client"
)

func basicAuthSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Description: "The HTTP basic authentication settings. If this parameter is not provided, anyone with the URL will be able to send events. While it's OK to test without HTTP Basic authentication, we recommend enabling it.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"username": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Username for HTTP Basic authentication that must be included in the Authorization header when uploading new data.",
				},
				"password": {
					Type:        schema.TypeString,
					Required:    true,
					Sensitive:   true,
					Description: "Password for HTTP Basic authentication that must be included in the Authorization header when uploading new data.",
				},
			},
		},
	}
}

func expandBasicAuth(def []any) *pc.HttpBasicAuthInput {
	basicAuth := def[0].(map[string]any)

	return &pc.HttpBasicAuthInput{
		Username: basicAuth["username"].(string),
		Password: basicAuth["password"].(string),
	}
}
