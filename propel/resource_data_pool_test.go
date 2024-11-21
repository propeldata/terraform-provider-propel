package propel

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"

	pc "github.com/propeldata/terraform-provider-propel/propel_client"
)

func TestAccPropelDataPoolBasic(t *testing.T) {
	ctx := map[string]any{
		"unique_name": acctest.RandString(12),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckPropelDataPoolDestroy,
		Steps: []resource.TestStep{
			// should create the Data Pool
			{
				Config: testAccCheckPropelDataPoolConfigBasic(ctx),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPropelDataPoolExists("propel_data_pool.bar"),
					resource.TestCheckResourceAttr("propel_data_pool.bar", "unique_name", ctx["unique_name"].(string)),
					resource.TestCheckResourceAttrSet("propel_data_pool.bar", "table"),
					resource.TestCheckResourceAttr("propel_data_pool.bar", "tenant_id", "account_id"),
					resource.TestCheckResourceAttr("propel_data_pool.bar", "description", "Data Pool test"),
				),
			},
			// should update the Data Pool
			{
				Config: testAccUpdatePropelDataPoolConfigBasic(ctx),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPropelDataPoolExists("propel_data_pool.bar"),
					resource.TestCheckResourceAttrSet("propel_data_pool.bar", "table"),
					resource.TestCheckResourceAttr("propel_data_pool.bar", "tenant_id", "account_id"),
					resource.TestCheckResourceAttr("propel_data_pool.bar", "description", "Updated description"),
				),
			},
		},
	})
}

func testAccCheckPropelDataPoolConfigBasic(ctx map[string]any) string {
	// language=hcl-terraform
	return Nprintf(`

	resource "propel_data_pool" "bar" {
		unique_name = "%{unique_name}"
		description = "Data Pool test"

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
		timestamp = "timestamp_tz"
	}`, ctx)
}

func testAccUpdatePropelDataPoolConfigBasic(ctx map[string]any) string {
	// language=hcl-terraform
	return Nprintf(`

	resource "propel_data_pool" "bar" {
		unique_name = "%{unique_name}"
		description = "Updated description"

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
		column {
			name = "product_id"
			type = "CLICKHOUSE"
			clickhouse_type = "String"
			nullable = true
		}
		tenant_id = "account_id"
		timestamp = "timestamp_tz"
	}`, ctx)
}

func testAccCheckPropelDataPoolDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(graphql.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "propel_data_pool" {
			continue
		}

		dataPoolID := rs.Primary.ID

		_, err := pc.DeleteDataPool(context.Background(), c, dataPoolID)
		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckPropelDataPoolExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no Data Pool ID set")
		}

		return nil
	}
}

func Test_getNewDataPoolColumns(t *testing.T) {
	clickHouseType := "String"

	tests := []struct {
		name               string
		oldItemDef         []any
		newItemDef         []any
		expectedNewColumns map[string]pc.DataPoolColumnInput
		expectedError      string
	}{
		{
			name: "Successful new columns",
			oldItemDef: []any{
				map[string]any{"name": "COLUMN_A", "type": "STRING", "nullable": false, "clickhouse_type": ""},
				map[string]any{"name": "COLUMN_B", "type": "FLOAT", "nullable": false, "clickhouse_type": ""},
			},
			newItemDef: []any{
				map[string]any{"name": "COLUMN_A", "type": "STRING", "nullable": false, "clickhouse_type": ""},
				map[string]any{"name": "COLUMN_B", "type": "FLOAT", "nullable": false, "clickhouse_type": ""},
				map[string]any{"name": "COLUMN_C", "type": "INT64", "nullable": false, "clickhouse_type": ""},
				map[string]any{"name": "COLUMN_D", "type": "TIMESTAMP", "nullable": false, "clickhouse_type": ""},
				map[string]any{"name": "COLUMN_E", "type": "CLICKHOUSE", "nullable": true, "clickhouse_type": clickHouseType},
			},
			expectedNewColumns: map[string]pc.DataPoolColumnInput{
				"COLUMN_C": {ColumnName: "COLUMN_C", Type: "INT64", IsNullable: false, ClickHouseType: nil},
				"COLUMN_D": {ColumnName: "COLUMN_D", Type: "TIMESTAMP", IsNullable: false, ClickHouseType: nil},
				"COLUMN_E": {ColumnName: "COLUMN_E", Type: "CLICKHOUSE", IsNullable: true, ClickHouseType: &clickHouseType},
			},
			expectedError: "",
		},
		{
			name: "No new columns",
			oldItemDef: []any{
				map[string]any{"name": "COLUMN_A", "type": "STRING", "nullable": false, "clickhouse_type": ""},
				map[string]any{"name": "COLUMN_B", "type": "FLOAT", "nullable": false, "clickhouse_type": ""},
			},
			newItemDef: []any{
				map[string]any{"name": "COLUMN_A", "type": "STRING", "nullable": false, "clickhouse_type": ""},
				map[string]any{"name": "COLUMN_B", "type": "FLOAT", "nullable": false, "clickhouse_type": ""},
			},
			expectedNewColumns: map[string]pc.DataPoolColumnInput{},
			expectedError:      "",
		},
		{
			name: "Repeated column names",
			oldItemDef: []any{
				map[string]any{"name": "COLUMN_A", "type": "STRING", "nullable": false, "clickhouse_type": ""},
			},
			newItemDef: []any{
				map[string]any{"name": "COLUMN_B", "type": "FLOAT", "nullable": false, "clickhouse_type": ""},
				map[string]any{"name": "COLUMN_B", "type": "INT64", "nullable": false, "clickhouse_type": ""},
			},
			expectedNewColumns: map[string]pc.DataPoolColumnInput{},
			expectedError:      `column "COLUMN_B" already exists`,
		},
		{
			name: "Unsupported column deletion",
			oldItemDef: []any{
				map[string]any{"name": "COLUMN_A", "type": "STRING", "nullable": false, "clickhouse_type": ""},
				map[string]any{"name": "COLUMN_B", "type": "INT64", "nullable": false, "clickhouse_type": ""},
			},
			newItemDef: []any{
				map[string]any{"name": "COLUMN_A", "type": "STRING", "nullable": false, "clickhouse_type": ""},
			},
			expectedNewColumns: map[string]pc.DataPoolColumnInput{},
			expectedError:      `column "COLUMN_B" was removed, column deletions are not supported`,
		},
		{
			name: "Unsupported column update",
			oldItemDef: []any{
				map[string]any{"name": "COLUMN_A", "type": "STRING", "nullable": false, "clickhouse_type": ""},
				map[string]any{"name": "COLUMN_B", "type": "INT64", "nullable": false, "clickhouse_type": ""},
			},
			newItemDef: []any{
				map[string]any{"name": "COLUMN_A", "type": "FLOAT", "nullable": false, "clickhouse_type": ""},
			},
			expectedNewColumns: map[string]pc.DataPoolColumnInput{},
			expectedError:      `column "COLUMN_A" was modified, column updates are not supported`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(st *testing.T) {
			a := assert.New(st)

			result, err := getNewDataPoolColumns(tt.oldItemDef, tt.newItemDef)
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
