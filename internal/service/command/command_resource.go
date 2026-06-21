package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
	"github.com/alpaca744/terraform-provider-discord/internal/diagutil"
	"github.com/alpaca744/terraform-provider-discord/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*globalCommandResource)(nil)
	_ resource.ResourceWithConfigure   = (*globalCommandResource)(nil)
	_ resource.ResourceWithImportState = (*globalCommandResource)(nil)
)

// NewGlobalCommandResource is the resource factory for global commands.
func NewGlobalCommandResource() resource.Resource {
	return &globalCommandResource{}
}

type globalCommandResource struct {
	client *conns.Client
}

type globalCommandModel struct {
	ID                       types.String         `tfsdk:"id"`
	ApplicationID            types.String         `tfsdk:"application_id"`
	Name                     types.String         `tfsdk:"name"`
	Description              types.String         `tfsdk:"description"`
	Type                     types.Int64          `tfsdk:"type"`
	DefaultMemberPermissions types.Set            `tfsdk:"default_member_permissions"`
	DMPermission             types.Bool           `tfsdk:"dm_permission"`
	NSFW                     types.Bool           `tfsdk:"nsfw"`
	Options                  jsontypes.Normalized `tfsdk:"options"`
	Version                  types.String         `tfsdk:"version"`
}

func (r *globalCommandResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_command"
}

func (r *globalCommandResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a global Discord application command (slash, user, or message command).",
		Attributes: map[string]schema.Attribute{
			"id":                         computedID(),
			"application_id":             applicationIDAttr(),
			"name":                       nameAttr(),
			"description":                descriptionAttr(),
			"type":                       typeAttr(),
			"default_member_permissions": defaultPermsAttr(),
			"nsfw":                       nsfwAttr(),
			"dm_permission": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the command is available in DMs with the app (global commands only).",
			},
			"options": optionsAttr(),
			"version": versionAttr(),
		},
	}
}

func (r *globalCommandResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req, resp)
}

func (r *globalCommandResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan globalCommandModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := r.body(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cmd, err := CreateCommand(ctx, r.client, GlobalBasePath(plan.ApplicationID.ValueString()), body)
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("creating Discord application command", plan.ApplicationID.ValueString(), err))
		return
	}
	resp.Diagnostics.Append(r.apply(ctx, &plan, cmd)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *globalCommandResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state globalCommandModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cmd, err := GetCommand(ctx, r.client, GlobalBasePath(state.ApplicationID.ValueString()), state.ID.ValueString())
	if err != nil {
		if conns.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(diagutil.APIError("reading Discord application command", state.ApplicationID.ValueString(), err))
		return
	}
	resp.Diagnostics.Append(r.apply(ctx, &state, cmd)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *globalCommandResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan globalCommandModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := r.body(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cmd, err := EditCommand(ctx, r.client, GlobalBasePath(plan.ApplicationID.ValueString()), plan.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("updating Discord application command", plan.ApplicationID.ValueString(), err))
		return
	}
	resp.Diagnostics.Append(r.apply(ctx, &plan, cmd)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *globalCommandResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state globalCommandModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := DeleteCommand(ctx, r.client, GlobalBasePath(state.ApplicationID.ValueString()), state.ID.ValueString())
	if err != nil && !conns.IsNotFound(err) {
		resp.Diagnostics.Append(diagutil.APIError("deleting Discord application command", state.ApplicationID.ValueString(), err))
	}
}

// ImportState accepts the "application_id:command_id" format.
func (r *globalCommandResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("expected import ID in the form \"application_id:command_id\", got %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("application_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func (r *globalCommandResource) body(ctx context.Context, plan globalCommandModel) (WriteBody, diag.Diagnostics) {
	perms, diags := expandDefaultPerms(ctx, plan.DefaultMemberPermissions)
	body := WriteBody{
		Name:                     plan.Name.ValueString(),
		Description:              plan.Description.ValueString(),
		Type:                     plan.Type.ValueInt64(),
		DefaultMemberPermissions: perms,
		NSFW:                     plan.NSFW.ValueBool(),
		Options:                  expandOptions(plan.Options),
	}
	if !plan.DMPermission.IsNull() && !plan.DMPermission.IsUnknown() {
		v := plan.DMPermission.ValueBool()
		body.DMPermission = &v
	}
	return body, diags
}

func (r *globalCommandResource) apply(ctx context.Context, m *globalCommandModel, cmd *Command) diag.Diagnostics {
	m.ID = types.StringValue(cmd.ID)
	m.ApplicationID = types.StringValue(cmd.ApplicationID)
	m.Name = types.StringValue(cmd.Name)
	m.Description = optStr(cmd.Description)
	m.Type = types.Int64Value(cmd.Type)
	m.NSFW = types.BoolValue(cmd.NSFW)
	m.Version = types.StringValue(cmd.Version)
	if cmd.DMPermission != nil {
		m.DMPermission = types.BoolValue(*cmd.DMPermission)
	} else {
		m.DMPermission = types.BoolValue(true)
	}
	// Preserve the configured options (plan on create/update, prior state on
	// read) when they are semantically equal to what Discord returned. The API
	// echoes options with extra null keys and reordered members, which would
	// otherwise differ from the planned value and trigger an "inconsistent result
	// after apply" error. Only adopt the API value on a genuine difference.
	if !optionsSemanticEqual(m.Options, flattenOptions(cmd.Options)) {
		m.Options = flattenOptions(cmd.Options)
	}

	perms, diags := flattenDefaultPerms(ctx, cmd.DefaultMemberPermissions)
	m.DefaultMemberPermissions = perms
	return diags
}

// schema attribute helpers shared by both command resources.

func computedID() schema.StringAttribute {
	return schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "The command snowflake ID.",
		PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	}
}

func applicationIDAttr() schema.StringAttribute {
	return schema.StringAttribute{
		Required:            true,
		MarkdownDescription: "Snowflake ID of the application. Changing this forces a new command.",
		Validators:          []validator.String{validators.Snowflake()},
		PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
	}
}

func nameAttr() schema.StringAttribute {
	return schema.StringAttribute{
		Required:            true,
		MarkdownDescription: "Command name (1-32 characters).",
	}
}

func descriptionAttr() schema.StringAttribute {
	return schema.StringAttribute{
		Optional:            true,
		MarkdownDescription: "Command description (required for CHAT_INPUT/slash commands, 1-100 characters).",
	}
}

func typeAttr() schema.Int64Attribute {
	return schema.Int64Attribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: "Command type: 1 CHAT_INPUT (default), 2 USER, 3 MESSAGE. Changing forces a new command.",
		PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
	}
}

func defaultPermsAttr() schema.SetAttribute {
	return schema.SetAttribute{
		Optional:            true,
		ElementType:         types.StringType,
		MarkdownDescription: "Permission names a member needs by default to use the command. Empty/unset means everyone.",
		Validators:          []validator.Set{setvalidator.ValueStringsAre(validators.PermissionName())},
	}
}

func nsfwAttr() schema.BoolAttribute {
	return schema.BoolAttribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: "Whether the command is age-restricted.",
	}
}

func optionsAttr() schema.StringAttribute {
	return schema.StringAttribute{
		Optional:            true,
		CustomType:          jsontypes.NormalizedType{},
		MarkdownDescription: "Command options as a JSON array. Carried as JSON to support arbitrarily nested subcommands and groups. Compared semantically, so formatting differences do not cause drift.",
	}
}

func versionAttr() schema.StringAttribute {
	return schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "Autoincrementing version identifier updated on each command change.",
	}
}

func configureClient(req resource.ConfigureRequest, resp *resource.ConfigureResponse) *conns.Client {
	if req.ProviderData == nil {
		return nil
	}
	client, ok := req.ProviderData.(*conns.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data",
			fmt.Sprintf("expected *conns.Client, got %T", req.ProviderData))
		return nil
	}
	return client
}
