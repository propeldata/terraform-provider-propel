package propel

import (
	"reflect"
	"testing"

	pc "github.com/propeldata/terraform-provider-propel/propel_client"
)

func Test_expandMetricFilters(t *testing.T) {
	tests := []struct {
		name          string
		def           []interface{}
		want          []*pc.FilterInput
		expectedError bool
	}{
		{
			name: "Basic filter",
			def: []any{
				map[string]any{"column": "foo", "operator": "EQUALS", "value": "2"},
			},
			want: []*pc.FilterInput{
				{Column: "foo", Operator: pc.FilterOperatorEquals, Value: "2"},
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
					Value:    "2",
					And: []*pc.FilterInput{
						{
							Column:   "bar",
							Operator: pc.FilterOperatorGreaterThan,
							Value:    "5",
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
					Value:    "2",
					Or: []*pc.FilterInput{
						{
							Column:   "bar",
							Operator: pc.FilterOperatorGreaterThan,
							Value:    "5",
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
					Value:    "2",
					And: []*pc.FilterInput{
						{
							Column:   "bar",
							Operator: pc.FilterOperatorGreaterThan,
							Value:    "5",
						},
					},
					Or: []*pc.FilterInput{
						{
							Column:   "baz",
							Operator: pc.FilterOperatorEquals,
							Value:    "abc",
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
					Value:    "2",
					And: []*pc.FilterInput{
						{
							Column:   "bar",
							Operator: pc.FilterOperatorGreaterThan,
							Value:    "5",
							And: []*pc.FilterInput{
								{
									Column:   "baz",
									Operator: pc.FilterOperatorEquals,
									Value:    "abc",
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
					Value:    "2",
					Or: []*pc.FilterInput{
						{
							Column:   "bar",
							Operator: pc.FilterOperatorGreaterThan,
							Value:    "5",
							And: []*pc.FilterInput{
								{
									Column:   "baz",
									Operator: pc.FilterOperatorEquals,
									Value:    "abc",
								},
							},
						},
					},
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
