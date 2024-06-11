package propel

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourcePolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePolicyCreate,
		ReadContext:   resourcePolicyRead,
		UpdateContext: resourcePolicyUpdate,
		DeleteContext: resourcePolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion:      1,
		Description:        "Provides a Propel Policy resource. This can be used to create and manage Propel Access Policies. It governs an Application's access to a Metric's data.",
		DeprecationMessage: "Use Data Pool Access Policy instead",
		Schema: map[string]*schema.Schema{
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"ALL_ACCESS",
					"TENANT_ACCESS",
				}, false),
				Description: "The Policy type. The different Policy types determine the access to the Metric data.",
			},
			"application": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Application that is granted access.",
			},
			"metric": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Metric that the Application is granted access to.",
			},
		},
	}
}

func resourcePolicyCreate(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return diag.FromErr(errors.New("use propel_data_pool_access_policy resource instead"))
}

func resourcePolicyRead(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return diag.FromErr(errors.New("use propel_data_pool_access_policy resource instead"))
}

func resourcePolicyUpdate(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return diag.FromErr(errors.New("use propel_data_pool_access_policy resource instead"))
}

func resourcePolicyDelete(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return diag.FromErr(errors.New("use propel_data_pool_access_policy resource instead"))
}
