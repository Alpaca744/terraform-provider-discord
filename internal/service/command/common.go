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

// optStr maps an empty API string to null and a non-empty one to a value.
func optStr(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}
