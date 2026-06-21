// Package provider wires the Discord Terraform provider: configuration schema,
// credential resolution, the shared API client, and the resource/data-source
// registries.
package provider

import (
	"context"
	"os"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
	"github.com/alpaca744/terraform-provider-discord/internal/service/application"
	"github.com/alpaca744/terraform-provider-discord/internal/service/automod"
	"github.com/alpaca744/terraform-provider-discord/internal/service/channel"
	"github.com/alpaca744/terraform-provider-discord/internal/service/command"
	"github.com/alpaca744/terraform-provider-discord/internal/service/emoji"
	"github.com/alpaca744/terraform-provider-discord/internal/service/guild"
	"github.com/alpaca744/terraform-provider-discord/internal/service/invite"
	"github.com/alpaca744/terraform-provider-discord/internal/service/monetization"
	"github.com/alpaca744/terraform-provider-discord/internal/service/scheduledevent"
	"github.com/alpaca744/terraform-provider-discord/internal/service/soundboard"
	"github.com/alpaca744/terraform-provider-discord/internal/service/stage"
	"github.com/alpaca744/terraform-provider-discord/internal/service/sticker"
	"github.com/alpaca744/terraform-provider-discord/internal/service/user"
	"github.com/alpaca744/terraform-provider-discord/internal/service/voice"
	"github.com/alpaca744/terraform-provider-discord/internal/service/webhook"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the framework interface.
var _ provider.Provider = (*discordProvider)(nil)

type discordProvider struct {
	// version is injected at build time and surfaced in the provider metadata.
	version string
}

// New returns a provider factory for the given build version.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &discordProvider{version: version}
	}
}

// providerModel is the typed configuration model. Every attribute is optional so
// values can fall back to environment variables during Configure.
type providerModel struct {
	BotToken              types.String `tfsdk:"bot_token"`
	BearerToken           types.String `tfsdk:"bearer_token"`
	ClientID              types.String `tfsdk:"client_id"`
	ClientSecret          types.String `tfsdk:"client_secret"`
	APIBaseURL            types.String `tfsdk:"api_base_url"`
	DefaultAuditLogReason types.String `tfsdk:"default_audit_log_reason"`
}

func (p *discordProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "discord"
	resp.Version = p.version
}

func (p *discordProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage durable Discord configuration through the Discord REST API.",
		Attributes: map[string]schema.Attribute{
			"bot_token": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Discord bot token. Primary authentication for guild/server management. Falls back to `DISCORD_BOT_TOKEN`.",
			},
			"bearer_token": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "OAuth2 bearer token, required only for endpoints Discord restricts to OAuth (e.g. application command permissions). Falls back to `DISCORD_BEARER_TOKEN`.",
			},
			"client_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "OAuth2 client ID. Falls back to `DISCORD_CLIENT_ID`.",
			},
			"client_secret": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "OAuth2 client secret. Falls back to `DISCORD_CLIENT_SECRET`.",
			},
			"api_base_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Discord API base URL. Defaults to `" + conns.DefaultAPIBaseURL + "`. Falls back to `DISCORD_API_BASE_URL`.",
			},
			"default_audit_log_reason": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Default `X-Audit-Log-Reason` sent with audited actions. Falls back to `DISCORD_AUDIT_LOG_REASON`.",
			},
		},
	}
}

func (p *discordProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := conns.Config{
		BotToken:       resolve(data.BotToken, "DISCORD_BOT_TOKEN"),
		BearerToken:    resolve(data.BearerToken, "DISCORD_BEARER_TOKEN"),
		ClientID:       resolve(data.ClientID, "DISCORD_CLIENT_ID"),
		ClientSecret:   resolve(data.ClientSecret, "DISCORD_CLIENT_SECRET"),
		APIBaseURL:     resolve(data.APIBaseURL, "DISCORD_API_BASE_URL"),
		AuditLogReason: resolve(data.DefaultAuditLogReason, "DISCORD_AUDIT_LOG_REASON"),
	}

	client, err := conns.NewClient(cfg)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Discord provider configuration", err.Error())
		return
	}

	// Share the client with resources and data sources.
	resp.ResourceData = client
	resp.DataSourceData = client
}

func (p *discordProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		guild.NewRoleResource,
		guild.NewMemberRoleResource,
		guild.NewGuildSettingsResource,
		channel.NewChannelResource,
		channel.NewOverwriteResource,
		webhook.NewWebhookResource,
		automod.NewRuleResource,
		application.NewSettingsResource,
		application.NewRoleConnectionMetadataResource,
		command.NewGlobalCommandResource,
		command.NewGuildCommandResource,
		stage.NewStageInstanceResource,
		invite.NewInviteResource,
		emoji.NewEmojiResource,
		soundboard.NewSoundResource,
		scheduledevent.NewEventResource,
		guild.NewWidgetResource,
		guild.NewWelcomeScreenResource,
		guild.NewOnboardingResource,
		guild.NewTemplateResource,
		sticker.NewStickerResource,
	}
}

func (p *discordProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		guild.NewGuildDataSource,
		guild.NewGuildsDataSource,
		guild.NewGuildPreviewDataSource,
		guild.NewRolesDataSource,
		guild.NewRoleDataSource,
		channel.NewChannelsDataSource,
		channel.NewChannelDataSource,
		webhook.NewWebhookDataSource,
		invite.NewInviteDataSource,
		application.NewCurrentApplicationDataSource,
		user.NewCurrentUserDataSource,
		user.NewUserDataSource,
		voice.NewVoiceRegionsDataSource,
		guild.NewAuditLogDataSource,
		monetization.NewSKUDataSource,
		monetization.NewEntitlementDataSource,
		monetization.NewSubscriptionDataSource,
	}
}

// resolve returns the configured value, or the environment fallback when the
// config value is null/unknown/empty.
func resolve(v types.String, envKey string) string {
	if !v.IsNull() && !v.IsUnknown() && v.ValueString() != "" {
		return v.ValueString()
	}
	return os.Getenv(envKey)
}
