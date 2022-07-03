//go:build tools
// +build tools

package tools

import (
	_ "github.com/Khan/genqlient"
	// document generation
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)
