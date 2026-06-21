package application

import "github.com/hashicorp/terraform-plugin-framework/types"

// strPtr returns a pointer to the string value, or nil when null/unknown, so
// optional attributes are omitted from PATCH bodies rather than zeroed.
func strPtr(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	return &s
}

// optStr maps an empty API string to null and a non-empty one to a value.
func optStr(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}
