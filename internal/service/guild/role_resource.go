package guild

import (
	"context"
	"fmt"
	"strings"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
	"github.com/alpaca744/terraform-provider-discord/internal/diagutil"
	"github.com/alpaca744/terraform-provider-discord/internal/discord"
	"github.com/alpaca744/terraform-provider-discord/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*roleResource)(nil)
	_ resource.ResourceWithConfigure   = (*roleResource)(nil)
	_ resource.ResourceWithImportState = (*roleResource)(nil)
)

// NewRoleResource is the resource factory registered with the provider.
func NewRoleResource() resource.Resource {
	return &roleResource{}
}

type roleResource struct {
	client *conns.Client
}

type roleResourceModel struct {
	ID          types.String `tfsdk:"id"`
	GuildID     types.String `tfsdk:"guild_id"`
	Name        types.String `tfsdk:"name"`
	Color       types.Int64  `tfsdk:"color"`
	Hoist       types.Bool   `tfsdk:"hoist"`
	Mentionable types.Bool   `tfsdk:"mentionable"`
	Permissions types.Set    `tfsdk:"permissions"`
	Position    types.Int64  `tfsdk:"position"`
	Managed     types.Bool   `tfsdk:"managed"`
}

func (r *roleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (r *roleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Discord guild role. Requires the bot to have the `MANAGE_ROLES` permission.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The role snowflake ID.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"guild_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The snowflake ID of the guild the role belongs to. Changing this forces a new role.",
				Validators:          []validator.String{validators.Snowflake()},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The role name. Defaults to Discord's `new role` when unset.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"color": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Integer RGB color of the role (0 means no color / inherit).",
			},
			"hoist": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the role is displayed separately in the member list.",
			},
			"mentionable": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the role can be mentioned by anyone.",
			},
			"permissions": schema.SetAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Set of Discord permission names granted to the role (e.g. `VIEW_CHANNEL`).",
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(validators.PermissionName()),
				},
				PlanModifiers: []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
			},
			"position": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The position of the role. Read-only; Discord assigns positions via a separate reorder endpoint not modeled here.",
			},
			"managed": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the role is managed by an integration and cannot be edited.",
			},
		},
	}
}

func (r *roleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *roleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan roleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildBody(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	role, err := CreateRole(ctx, r.client, plan.GuildID.ValueString(), body, "")
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("creating Discord role", plan.GuildID.ValueString(), err))
		return
	}

	resp.Diagnostics.Append(r.apply(ctx, &plan, role)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *roleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state roleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	role, err := GetRole(ctx, r.client, state.GuildID.ValueString(), state.ID.ValueString())
	if err != nil {
		if conns.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(diagutil.APIError("reading Discord role", state.GuildID.ValueString(), err))
		return
	}

	resp.Diagnostics.Append(r.apply(ctx, &state, role)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *roleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan roleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildBody(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	role, err := ModifyRole(ctx, r.client, plan.GuildID.ValueString(), plan.ID.ValueString(), body, "")
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("updating Discord role", plan.GuildID.ValueString(), err))
		return
	}

	resp.Diagnostics.Append(r.apply(ctx, &plan, role)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *roleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state roleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := DeleteRole(ctx, r.client, state.GuildID.ValueString(), state.ID.ValueString(), "")
	if err != nil && !conns.IsNotFound(err) {
		resp.Diagnostics.Append(diagutil.APIError("deleting Discord role", state.GuildID.ValueString(), err))
	}
}

// ImportState accepts the documented "guild_id:role_id" import format.
func (r *roleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("expected import ID in the form \"guild_id:role_id\", got %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("guild_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

// buildBody translates the plan into a Discord write payload.
func (r *roleResource) buildBody(ctx context.Context, plan roleResourceModel) (roleWriteBody, diag.Diagnostics) {
	var body roleWriteBody
	var diags diag.Diagnostics

	if !plan.Name.IsNull() && !plan.Name.IsUnknown() {
		v := plan.Name.ValueString()
		body.Name = &v
	}
	if !plan.Color.IsNull() && !plan.Color.IsUnknown() {
		v := plan.Color.ValueInt64()
		body.Color = &v
	}
	if !plan.Hoist.IsNull() && !plan.Hoist.IsUnknown() {
		v := plan.Hoist.ValueBool()
		body.Hoist = &v
	}
	if !plan.Mentionable.IsNull() && !plan.Mentionable.IsUnknown() {
		v := plan.Mentionable.ValueBool()
		body.Mentionable = &v
	}
	if !plan.Permissions.IsNull() && !plan.Permissions.IsUnknown() {
		var names []string
		diags.Append(plan.Permissions.ElementsAs(ctx, &names, false)...)
		if diags.HasError() {
			return body, diags
		}
		bitfield, err := discord.PermissionsToBitfield(names)
		if err != nil {
			diags.AddError("Invalid permissions", err.Error())
			return body, diags
		}
		body.Permissions = &bitfield
	}
	return body, diags
}

// apply writes a Discord role object back into the Terraform model.
func (r *roleResource) apply(ctx context.Context, m *roleResourceModel, role *Role) diag.Diagnostics {
	var diags diag.Diagnostics

	m.ID = types.StringValue(role.ID)
	m.Name = types.StringValue(role.Name)
	m.Color = types.Int64Value(role.Color)
	m.Hoist = types.BoolValue(role.Hoist)
	m.Mentionable = types.BoolValue(role.Mentionable)
	m.Position = types.Int64Value(role.Position)
	m.Managed = types.BoolValue(role.Managed)

	names, err := discord.BitfieldToPermissions(role.Permissions)
	if err != nil {
		diags.AddError("Invalid permissions from API", err.Error())
		return diags
	}
	set, d := types.SetValueFrom(ctx, types.StringType, names)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	m.Permissions = set
	return diags
}
