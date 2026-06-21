package invite

import (
	"context"
	"fmt"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
	"github.com/alpaca744/terraform-provider-discord/internal/diagutil"
	"github.com/alpaca744/terraform-provider-discord/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*inviteResource)(nil)
	_ resource.ResourceWithConfigure   = (*inviteResource)(nil)
	_ resource.ResourceWithImportState = (*inviteResource)(nil)
)

// NewInviteResource is the resource factory.
func NewInviteResource() resource.Resource {
	return &inviteResource{}
}

type inviteResource struct {
	client *conns.Client
}

type inviteModel struct {
	Code      types.String `tfsdk:"code"`
	ChannelID types.String `tfsdk:"channel_id"`
	GuildID   types.String `tfsdk:"guild_id"`
	MaxAge    types.Int64  `tfsdk:"max_age"`
	MaxUses   types.Int64  `tfsdk:"max_uses"`
	Temporary types.Bool   `tfsdk:"temporary"`
	Unique    types.Bool   `tfsdk:"unique"`
	URL       types.String `tfsdk:"url"`
}

func (r *inviteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_invite"
}

func (r *inviteResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	forceNewStr := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Discord channel invite. Invites are immutable, so any change recreates the invite.",
		Attributes: map[string]schema.Attribute{
			"code": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique invite code (also the resource ID).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"channel_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Snowflake ID of the channel the invite targets. Changing this forces a new invite.",
				Validators:          []validator.String{validators.Snowflake()},
				PlanModifiers:       forceNewStr,
			},
			"guild_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Snowflake ID of the guild the invite belongs to.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"max_age": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(86400),
				MarkdownDescription: "Duration in seconds before expiry (0 = never). Default 86400 (24h). Changing forces a new invite.",
				Validators:          []validator.Int64{int64validator.Between(0, 604800)},
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"max_uses": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				MarkdownDescription: "Maximum number of uses (0 = unlimited). Changing forces a new invite.",
				Validators:          []validator.Int64{int64validator.Between(0, 100)},
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"temporary": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether the invite grants temporary membership. Changing forces a new invite.",
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
			},
			"unique": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether to always create a new unique invite rather than reuse a similar one. Changing forces a new invite.",
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
			},
			"url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The full invite URL (https://discord.gg/<code>).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *inviteResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *inviteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan inviteModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := createBody{
		MaxAge:    plan.MaxAge.ValueInt64(),
		MaxUses:   plan.MaxUses.ValueInt64(),
		Temporary: plan.Temporary.ValueBool(),
		Unique:    plan.Unique.ValueBool(),
	}
	inv, err := createOnChannel(ctx, r.client, plan.ChannelID.ValueString(), body, "")
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("creating Discord invite", plan.ChannelID.ValueString(), err))
		return
	}
	r.apply(&plan, inv)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *inviteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state inviteModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inv, err := get(ctx, r.client, state.Code.ValueString())
	if err != nil {
		if conns.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(diagutil.APIError("reading Discord invite", state.Code.ValueString(), err))
		return
	}
	// Only refresh existence-derived fields; metadata (max_age/uses/etc.) is not
	// reliably returned by GET and is treated as authoritative from config.
	r.applyExisting(&state, inv)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is unreachable: every configurable attribute forces replacement.
func (r *inviteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan inviteModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *inviteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state inviteModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := deleteByCode(ctx, r.client, state.Code.ValueString(), ""); err != nil && !conns.IsNotFound(err) {
		resp.Diagnostics.Append(diagutil.APIError("deleting Discord invite", state.Code.ValueString(), err))
	}
}

// ImportState accepts the invite code. Create-time metadata (max_age, etc.)
// cannot be recovered and will reflect schema defaults after import.
func (r *inviteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("code"), req.ID)...)
}

func (r *inviteResource) apply(m *inviteModel, inv *Invite) {
	m.Code = types.StringValue(inv.Code)
	m.ChannelID = types.StringValue(inv.Channel.ID)
	m.GuildID = optStr(inv.Guild.ID)
	m.URL = types.StringValue("https://discord.gg/" + inv.Code)
}

func (r *inviteResource) applyExisting(m *inviteModel, inv *Invite) {
	m.Code = types.StringValue(inv.Code)
	if inv.Channel.ID != "" {
		m.ChannelID = types.StringValue(inv.Channel.ID)
	}
	m.GuildID = optStr(inv.Guild.ID)
	m.URL = types.StringValue("https://discord.gg/" + inv.Code)
}

func optStr(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}
