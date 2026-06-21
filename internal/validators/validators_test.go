package validators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func runString(v validator.String, value types.String) *validator.StringResponse {
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), validator.StringRequest{
		Path:        path.Root("test"),
		ConfigValue: value,
	}, resp)
	return resp
}

func TestSnowflakeValidator(t *testing.T) {
	if runString(Snowflake(), types.StringValue("123456789012345678")).Diagnostics.HasError() {
		t.Error("valid snowflake rejected")
	}
	if !runString(Snowflake(), types.StringValue("not-a-snowflake")).Diagnostics.HasError() {
		t.Error("invalid snowflake accepted")
	}
	// Null/unknown must not error (handled by Required/Optional, not the validator).
	if runString(Snowflake(), types.StringNull()).Diagnostics.HasError() {
		t.Error("null should be skipped")
	}
}

func TestPermissionNameValidator(t *testing.T) {
	if runString(PermissionName(), types.StringValue("VIEW_CHANNEL")).Diagnostics.HasError() {
		t.Error("valid permission rejected")
	}
	if !runString(PermissionName(), types.StringValue("FLY")).Diagnostics.HasError() {
		t.Error("invalid permission accepted")
	}
}
