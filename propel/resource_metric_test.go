package propel

import (
	"context"
	"reflect"
	"testing"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	pc "github.com/propeldata/terraform-provider-propel/propel_client"
)

var (
	two  = "2"
	five = "5"
	abc  = "abc"
)

func Test_expandMetricFilters(t *testing.T) {
	tests := []struct {
		name          string
		def           []any
		want          []*pc.FilterInput
		expectedError bool
	}{
		{
			name: "Basic filter",
			def: []any{
				map[string]any{"column": "foo", "operator": "EQUALS", "value": "2"},
			},
			want: []*pc.FilterInput{
				{Column: "foo", Operator: pc.FilterOperatorEquals, Value: &two},
			},
		},
		{
			name: "With AND and OR as empty strings",
			def: []any{
				map[string]any{"column": "foo", "operator": "EQUALS", "value": "2", "and": "", "or": ""},
			},
			want: []*pc.FilterInput{
				{Column: "foo", Operator: pc.FilterOperatorEquals, Value: &two},
			},
		},
		{
			name: "With one AND filter",
			def: []any{
				map[string]any{"column": "foo", "operator": "EQUALS", "value": "2", "and": `[{"column": "bar", "operator": "GREATER_THAN", "value": "5"}]`},
			},
			want: []*pc.FilterInput{
				{
					Column:   "foo",
					Operator: pc.FilterOperatorEquals,
					Value:    &two,
					And: []*pc.FilterInput{
						{
							Column:   "bar",
							Operator: pc.FilterOperatorGreaterThan,
							Value:    &five,
						},
					},
				},
			},
		},
		{
			name: "With one OR filter",
			def: []any{
				map[string]any{"column": "foo", "operator": "EQUALS", "value": "2", "or": `[{"column": "bar", "operator": "GREATER_THAN", "value": "5"}]`},
			},
			want: []*pc.FilterInput{
				{
					Column:   "foo",
					Operator: pc.FilterOperatorEquals,
					Value:    &two,
					Or: []*pc.FilterInput{
						{
							Column:   "bar",
							Operator: pc.FilterOperatorGreaterThan,
							Value:    &five,
						},
					},
				},
			},
		},
		{
			name: "With one AND and OR filter combined",
			def: []any{
				map[string]any{"column": "foo", "operator": "EQUALS", "value": "2", "and": `[{"column": "bar", "operator": "GREATER_THAN", "value": "5"}]`, "or": `[{"column": "baz", "operator": "EQUALS", "value": "abc"}]`},
			},
			want: []*pc.FilterInput{
				{
					Column:   "foo",
					Operator: pc.FilterOperatorEquals,
					Value:    &two,
					And: []*pc.FilterInput{
						{
							Column:   "bar",
							Operator: pc.FilterOperatorGreaterThan,
							Value:    &five,
						},
					},
					Or: []*pc.FilterInput{
						{
							Column:   "baz",
							Operator: pc.FilterOperatorEquals,
							Value:    &abc,
						},
					},
				},
			},
		},
		{
			name: "With AND nested filter",
			def: []any{
				map[string]any{"column": "foo", "operator": "EQUALS", "value": "2", "and": `[{"column": "bar", "operator": "GREATER_THAN", "value": "5", "and": [{"column": "baz", "operator": "EQUALS", "value": "abc"}]}]`},
			},
			want: []*pc.FilterInput{
				{
					Column:   "foo",
					Operator: pc.FilterOperatorEquals,
					Value:    &two,
					And: []*pc.FilterInput{
						{
							Column:   "bar",
							Operator: pc.FilterOperatorGreaterThan,
							Value:    &five,
							And: []*pc.FilterInput{
								{
									Column:   "baz",
									Operator: pc.FilterOperatorEquals,
									Value:    &abc,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "With OR nested filter",
			def: []any{
				map[string]any{"column": "foo", "operator": "EQUALS", "value": "2", "or": `[{"column": "bar", "operator": "GREATER_THAN", "value": "5", "and": [{"column": "baz", "operator": "EQUALS", "value": "abc"}]}]`},
			},
			want: []*pc.FilterInput{
				{
					Column:   "foo",
					Operator: pc.FilterOperatorEquals,
					Value:    &two,
					Or: []*pc.FilterInput{
						{
							Column:   "bar",
							Operator: pc.FilterOperatorGreaterThan,
							Value:    &five,
							And: []*pc.FilterInput{
								{
									Column:   "baz",
									Operator: pc.FilterOperatorEquals,
									Value:    &abc,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "With IS_NULL filter",
			def: []any{
				map[string]any{"column": "foo", "operator": "IS_NULL"},
			},
			want: []*pc.FilterInput{
				{
					Column:   "foo",
					Operator: pc.FilterOperatorIsNull,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, diags := expandMetricFilters(tt.def)
			if diags.HasError() != tt.expectedError {
				t.Errorf("expandMetricFilters() to return an error, got %v", diags)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expandMetricFilters() got = %v, want %v", got, tt.want)
			}

		})
	}
}

func TestAccPropelMetricBasic(t *testing.T) {
	ctx := map[string]any{}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckPropelMetricDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPropelMetricConfigBasic(ctx),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPropelDataPoolExists("propel_metric.baz"),
					resource.TestCheckResourceAttr("propel_metric.baz", "type", "CUSTOM"),
					resource.TestCheckResourceAttr("propel_metric.baz", "expression", "COUNT_DISTINCT(account_id) / COUNT()"),
				),
			},
		},
	})
}

func testAccCheckPropelMetricConfigBasic(ctx map[string]any) string {
	// language=hcl-terraform
	return Nprintf(`
		resource "propel_data_source" "foo" {
		unique_name = "terraform-test-4"
		type = "HTTP"

		http_connection_settings {
			basic_auth {
				username = "foo"
				password = "bar"
			}
		}

		table {
			name = "CLUSTER_TEST_TABLE_1"

			column {
				name = "timestamp_tz"
				type = "TIMESTAMP"
				nullable = false
			}

			column {
				name = "account_id"
				type = "STRING"
				nullable = false
			}
		}
	}

	resource "propel_data_pool" "bar" {
		unique_name = "terraform-test-4"
		table = "${propel_data_source.foo.table[0].name}"

		column {
			name = "timestamp_tz"
			type = "TIMESTAMP"
			nullable = false
		}
		column {
			name = "account_id"
			type = "STRING"
			nullable = false
		}
		tenant_id = "account_id"
		timestamp = "${propel_data_source.foo.table[0].column[0].name}"
		data_source = "${propel_data_source.foo.id}"
	}
	
	resource "propel_metric" "baz" {
		unique_name = "terraform-test-4"
		description = "This is an example of a Custom Metric"
		data_pool   = propel_data_pool.bar.id

		type         = "CUSTOM"
		expression   = "COUNT_DISTINCT(account_id) / COUNT()"

		filter {
		    column   = "account_id"
			operator = "IS_NOT_NULL"
		}
	}
	`, ctx)
}

func testAccCheckPropelMetricDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(graphql.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "propel_metric" {
			continue
		}

		metricID := rs.Primary.ID

		_, err := pc.DeleteMetric(context.Background(), c, metricID)
		if err != nil {
			return err
		}
	}

	return nil
}
