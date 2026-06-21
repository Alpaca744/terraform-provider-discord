package application

import (
	"context"
	"fmt"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
	"github.com/alpaca744/terraform-provider-discord/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*currentApplicationDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*currentApplicationDataSource)(nil)
)

// NewCurrentApplicationDataSource is the data source factory.
func NewCurrentApplicationDataSource() datasource.DataSource {
	return &currentApplicationDataSource{}
}

type currentApplicationDataSource struct {
	client *conns.Client
}

type currentApplicationModel struct {
	ID                             types.String `tfsdk:"id"`
	Name                           types.String `tfsdk:"name"`
	Description                    types.String `tfsdk:"description"`
	Icon                           types.String `tfsdk:"icon"`
	BotPublic                      types.Bool   `tfsdk:"bot_public"`
	BotRequireCodeGrant            types.Bool   `tfsdk:"bot_require_code_grant"`
	TermsOfServiceURL              types.String `tfsdk:"terms_of_service_url"`
	PrivacyPolicyURL               types.String `tfsdk:"privacy_policy_url"`
	CustomInstallURL               types.String `tfsdk:"custom_install_url"`
	RoleConnectionsVerificationURL types.String `tfsdk:"role_connections_verification_url"`
	InteractionsEndpointURL        types.String `tfsdk:"interactions_endpoint_url"`
	Flags                          types.Int64  `tfsdk:"flags"`
	Tags                           types.List   `tfsdk:"tags"`
	ApproximateGuildCount          types.Int64  `tfsdk:"approximate_guild_count"`
}

func (d *currentApplicationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_current_application"
}

func (d *currentApplicationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads the Discord application associated with the configured bot token.",
		Attributes: map[string]schema.Attribute{
			"id":                                schema.StringAttribute{Computed: true, MarkdownDescription: "Application snowflake ID."},
			"name":                              schema.StringAttribute{Computed: true, MarkdownDescription: "Application name."},
			"description":                       schema.StringAttribute{Computed: true, MarkdownDescription: "Application description."},
			"icon":                              schema.StringAttribute{Computed: true, MarkdownDescription: "Application icon hash."},
			"bot_public":                        schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether anyone can add the bot."},
			"bot_require_code_grant":            schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the bot requires the OAuth2 code grant to be added."},
			"terms_of_service_url":              schema.StringAttribute{Computed: true, MarkdownDescription: "Terms of service URL."},
			"privacy_policy_url":                schema.StringAttribute{Computed: true, MarkdownDescription: "Privacy policy URL."},
			"custom_install_url":                schema.StringAttribute{Computed: true, MarkdownDescription: "Custom install URL."},
			"role_connections_verification_url": schema.StringAttribute{Computed: true, MarkdownDescription: "Role connections verification URL."},
			"interactions_endpoint_url":         schema.StringAttribute{Computed: true, MarkdownDescription: "Interactions endpoint URL."},
			"flags":                             schema.Int64Attribute{Computed: true, MarkdownDescription: "Application flags bitfield."},
			"tags": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Up to 5 descriptive tags.",
			},
			"approximate_guild_count": schema.Int64Attribute{Computed: true, MarkdownDescription: "Approximate number of guilds the app is in."},
		},
	}
}

func (d *currentApplicationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *currentApplicationDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	app, err := GetCurrentApplication(ctx, d.client)
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("reading current Discord application", "", err))
		return
	}

	tags, diags := types.ListValueFrom(ctx, types.StringType, app.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := currentApplicationModel{
		ID:                             types.StringValue(app.ID),
		Name:                           types.StringValue(app.Name),
		Description:                    types.StringValue(app.Description),
		Icon:                           types.StringValue(app.Icon),
		BotPublic:                      types.BoolValue(app.BotPublic),
		BotRequireCodeGrant:            types.BoolValue(app.BotRequireCodeGrant),
		TermsOfServiceURL:              types.StringValue(app.TermsOfServiceURL),
		PrivacyPolicyURL:               types.StringValue(app.PrivacyPolicyURL),
		CustomInstallURL:               types.StringValue(app.CustomInstallURL),
		RoleConnectionsVerificationURL: types.StringValue(app.RoleConnectionsVerificationURL),
		InteractionsEndpointURL:        types.StringValue(app.InteractionsEndpointURL),
		Flags:                          types.Int64Value(app.Flags),
		Tags:                           tags,
		ApproximateGuildCount:          types.Int64Value(app.ApproximateGuildCount),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
