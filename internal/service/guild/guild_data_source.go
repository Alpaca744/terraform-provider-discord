package guild

import (
	"context"
	"fmt"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
	"github.com/alpaca744/terraform-provider-discord/internal/diagutil"
	"github.com/alpaca744/terraform-provider-discord/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*guildDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*guildDataSource)(nil)
)

// NewGuildDataSource is the data source factory registered with the provider.
func NewGuildDataSource() datasource.DataSource {
	return &guildDataSource{}
}

type guildDataSource struct {
	client *conns.Client
}

type guildDataSourceModel struct {
	ID                          types.String `tfsdk:"id"`
	Name                        types.String `tfsdk:"name"`
	Description                 types.String `tfsdk:"description"`
	OwnerID                     types.String `tfsdk:"owner_id"`
	Icon                        types.String `tfsdk:"icon"`
	Splash                      types.String `tfsdk:"splash"`
	Banner                      types.String `tfsdk:"banner"`
	AFKChannelID                types.String `tfsdk:"afk_channel_id"`
	AFKTimeout                  types.Int64  `tfsdk:"afk_timeout"`
	VerificationLevel           types.Int64  `tfsdk:"verification_level"`
	DefaultMessageNotifications types.Int64  `tfsdk:"default_message_notifications"`
	ExplicitContentFilter       types.Int64  `tfsdk:"explicit_content_filter"`
	PreferredLocale             types.String `tfsdk:"preferred_locale"`
	PremiumTier                 types.Int64  `tfsdk:"premium_tier"`
	PremiumSubscriptionCount    types.Int64  `tfsdk:"premium_subscription_count"`
}

func (d *guildDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_guild"
}

func (d *guildDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Discord guild (server) by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The snowflake ID of the guild to read.",
				Validators:          []validator.String{validators.Snowflake()},
			},
			"name":                          schema.StringAttribute{Computed: true, MarkdownDescription: "The guild name."},
			"description":                   schema.StringAttribute{Computed: true, MarkdownDescription: "The guild description."},
			"owner_id":                      schema.StringAttribute{Computed: true, MarkdownDescription: "Snowflake ID of the guild owner."},
			"icon":                          schema.StringAttribute{Computed: true, MarkdownDescription: "The guild icon hash."},
			"splash":                        schema.StringAttribute{Computed: true, MarkdownDescription: "The guild splash hash."},
			"banner":                        schema.StringAttribute{Computed: true, MarkdownDescription: "The guild banner hash."},
			"afk_channel_id":                schema.StringAttribute{Computed: true, MarkdownDescription: "Snowflake ID of the AFK channel."},
			"afk_timeout":                   schema.Int64Attribute{Computed: true, MarkdownDescription: "AFK timeout in seconds."},
			"verification_level":            schema.Int64Attribute{Computed: true, MarkdownDescription: "Verification level required for the guild."},
			"default_message_notifications": schema.Int64Attribute{Computed: true, MarkdownDescription: "Default message notification level."},
			"explicit_content_filter":       schema.Int64Attribute{Computed: true, MarkdownDescription: "Explicit content filter level."},
			"preferred_locale":              schema.StringAttribute{Computed: true, MarkdownDescription: "Preferred locale of a community guild."},
			"premium_tier":                  schema.Int64Attribute{Computed: true, MarkdownDescription: "Server boost (premium) tier."},
			"premium_subscription_count":    schema.Int64Attribute{Computed: true, MarkdownDescription: "Number of boosts the guild has."},
		},
	}
}

func (d *guildDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*conns.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data",
			fmt.Sprintf("expected *conns.Client, got %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *guildDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data guildDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	g, err := GetGuild(ctx, d.client, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("reading Discord guild", data.ID.ValueString(), err))
		return
	}

	data.Name = types.StringValue(g.Name)
	data.Description = types.StringValue(g.Description)
	data.OwnerID = types.StringValue(g.OwnerID)
	data.Icon = types.StringValue(g.Icon)
	data.Splash = types.StringValue(g.Splash)
	data.Banner = types.StringValue(g.Banner)
	data.AFKChannelID = types.StringValue(g.AFKChannelID)
	data.AFKTimeout = types.Int64Value(g.AFKTimeout)
	data.VerificationLevel = types.Int64Value(g.VerificationLevel)
	data.DefaultMessageNotifications = types.Int64Value(g.DefaultMessageNotifications)
	data.ExplicitContentFilter = types.Int64Value(g.ExplicitContentFilter)
	data.PreferredLocale = types.StringValue(g.PreferredLocale)
	data.PremiumTier = types.Int64Value(g.PremiumTier)
	data.PremiumSubscriptionCount = types.Int64Value(g.PremiumSubscriptionCount)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
