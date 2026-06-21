package guild

import (
	"context"
	"fmt"

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
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = (*guildSettingsResource)(nil)
	_ resource.ResourceWithConfigure   = (*guildSettingsResource)(nil)
	_ resource.ResourceWithImportState = (*guildSettingsResource)(nil)
)

// NewGuildSettingsResource is the resource factory registered with the provider.
func NewGuildSettingsResource() resource.Resource {
	return &guildSettingsResource{}
}

// guildSettingsResource manages settings on an existing guild. A guild cannot be
// created or deleted through the bot API, so Create applies settings to the
// existing guild and Delete is a no-op that simply drops the resource from
// state, leaving the guild's configuration untouched.
type guildSettingsResource struct {
	client *conns.Client
}

type guildSettingsModel struct {
	GuildID                     types.String `tfsdk:"guild_id"`
	Name                        types.String `tfsdk:"name"`
	Description                 types.String `tfsdk:"description"`
	VerificationLevel           types.Int64  `tfsdk:"verification_level"`
	DefaultMessageNotifications types.Int64  `tfsdk:"default_message_notifications"`
	ExplicitContentFilter       types.Int64  `tfsdk:"explicit_content_filter"`
	AFKChannelID                types.String `tfsdk:"afk_channel_id"`
	AFKTimeout                  types.Int64  `tfsdk:"afk_timeout"`
	SystemChannelID             types.String `tfsdk:"system_channel_id"`
	RulesChannelID              types.String `tfsdk:"rules_channel_id"`
	PublicUpdatesChannelID      types.String `tfsdk:"public_updates_channel_id"`
	PreferredLocale             types.String `tfsdk:"preferred_locale"`
	PremiumProgressBarEnabled   types.Bool   `tfsdk:"premium_progress_bar_enabled"`
}

func (r *guildSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_guild_settings"
}

func (r *guildSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	optComputedStr := func(desc string) schema.StringAttribute {
		return schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: desc}
	}
	optComputedInt := func(desc string) schema.Int64Attribute {
		return schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: desc}
	}
	optComputedSnowflake := func(desc string) schema.StringAttribute {
		return schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: desc, Validators: []validator.String{validators.Snowflake()}}
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages settings on an existing Discord guild. The guild itself is not created or destroyed by this resource; deleting it from Terraform leaves the guild's current configuration in place.",
		Attributes: map[string]schema.Attribute{
			"guild_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Snowflake ID of the guild. Changing this forces a new resource.",
				Validators:          []validator.String{validators.Snowflake()},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name":                          optComputedStr("Guild name."),
			"description":                   schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Guild description (community guilds)."},
			"verification_level":            optComputedInt("Verification level (0-4)."),
			"default_message_notifications": optComputedInt("Default message notification level (0 all messages, 1 only mentions)."),
			"explicit_content_filter":       optComputedInt("Explicit content filter level (0-2)."),
			"afk_channel_id":                optComputedSnowflake("Snowflake ID of the AFK voice channel."),
			"afk_timeout":                   optComputedInt("AFK timeout in seconds (60, 300, 900, 1800, 3600)."),
			"system_channel_id":             optComputedSnowflake("Snowflake ID of the system messages channel."),
			"rules_channel_id":              optComputedSnowflake("Snowflake ID of the rules channel (community guilds)."),
			"public_updates_channel_id":     optComputedSnowflake("Snowflake ID of the public updates channel (community guilds)."),
			"preferred_locale":              optComputedStr("Preferred locale of a community guild."),
			"premium_progress_bar_enabled":  schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Whether the boost progress bar is enabled."},
		},
	}
}

func (r *guildSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *guildSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan guildSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	g, err := ModifyGuild(ctx, r.client, plan.GuildID.ValueString(), r.body(plan), "")
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("applying Discord guild settings", plan.GuildID.ValueString(), err))
		return
	}
	r.apply(&plan, g)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *guildSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state guildSettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	g, err := GetGuild(ctx, r.client, state.GuildID.ValueString())
	if err != nil {
		if conns.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(diagutil.APIError("reading Discord guild settings", state.GuildID.ValueString(), err))
		return
	}
	r.apply(&state, g)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *guildSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan guildSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	g, err := ModifyGuild(ctx, r.client, plan.GuildID.ValueString(), r.body(plan), "")
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("updating Discord guild settings", plan.GuildID.ValueString(), err))
		return
	}
	r.apply(&plan, g)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete is intentionally a no-op: a guild's settings cannot be "deleted", and
// the guild is not owned by this resource. State removal is automatic.
func (r *guildSettingsResource) Delete(ctx context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	tflog.Info(ctx, "discord_guild_settings deletion is a no-op; the guild configuration is left unchanged")
}

// ImportState accepts a bare guild ID.
func (r *guildSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("guild_id"), req.ID)...)
}

func (r *guildSettingsResource) body(m guildSettingsModel) GuildSettingsBody {
	return GuildSettingsBody{
		Name:                        strPtr(m.Name),
		Description:                 strPtr(m.Description),
		VerificationLevel:           int64Ptr(m.VerificationLevel),
		DefaultMessageNotifications: int64Ptr(m.DefaultMessageNotifications),
		ExplicitContentFilter:       int64Ptr(m.ExplicitContentFilter),
		AFKChannelID:                strPtr(m.AFKChannelID),
		AFKTimeout:                  int64Ptr(m.AFKTimeout),
		SystemChannelID:             strPtr(m.SystemChannelID),
		RulesChannelID:              strPtr(m.RulesChannelID),
		PublicUpdatesChannelID:      strPtr(m.PublicUpdatesChannelID),
		PreferredLocale:             strPtr(m.PreferredLocale),
		PremiumProgressBarEnabled:   boolPtr(m.PremiumProgressBarEnabled),
	}
}

func (r *guildSettingsResource) apply(m *guildSettingsModel, g *Guild) {
	m.GuildID = types.StringValue(g.ID)
	m.Name = types.StringValue(g.Name)
	m.VerificationLevel = types.Int64Value(g.VerificationLevel)
	m.DefaultMessageNotifications = types.Int64Value(g.DefaultMessageNotifications)
	m.ExplicitContentFilter = types.Int64Value(g.ExplicitContentFilter)
	m.AFKTimeout = types.Int64Value(g.AFKTimeout)
	m.PreferredLocale = types.StringValue(g.PreferredLocale)
	m.PremiumProgressBarEnabled = types.BoolValue(g.PremiumProgressBarEnabled)

	// Optional (non-computed) attributes: keep null when Discord reports empty,
	// so an unset config value does not show perpetual drift.
	m.Description = optStr(g.Description)
	m.AFKChannelID = optStr(g.AFKChannelID)
	m.SystemChannelID = optStr(g.SystemChannelID)
	m.RulesChannelID = optStr(g.RulesChannelID)
	m.PublicUpdatesChannelID = optStr(g.PublicUpdatesChannelID)
}
