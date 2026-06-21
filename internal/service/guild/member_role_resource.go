package guild

import (
	"context"
	"fmt"
	"strings"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
	"github.com/alpaca744/terraform-provider-discord/internal/diagutil"
	"github.com/alpaca744/terraform-provider-discord/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*memberRoleResource)(nil)
	_ resource.ResourceWithConfigure   = (*memberRoleResource)(nil)
	_ resource.ResourceWithImportState = (*memberRoleResource)(nil)
)

// NewMemberRoleResource is the resource factory registered with the provider.
func NewMemberRoleResource() resource.Resource {
	return &memberRoleResource{}
}

// memberRoleResource models the assignment of a single role to a single member.
// Discord exposes only PUT/DELETE for this association and no GET, so reads
// verify membership via the guild member object. All attributes are immutable:
// any change describes a different association and forces replacement.
type memberRoleResource struct {
	client *conns.Client
}

type memberRoleResourceModel struct {
	GuildID types.String `tfsdk:"guild_id"`
	UserID  types.String `tfsdk:"user_id"`
	RoleID  types.String `tfsdk:"role_id"`
}

func (r *memberRoleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_member_role"
}

func (r *memberRoleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	requiresReplace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Assigns a single Discord role to a single guild member. Requires `MANAGE_ROLES`, and the bot's highest role must be above the assigned role.",
		Attributes: map[string]schema.Attribute{
			"guild_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Snowflake ID of the guild.",
				Validators:          []validator.String{validators.Snowflake()},
				PlanModifiers:       requiresReplace,
			},
			"user_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Snowflake ID of the member.",
				Validators:          []validator.String{validators.Snowflake()},
				PlanModifiers:       requiresReplace,
			},
			"role_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Snowflake ID of the role to assign.",
				Validators:          []validator.String{validators.Snowflake()},
				PlanModifiers:       requiresReplace,
			},
		},
	}
}

func (r *memberRoleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*conns.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data",
			fmt.Sprintf("expected *conns.Client, got %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *memberRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan memberRoleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := AddMemberRole(ctx, r.client, plan.GuildID.ValueString(), plan.UserID.ValueString(), plan.RoleID.ValueString(), "")
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("assigning Discord member role", plan.GuildID.ValueString(), err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *memberRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state memberRoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	member, err := GetMember(ctx, r.client, state.GuildID.ValueString(), state.UserID.ValueString())
	if err != nil {
		if conns.IsNotFound(err) {
			// The member left the guild; the association no longer exists.
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(diagutil.APIError("reading Discord guild member", state.GuildID.ValueString(), err))
		return
	}

	if !member.MemberHasRole(state.RoleID.ValueString()) {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is unreachable: every attribute forces replacement. It exists only to
// satisfy the resource interface.
func (r *memberRoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan memberRoleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *memberRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state memberRoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := RemoveMemberRole(ctx, r.client, state.GuildID.ValueString(), state.UserID.ValueString(), state.RoleID.ValueString(), "")
	if err != nil && !conns.IsNotFound(err) {
		resp.Diagnostics.Append(diagutil.APIError("removing Discord member role", state.GuildID.ValueString(), err))
	}
}

// ImportState accepts the "guild_id:user_id:role_id" format.
func (r *memberRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ":")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("expected import ID in the form \"guild_id:user_id:role_id\", got %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("guild_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("role_id"), parts[2])...)
}
