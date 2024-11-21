package utils

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var IsValidColumnType = validation.StringInSlice([]string{
	"BOOLEAN",
	"DATE",
	"DOUBLE",
	"FLOAT",
	"INT8",
	"INT16",
	"INT32",
	"INT64",
	"JSON",
	"STRING",
	"TIMESTAMP",
	"CLICKHOUSE",
}, false)

var IsValidOperator = validation.StringInSlice([]string{
	"EQUALS",
	"NOT_EQUALS",
	"GREATER_THAN",
	"GREATER_THAN_OR_EQUAL_TO",
	"LESS_THAN",
	"LESS_THAN_OR_EQUAL_TO",
	"IS_NULL",
	"IS_NOT_NULL",
}, false)
