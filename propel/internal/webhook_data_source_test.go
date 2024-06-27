package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"

	pc "github.com/propeldata/terraform-provider-propel/propel_client"
)

func Test_getNewDataSourceColumns(t *testing.T) {
	tests := []struct {
		name               string
		oldItemDef         []any
		newItemDef         []any
		expectedNewColumns map[string]pc.WebhookDataSourceColumnInput
		expectedError      string
	}{
		{
			name: "Successful new columns",
			oldItemDef: []any{
				map[string]any{"name": "COLUMN_A", "type": "STRING", "nullable": false, "json_property": "column_a"},
				map[string]any{"name": "COLUMN_B", "type": "FLOAT", "nullable": false, "json_property": "column_b"},
			},
			newItemDef: []any{
				map[string]any{"name": "COLUMN_A", "type": "STRING", "nullable": false, "json_property": "column_a"},
				map[string]any{"name": "COLUMN_B", "type": "FLOAT", "nullable": false, "json_property": "column_b"},
				map[string]any{"name": "COLUMN_C", "type": "INT64", "nullable": false, "json_property": "column_c"},
				map[string]any{"name": "COLUMN_D", "type": "TIMESTAMP", "nullable": false, "json_property": "column_d"},
			},
			expectedNewColumns: map[string]pc.WebhookDataSourceColumnInput{
				"COLUMN_C": {Name: "COLUMN_C", Type: "INT64", Nullable: false, JsonProperty: "column_c"},
				"COLUMN_D": {Name: "COLUMN_D", Type: "TIMESTAMP", Nullable: false, JsonProperty: "column_d"},
			},
			expectedError: "",
		},
		{
			name: "No new columns",
			oldItemDef: []any{
				map[string]any{"name": "COLUMN_A", "type": "STRING", "nullable": false, "json_property": "column_a"},
				map[string]any{"name": "COLUMN_B", "type": "FLOAT", "nullable": false, "json_property": "column_b"},
			},
			newItemDef: []any{
				map[string]any{"name": "COLUMN_A", "type": "STRING", "nullable": false, "json_property": "column_a"},
				map[string]any{"name": "COLUMN_B", "type": "FLOAT", "nullable": false, "json_property": "column_b"},
			},
			expectedNewColumns: map[string]pc.WebhookDataSourceColumnInput{},
			expectedError:      "",
		},
		{
			name: "Repeated column names",
			oldItemDef: []any{
				map[string]any{"name": "COLUMN_A", "type": "STRING", "nullable": false, "json_property": "column_a"},
			},
			newItemDef: []any{
				map[string]any{"name": "COLUMN_B", "type": "FLOAT", "nullable": false, "json_property": "column_b"},
				map[string]any{"name": "COLUMN_B", "type": "TIMESTAMP", "nullable": false, "json_property": "column_sb"},
			},
			expectedNewColumns: map[string]pc.WebhookDataSourceColumnInput{},
			expectedError:      `column "COLUMN_B" already exists`,
		},
		{
			name: "Unsupported column deletion",
			oldItemDef: []any{
				map[string]any{"name": "COLUMN_A", "type": "STRING", "nullable": false, "json_property": "column_a"},
				map[string]any{"name": "COLUMN_B", "type": "FLOAT", "nullable": false, "json_property": "column_b"},
			},
			newItemDef: []any{
				map[string]any{"name": "COLUMN_A", "type": "STRING", "nullable": false, "json_property": "column_a"},
			},
			expectedNewColumns: map[string]pc.WebhookDataSourceColumnInput{},
			expectedError:      `column "COLUMN_B" was removed, column deletions are not supported`,
		},
		{
			name: "Unsupported column update",
			oldItemDef: []any{
				map[string]any{"name": "COLUMN_A", "type": "STRING", "nullable": false, "json_property": "column_a"},
				map[string]any{"name": "COLUMN_B", "type": "FLOAT", "nullable": false, "json_property": "column_b"},
			},
			newItemDef: []any{
				map[string]any{"name": "COLUMN_A", "type": "STRING", "nullable": false, "json_property": "column_z"},
			},
			expectedNewColumns: map[string]pc.WebhookDataSourceColumnInput{},
			expectedError:      `column "COLUMN_A" was modified, column updates are not supported`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(st *testing.T) {
			a := assert.New(st)

			result, err := newWebhookColumns(tt.oldItemDef, tt.newItemDef)
			if tt.expectedError != "" {
				a.Error(err)
				a.EqualError(err, tt.expectedError)
				return
			}

			a.NoError(err)
			a.Equal(tt.expectedNewColumns, result)
		})
	}
}
