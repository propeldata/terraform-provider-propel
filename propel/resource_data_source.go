package propel

import (
	"context"
	"strconv"
	"time"

	hc "github.com/hashicorp-demoapp/hashicups-client-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDataSource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDataSourceCreate,
		ReadContext:   resourceDataSourceRead,
		UpdateContext: resourceDataSourceUpdate,
		DeleteContext: resourceDataSourceDelete,
		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Computed:    true,
				Description: "The DataSource ID",
			},
			"uniqueName": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    false,
				Computed:    true,
				Description: "The DataSource name",
			},
			"createdAt": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of when the DataSource was created",
			},
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceDataSourceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hc.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	items := d.Get("items").([]interface{})
	ois := []hc.OrderItem{}

	for _, item := range items {
		i := item.(map[string]interface{})

		co := i["coffee"].([]interface{})[0]
		coffee := co.(map[string]interface{})

		oi := hc.OrderItem{
			Coffee: hc.Coffee{
				ID: coffee["id"].(int),
			},
			Quantity: i["quantity"].(int),
		}

		ois = append(ois, oi)
	}

	o, err := c.CreateOrder(ois)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(o.ID))

	resourceDataSourceRead(ctx, d, m)

	return diags
}

func resourceDataSourceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hc.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	orderID := d.Id()

	order, err := c.GetOrder(orderID)
	if err != nil {
		return diag.FromErr(err)
	}

	orderItems := flattenOrderItems(&order.Items)
	if err := d.Set("items", orderItems); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceDataSourceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hc.Client)

	orderID := d.Id()

	if d.HasChange("items") {
		items := d.Get("items").([]interface{})
		ois := []hc.OrderItem{}

		for _, item := range items {
			i := item.(map[string]interface{})

			co := i["coffee"].([]interface{})[0]
			coffee := co.(map[string]interface{})

			oi := hc.OrderItem{
				Coffee: hc.Coffee{
					ID: coffee["id"].(int),
				},
				Quantity: i["quantity"].(int),
			}
			ois = append(ois, oi)
		}

		_, err := c.UpdateOrder(orderID, ois)
		if err != nil {
			return diag.FromErr(err)
		}

		d.Set("last_updated", time.Now().Format(time.RFC850))
	}

	return resourceDataSourceRead(ctx, d, m)
}

func resourceDataSourceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hc.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	orderID := d.Id()

	err := c.DeleteOrder(orderID)
	if err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}

func flattenOrderItems(orderItems *[]hc.OrderItem) []interface{} {
	if orderItems != nil {
		ois := make([]interface{}, len(*orderItems), len(*orderItems))

		for i, orderItem := range *orderItems {
			oi := make(map[string]interface{})

			oi["coffee"] = flattenCoffee(orderItem.Coffee)
			oi["quantity"] = orderItem.Quantity
			ois[i] = oi
		}

		return ois
	}

	return make([]interface{}, 0)
}

func flattenCoffee(coffee hc.Coffee) []interface{} {
	c := make(map[string]interface{})
	c["id"] = coffee.ID
	c["name"] = coffee.Name
	c["teaser"] = coffee.Teaser
	c["description"] = coffee.Description
	c["price"] = coffee.Price
	c["image"] = coffee.Image

	return []interface{}{c}
}
