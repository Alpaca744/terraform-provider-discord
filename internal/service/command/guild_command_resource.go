package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
	"github.com/alpaca744/terraform-provider-discord/internal/diagutil"
	"github.com/alpaca744/terraform-provider-discord/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*guildCommandResource)(nil)
	_ resource.ResourceWithConfigure   = (*guildCommandResource)(nil)
	_ resource.ResourceWithImportState = (*guildCommandResource)(nil)
)

// NewGuildCommandResource is the resource factory for guild commands.
func NewGuildCommandResource() resource.Resource {
	return &guildCommandResource{}
}

type guildCommandResource struct {
	client *conns.Client
}

type guildCommandModel struct {
	ID                       types.String         `tfsdk:"id"`
	ApplicationID            types.String         `tfsdk:"application_id"`
	GuildID                  types.String         `tfsdk:"guild_id"`
	Name                     types.String         `tfsdk:"name"`
	Description              types.String         `tfsdk:"description"`
	Type                     types.Int64          `tfsdk:"type"`
	DefaultMemberPermissions types.Set            `tfsdk:"default_member_permissions"`
	NSFW                     types.Bool           `tfsdk:"nsfw"`
	Options                  jsontypes.Normalized `tfsdk:"options"`
	Version                  types.String         `tfsdk:"version"`
}

func (r *guildCommandResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_guild_application_command"
}

func (r *guildCommandResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a guild-scoped Discord application command. Guild commands appear only in the specified guild and update instantly.",
		Attributes: map[string]schema.Attribute{
			"id":             computedID(),
			"application_id": applicationIDAttr(),
			"guild_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Snowflake ID of the guild. Changing this forces a new command.",
				Validators:          []validator.String{validators.Snowflake()},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name":                       nameAttr(),
			"description":                descriptionAttr(),
			"type":                       typeAttr(),
			"default_member_permissions": defaultPermsAttr(),
			"nsfw":                       nsfwAttr(),
			"options":                    optionsAttr(),
			"version":                    versionAttr(),
		},
	}
}

func (r *guildCommandResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req, resp)
}

func (r *guildCommandResource) base(m guildCommandModel) string {
	return GuildBasePath(m.ApplicationID.ValueString(), m.GuildID.ValueString())
}

func (r *guildCommandResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan guildCommandModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := r.body(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cmd, err := CreateCommand(ctx, r.client, r.base(plan), body)
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("creating Discord guild application command", plan.GuildID.ValueString(), err))
		return
	}
	resp.Diagnostics.Append(r.apply(ctx, &plan, cmd)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *guildCommandResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state guildCommandModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cmd, err := GetCommand(ctx, r.client, r.base(state), state.ID.ValueString())
	if err != nil {
		if conns.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(diagutil.APIError("reading Discord guild application command", state.GuildID.ValueString(), err))
		return
	}
	resp.Diagnostics.Append(r.apply(ctx, &state, cmd)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *guildCommandResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan guildCommandModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := r.body(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cmd, err := EditCommand(ctx, r.client, r.base(plan), plan.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("updating Discord guild application command", plan.GuildID.ValueString(), err))
		return
	}
	resp.Diagnostics.Append(r.apply(ctx, &plan, cmd)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *guildCommandResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state guildCommandModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := DeleteCommand(ctx, r.client, r.base(state), state.ID.ValueString())
	if err != nil && !conns.IsNotFound(err) {
		resp.Diagnostics.Append(diagutil.APIError("deleting Discord guild application command", state.GuildID.ValueString(), err))
	}
}

// ImportState accepts the "application_id:guild_id:command_id" format.
func (r *guildCommandResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ":")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("expected import ID in the form \"application_id:guild_id:command_id\", got %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("application_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("guild_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[2])...)
}

func (r *guildCommandResource) body(ctx context.Context, plan guildCommandModel) (WriteBody, diag.Diagnostics) {
	perms, diags := expandDefaultPerms(ctx, plan.DefaultMemberPermissions)
	return WriteBody{
		Name:                     plan.Name.ValueString(),
		Description:              plan.Description.ValueString(),
		Type:                     plan.Type.ValueInt64(),
		DefaultMemberPermissions: perms,
		NSFW:                     plan.NSFW.ValueBool(),
		Options:                  expandOptions(plan.Options),
	}, diags
}

func (r *guildCommandResource) apply(ctx context.Context, m *guildCommandModel, cmd *Command) diag.Diagnostics {
	m.ID = types.StringValue(cmd.ID)
	m.ApplicationID = types.StringValue(cmd.ApplicationID)
	if cmd.GuildID != "" {
		m.GuildID = types.StringValue(cmd.GuildID)
	}
	m.Name = types.StringValue(cmd.Name)
	m.Description = optStr(cmd.Description)
	m.Type = types.Int64Value(cmd.Type)
	m.NSFW = types.BoolValue(cmd.NSFW)
	m.Version = types.StringValue(cmd.Version)
	m.Options = flattenOptions(cmd.Options)

	perms, diags := flattenDefaultPerms(ctx, cmd.DefaultMemberPermissions)
	m.DefaultMemberPermissions = perms
	return diags
}
