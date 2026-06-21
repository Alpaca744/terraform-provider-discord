package channel

import (
	"context"
	"fmt"
	"strings"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
	"github.com/alpaca744/terraform-provider-discord/internal/diagutil"
	"github.com/alpaca744/terraform-provider-discord/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	_ resource.Resource                = (*channelResource)(nil)
	_ resource.ResourceWithConfigure   = (*channelResource)(nil)
	_ resource.ResourceWithImportState = (*channelResource)(nil)
)

// NewChannelResource is the resource factory registered with the provider.
func NewChannelResource() resource.Resource {
	return &channelResource{}
}

type channelResource struct {
	client *conns.Client
}

type channelResourceModel struct {
	ID               types.String `tfsdk:"id"`
	GuildID          types.String `tfsdk:"guild_id"`
	Type             types.Int64  `tfsdk:"type"`
	Name             types.String `tfsdk:"name"`
	Topic            types.String `tfsdk:"topic"`
	Position         types.Int64  `tfsdk:"position"`
	NSFW             types.Bool   `tfsdk:"nsfw"`
	ParentID         types.String `tfsdk:"parent_id"`
	RateLimitPerUser types.Int64  `tfsdk:"rate_limit_per_user"`
	Bitrate          types.Int64  `tfsdk:"bitrate"`
	UserLimit        types.Int64  `tfsdk:"user_limit"`
}

func (r *channelResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_channel"
}

func (r *channelResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Discord guild channel. Requires the bot to have `MANAGE_CHANNELS`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"guild_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Snowflake ID of the guild. Changing this forces a new channel.",
				Validators:          []validator.String{validators.Snowflake()},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"type": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Discord channel type (0 text, 2 voice, 4 category, 5 announcement, 13 stage, 15 forum). Changing this forces a new channel.",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Channel name (1-100 characters).",
				Validators:          []validator.String{stringvalidator.LengthBetween(1, 100)},
			},
			"topic": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Channel topic (0-1024 characters).",
			},
			"position": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Sorting position of the channel.",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"nsfw": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the channel is NSFW.",
			},
			"parent_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Snowflake ID of the parent category channel.",
				Validators:          []validator.String{validators.Snowflake()},
			},
			"rate_limit_per_user": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Slowmode: seconds a user must wait between messages (0-21600).",
				Validators:          []validator.Int64{int64validator.Between(0, 21600)},
			},
			"bitrate": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Voice channel bitrate in bits per second.",
			},
			"user_limit": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Voice channel user limit (0 means no limit).",
			},
		},
	}
}

func (r *channelResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *channelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan channelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := createBody{
		Name:             plan.Name.ValueString(),
		Type:             plan.Type.ValueInt64(),
		Topic:            strPtr(plan.Topic),
		NSFW:             boolPtr(plan.NSFW),
		ParentID:         strPtr(plan.ParentID),
		RateLimitPerUser: int64Ptr(plan.RateLimitPerUser),
		Bitrate:          int64Ptr(plan.Bitrate),
		UserLimit:        int64Ptr(plan.UserLimit),
	}

	ch, err := create(ctx, r.client, plan.GuildID.ValueString(), body, "")
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("creating Discord channel", plan.GuildID.ValueString(), err))
		return
	}

	r.apply(&plan, ch)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *channelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state channelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ch, err := get(ctx, r.client, state.ID.ValueString())
	if err != nil {
		if conns.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(diagutil.APIError("reading Discord channel", state.ID.ValueString(), err))
		return
	}

	r.apply(&state, ch)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *channelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan channelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := modifyBody{
		Name:             strPtr(plan.Name),
		Topic:            strPtr(plan.Topic),
		NSFW:             boolPtr(plan.NSFW),
		ParentID:         strPtr(plan.ParentID),
		RateLimitPerUser: int64Ptr(plan.RateLimitPerUser),
		Bitrate:          int64Ptr(plan.Bitrate),
		UserLimit:        int64Ptr(plan.UserLimit),
	}

	ch, err := modify(ctx, r.client, plan.ID.ValueString(), body, "")
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("updating Discord channel", plan.ID.ValueString(), err))
		return
	}

	r.apply(&plan, ch)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *channelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state channelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := deleteChannel(ctx, r.client, state.ID.ValueString(), ""); err != nil && !conns.IsNotFound(err) {
		resp.Diagnostics.Append(diagutil.APIError("deleting Discord channel", state.ID.ValueString(), err))
	}
}

// ImportState accepts a bare channel ID; guild_id, type, and the rest are
// populated by the subsequent Read.
func (r *channelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id := strings.TrimSpace(req.ID)
	if id == "" {
		resp.Diagnostics.AddError("Invalid import ID", "expected a channel snowflake ID")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

// apply writes a Discord channel object into the Terraform model. Optional
// attributes are only set from the API when the corresponding field is
// meaningful for the channel type, to avoid spurious drift.
func (r *channelResource) apply(m *channelResourceModel, ch *Channel) {
	m.ID = types.StringValue(ch.ID)
	m.Type = types.Int64Value(ch.Type)
	m.Name = types.StringValue(ch.Name)
	m.Position = types.Int64Value(ch.Position)
	m.NSFW = types.BoolValue(ch.NSFW)

	if ch.GuildID != "" {
		m.GuildID = types.StringValue(ch.GuildID)
	}
	m.Topic = optStr(ch.Topic)
	m.ParentID = optStr(ch.ParentID)

	// Only reflect type-specific numeric fields when configured, so unset
	// optionals stay null rather than flipping to API zero-defaults.
	if !m.RateLimitPerUser.IsNull() {
		m.RateLimitPerUser = types.Int64Value(ch.RateLimitPerUser)
	}
	if !m.Bitrate.IsNull() {
		m.Bitrate = types.Int64Value(ch.Bitrate)
	}
	if !m.UserLimit.IsNull() {
		m.UserLimit = types.Int64Value(ch.UserLimit)
	}
}

func strPtr(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	return &s
}

func optStr(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

func boolPtr(v types.Bool) *bool {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	b := v.ValueBool()
	return &b
}

func int64Ptr(v types.Int64) *int64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	i := v.ValueInt64()
	return &i
}
