package guild

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestRoleBuildBody(t *testing.T) {
	r := &roleResource{}
	ctx := context.Background()

	perms, d := types.SetValueFrom(ctx, types.StringType, []string{"VIEW_CHANNEL", "SEND_MESSAGES"})
	if d.HasError() {
		t.Fatalf("set build: %v", d)
	}

	plan := roleResourceModel{
		Name:        types.StringValue("mods"),
		Color:       types.Int64Value(16711680),
		Hoist:       types.BoolValue(true),
		Mentionable: types.BoolValue(false),
		Permissions: perms,
		// Position and Managed are computed-only and must not appear in the body.
	}

	body, diags := r.buildBody(ctx, plan)
	if diags.HasError() {
		t.Fatalf("buildBody: %v", diags)
	}
	if body.Name == nil || *body.Name != "mods" {
		t.Errorf("name = %v", body.Name)
	}
	if body.Color == nil || *body.Color != 16711680 {
		t.Errorf("color = %v", body.Color)
	}
	if body.Hoist == nil || *body.Hoist != true {
		t.Errorf("hoist = %v", body.Hoist)
	}
	// VIEW_CHANNEL(1024) | SEND_MESSAGES(2048) = 3072
	if body.Permissions == nil || *body.Permissions != "3072" {
		t.Errorf("permissions = %v", body.Permissions)
	}
}

func TestRoleBuildBodyOmitsNull(t *testing.T) {
	r := &roleResource{}
	plan := roleResourceModel{
		Name:        types.StringNull(),
		Color:       types.Int64Null(),
		Hoist:       types.BoolNull(),
		Mentionable: types.BoolNull(),
		Permissions: types.SetNull(types.StringType),
	}
	body, diags := r.buildBody(context.Background(), plan)
	if diags.HasError() {
		t.Fatalf("buildBody: %v", diags)
	}
	if body.Name != nil || body.Color != nil || body.Hoist != nil || body.Permissions != nil {
		t.Errorf("null plan should produce empty body, got %+v", body)
	}
}

func TestRoleApply(t *testing.T) {
	r := &roleResource{}
	ctx := context.Background()
	var m roleResourceModel

	role := &Role{
		ID:          "555",
		Name:        "mods",
		Color:       42,
		Hoist:       true,
		Position:    3,
		Managed:     false,
		Mentionable: true,
		Permissions: "3072",
	}
	if d := r.apply(ctx, &m, role); d.HasError() {
		t.Fatalf("apply: %v", d)
	}
	if m.ID.ValueString() != "555" || m.Name.ValueString() != "mods" {
		t.Errorf("id/name = %s/%s", m.ID.ValueString(), m.Name.ValueString())
	}
	if m.Position.ValueInt64() != 3 {
		t.Errorf("position = %d", m.Position.ValueInt64())
	}
	var names []string
	if d := m.Permissions.ElementsAs(ctx, &names, false); d.HasError() {
		t.Fatalf("perms read: %v", d)
	}
	if len(names) != 2 {
		t.Errorf("permissions = %v", names)
	}
}
