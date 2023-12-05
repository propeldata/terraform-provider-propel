package propel

import (
	"context"
	"fmt"
	"runtime"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/propeldata/terraform-provider-propel/propel/internal/utils"
	pc "github.com/propeldata/terraform-provider-propel/propel_client"
)

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   false,
				DefaultFunc: schema.EnvDefaultFunc("PROPEL_CLIENT_ID", nil),
				Description: "Your Propel Application's ID.",
			},
			"client_secret": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("PROPEL_CLIENT_SECRET", nil),
				Description: "Your Propel Application's secret.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"propel_data_source":             resourceDataSource(),
			"propel_data_pool":               resourceDataPool(),
			"propel_data_pool_access_policy": resourceDataPoolAccessPolicy(),
			"propel_metric":                  resourceMetric(),
			"propel_policy":                  resourcePolicy(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
	clientID := d.Get("client_id").(string)
	clientSecret := d.Get("client_secret").(string)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if (clientID == "") || (clientSecret == "") {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Credentials are required",
			Detail:   "Unable to authenticate for the Propel client",
		})

		return nil, diags
	}

	userAgent := utils.GetUserAgent(fmt.Sprintf(
		"propel-client-go (go %s; os %s; arch %s)",
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	))

	c, err := pc.NewPropelClient(clientID, clientSecret, userAgent)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return c, nil
}
