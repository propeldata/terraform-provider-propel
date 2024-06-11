package propel

import (
	"context"
	"errors"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	pc "github.com/propeldata/terraform-provider-propel/propel_client"
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

func resourcePolicyRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(graphql.Client)

	response, err := pc.Policy(ctx, c, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(response.Policy.Id)

	if err := d.Set("type", response.Policy.Type); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("application", response.Policy.Application.Id); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("metric", response.Policy.Metric.Id); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourcePolicyUpdate(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return diag.FromErr(errors.New("use propel_data_pool_access_policy resource instead"))
}

func resourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(graphql.Client)

	_, err := pc.DeletePolicy(ctx, c, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
