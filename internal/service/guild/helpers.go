package guild

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

func int64Ptr(v types.Int64) *int64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	i := v.ValueInt64()
	return &i
}

func boolPtr(v types.Bool) *bool {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	b := v.ValueBool()
	return &b
}

// optStr maps an empty API string to null and a non-empty one to a value, so
// unset optional attributes do not show perpetual drift.
func optStr(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}
