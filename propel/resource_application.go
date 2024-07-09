package propel

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	pc "github.com/propeldata/terraform-provider-propel/propel_client"
)

func resourceApplication() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApplicationCreate,
		ReadContext:   resourceApplicationRead,
		UpdateContext: resourceApplicationUpdate,
		DeleteContext: resourceApplicationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Description:   "Provides a Propel Application resource.",
		Schema: map[string]*schema.Schema{
			"unique_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The Application's name.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The Application's description.",
			},
			"account": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Account that the Application belongs to.",
			},
			"environment": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Environment that the Application belongs to.",
			},
			"client_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Application's OAuth 2.0 client identifier.",
			},
			"secret": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The Application's OAuth 2.0 client secret.",
			},
			"propeller": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Application's Propeller. If no Propeller is provided, Propel will set the Propeller to `P1_X_SMALL`. The valid values are `P1_X_SMALL`, `P1_SMALL`, `P1_MEDIUM`, `P1_LARGE` and `P1_X_LARGE`",
				ValidateFunc: validation.StringInSlice([]string{
					"P1_X_SMALL",
					"P1_SMALL",
					"P1_MEDIUM",
					"P1_LARGE",
					"P1_X_LARGE",
				}, false),
			},
			"scopes": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The Application's API authorization scopes. If specified, at least one scope must be provided; otherwise, all scopes will be granted to the Application by default.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceApplicationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(graphql.Client)

	input := &pc.CreateApplicationInput{
		Scopes: make([]pc.ApplicationScope, 0),
	}

	if v, exists := d.GetOk("unique_name"); exists && v.(string) != "" {
		uniqueName := v.(string)
		input.UniqueName = &uniqueName
	}

	if v, exists := d.GetOk("description"); exists && v.(string) != "" {
		description := v.(string)
		input.Description = &description
	}

	if v, exists := d.GetOk("propeller"); exists && v.(string) != "" {
		propeller := parsePropeller(v.(string))
		input.Propeller = &propeller
	}

	if def, ok := d.GetOk("scopes"); ok {
		for _, v := range def.(*schema.Set).List() {
			scope := parseApplicationScope(v.(string))
			input.Scopes = append(input.Scopes, scope)
		}
	}

	response, err := pc.CreateApplication(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	switch r := (*response.GetCreateApplication()).(type) {
	case *pc.CreateApplicationCreateApplicationApplicationResponse:
		d.SetId(r.Application.Id)
	case *pc.CreateApplicationCreateApplicationFailureResponse:
		return diag.FromErr(fmt.Errorf("failed to create Application: %s", r.GetError().GetMessage()))
	}

	return nil
}

func resourceApplicationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(graphql.Client)

	response, err := pc.Application(ctx, c, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(response.Application.Id)

	if err := d.Set("unique_name", response.Application.UniqueName); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("description", response.Application.Description); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("environment", response.Application.Environment.Id); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("account", response.Application.Account.Id); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("client_id", response.Application.ClientId); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("secret", response.Application.Secret); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("propeller", response.Application.Propeller); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("scopes", response.Application.Scopes); err != nil {

		return diag.FromErr(err)
	}

	return nil
}

func resourceApplicationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(graphql.Client)

	id := d.Id()

	input := &pc.ModifyApplicationInput{
		IdOrUniqueName: &pc.IdOrUniqueName{Id: &id},
	}

	if d.HasChanges("unique_name", "description", "scopes", "propeller") {
		uniqueName := d.Get("unique_name").(string)
		input.UniqueName = &uniqueName

		description := d.Get("description").(string)
		input.Description = &description

		scopes := make([]pc.ApplicationScope, 0)
		for _, v := range d.Get("scopes").(*schema.Set).List() {
			scope := parseApplicationScope(v.(string))
			scopes = append(scopes, scope)
		}

		input.Scopes = scopes

		propeller := parsePropeller(d.Get("propeller").(string))
		input.Propeller = &propeller
	}

	if _, err := pc.ModifyApplication(ctx, c, input); err != nil {
		return diag.FromErr(err)
	}

	return resourceApplicationRead(ctx, d, meta)
}

func resourceApplicationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(graphql.Client)

	if _, err := pc.DeleteApplication(ctx, c, d.Id()); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func parsePropeller(text string) pc.Propeller {
	var propeller pc.Propeller

	switch text {
	case "P1_X_SMALL":
		propeller = pc.PropellerP1XSmall
	case "P1_SMALL":
		propeller = pc.PropellerP1Small
	case "P1_MEDIUM":
		propeller = pc.PropellerP1Medium
	case "P1_LARGE":
		propeller = pc.PropellerP1Large
	case "P1_X_LARGE":
		propeller = pc.PropellerP1XLarge
	}

	return propeller
}

func parseApplicationScope(text string) pc.ApplicationScope {
	var scope pc.ApplicationScope

	switch text {
	case "ADMIN":
		scope = pc.ApplicationScopeAdmin
	case "APPLICATION_ADMIN":
		scope = pc.ApplicationScopeApplicationAdmin
	case "DATA_POOL_QUERY":
		scope = pc.ApplicationScopeDataPoolQuery
	case "DATA_POOL_READ":
		scope = pc.ApplicationScopeDataPoolRead
	case "DATA_POOL_STATS":
		scope = pc.ApplicationScopeDataPoolStats
	case "METRIC_QUERY":
		scope = pc.ApplicationScopeMetricQuery
	case "METRIC_READ":
		scope = pc.ApplicationScopeMetricRead
	case "METRIC_STATS":
		scope = pc.ApplicationScopeMetricStats
	}

	return scope
}
