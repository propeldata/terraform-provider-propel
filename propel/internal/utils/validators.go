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
	"STRING",
	"TIMESTAMP",
}, false)
