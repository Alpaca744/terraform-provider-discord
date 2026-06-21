package command

import (
	"context"
	"encoding/json"

	"github.com/alpaca744/terraform-provider-discord/internal/discord"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// expandDefaultPerms converts a set of permission names into the nullable
// decimal-string bitfield Discord stores in default_member_permissions. A null
// set yields a nil pointer (no restriction).
func expandDefaultPerms(ctx context.Context, set types.Set) (*string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if set.IsNull() || set.IsUnknown() {
		return nil, diags
	}
	var names []string
	diags.Append(set.ElementsAs(ctx, &names, false)...)
	if diags.HasError() {
		return nil, diags
	}
	bf, err := discord.PermissionsToBitfield(names)
	if err != nil {
		diags.AddError("Invalid default_member_permissions", err.Error())
		return nil, diags
	}
	return &bf, diags
}

// flattenDefaultPerms converts the API's nullable bitfield into a set of
// permission names, or a null set when no restriction is configured.
func flattenDefaultPerms(ctx context.Context, bitfield *string) (types.Set, diag.Diagnostics) {
	if bitfield == nil || *bitfield == "" {
		return types.SetNull(types.StringType), nil
	}
	names, err := discord.BitfieldToPermissions(*bitfield)
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("Invalid default_member_permissions from API", err.Error())
		return types.SetNull(types.StringType), diags
	}
	return types.SetValueFrom(ctx, types.StringType, names)
}

// expandOptions converts the normalized JSON options attribute into raw JSON for
// the request body. A null/empty value yields nil so the field is omitted.
func expandOptions(opts jsontypes.Normalized) json.RawMessage {
	if opts.IsNull() || opts.IsUnknown() || opts.ValueString() == "" {
		return nil
	}
	return json.RawMessage(opts.ValueString())
}

// flattenOptions converts raw JSON options from the API into the normalized
// attribute, preserving null when the command has no options.
func flattenOptions(raw json.RawMessage) jsontypes.Normalized {
	if len(raw) == 0 || string(raw) == "null" {
		return jsontypes.NewNormalizedNull()
	}
	return jsontypes.NewNormalizedValue(string(raw))
}

// optionsRaw returns the raw JSON backing an options value, or nil when it is
// null, unknown, or empty.
func optionsRaw(opts jsontypes.Normalized) json.RawMessage {
	if opts.IsNull() || opts.IsUnknown() || opts.ValueString() == "" {
		return nil
	}
	return json.RawMessage(opts.ValueString())
}

// optionsSemanticEqual reports whether two options values are equal ignoring
// formatting, object key order, and null-valued keys. Discord echoes options
// back with extra null keys (for example name_localizations and
// description_localizations) and reordered members, so a byte- or
// jsontypes-level comparison of the configured value against the API value
// reports a spurious difference. Canonicalizing both sides avoids that.
func optionsSemanticEqual(a, b jsontypes.Normalized) bool {
	return canonicalOptions(optionsRaw(a)) == canonicalOptions(optionsRaw(b))
}

// canonicalOptions normalizes options JSON for semantic comparison: it strips
// keys whose value is null and relies on encoding/json's sorted map-key output
// for stable ordering. Invalid JSON is returned unchanged so genuinely
// malformed input still compares unequal to a valid canonical form.
func canonicalOptions(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return string(raw)
	}
	b, err := json.Marshal(stripNullValues(v))
	if err != nil {
		return string(raw)
	}
	return string(b)
}

// stripNullValues recursively removes null-valued keys from JSON objects so that
// a missing key and an explicit null key compare as equal.
func stripNullValues(v any) any {
	switch t := v.(type) {
	case map[string]any:
		m := make(map[string]any, len(t))
		for k, val := range t {
			if val == nil {
				continue
			}
			m[k] = stripNullValues(val)
		}
		return m
	case []any:
		for i, e := range t {
			t[i] = stripNullValues(e)
		}
		return t
	default:
		return v
	}
}

// optStr maps an empty API string to null and a non-empty one to a value.
func optStr(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}
