//go:build tools

// Package tools pins the code-generation tooling used by the provider so it is
// tracked in go.mod. It is never compiled into the provider binary.
package tools

import (
	// Generates Terraform Registry documentation from schemas, templates, and examples.
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)
