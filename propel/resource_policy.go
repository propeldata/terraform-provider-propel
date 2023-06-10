package propel

import (
	"context"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	pc "github.com/propeldata/terraform-provider-propel/propel_client"
)

func resourcePolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePolicyCreate,
		UpdateContext: resourcePolicyUpdate,
		DeleteContext: resourcePolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Provides a Propel Policy resource. This can be used to create and manage Propel Policies.",
		Schema: map[string]*schema.Schema{
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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

func resourcePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(graphql.Client)

	var diags diag.Diagnostics

	policyType := d.Get("type").(string)

	input := &pc.CreatePolicyInput{
		Metric:      d.Get("metric").(string),
		Type:        pc.PolicyType(policyType),
		Application: d.Get("application").(string),
	}

	response, err := pc.CreatePolicy(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	policy := response.GetCreatePolicy().Policy
	d.SetId(policy.Id)

	if err := d.Set("type", policy.Type); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("application", policy.Application.Id); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("metric", policy.Metric.Id); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourcePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(graphql.Client)

	if d.HasChanges("type") {
		input := &pc.ModifyPolicyInput{
			Policy: d.Id(),
			Type:   pc.PolicyType(d.Get("type").(string)),
		}

		response, err := pc.ModifyPolicy(ctx, c, input)
		if err != nil {
			return diag.FromErr(err)
		}

		if err := d.Set("type", response.GetModifyPolicy().Policy.Type); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(graphql.Client)

	_, err := pc.DeletePolicy(ctx, c, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
