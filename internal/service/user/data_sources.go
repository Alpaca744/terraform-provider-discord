package user

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

// userModel is shared by both user data sources.
type userModel struct {
	ID            types.String `tfsdk:"id"`
	Username      types.String `tfsdk:"username"`
	GlobalName    types.String `tfsdk:"global_name"`
	Discriminator types.String `tfsdk:"discriminator"`
	Avatar        types.String `tfsdk:"avatar"`
	Bot           types.Bool   `tfsdk:"bot"`
}

func userAttributes(idRequired bool) map[string]schema.Attribute {
	id := schema.StringAttribute{Computed: true, MarkdownDescription: "User snowflake ID."}
	if idRequired {
		id = schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Snowflake ID of the user to read.",
			Validators:          []validator.String{validators.Snowflake()},
		}
	}
	return map[string]schema.Attribute{
		"id":            id,
		"username":      schema.StringAttribute{Computed: true, MarkdownDescription: "The user's username."},
		"global_name":   schema.StringAttribute{Computed: true, MarkdownDescription: "The user's display (global) name."},
		"discriminator": schema.StringAttribute{Computed: true, MarkdownDescription: "The user's legacy discriminator."},
		"avatar":        schema.StringAttribute{Computed: true, MarkdownDescription: "The user's avatar hash."},
		"bot":           schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the user is a bot."},
	}
}

func applyUser(m *userModel, u *User) {
	m.ID = types.StringValue(u.ID)
	m.Username = types.StringValue(u.Username)
	m.GlobalName = optStr(u.GlobalName)
	m.Discriminator = types.StringValue(u.Discriminator)
	m.Avatar = optStr(u.Avatar)
	m.Bot = types.BoolValue(u.Bot)
}

func optStr(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

func clientFromProvider(providerData any, diags interface{ AddError(string, string) }) *conns.Client {
	if providerData == nil {
		return nil
	}
	client, ok := providerData.(*conns.Client)
	if !ok {
		diags.AddError("Unexpected provider data", fmt.Sprintf("expected *conns.Client, got %T", providerData))
		return nil
	}
	return client
}

// ---- discord_current_user ----

var (
	_ datasource.DataSource              = (*currentUserDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*currentUserDataSource)(nil)
)

// NewCurrentUserDataSource is the data source factory.
func NewCurrentUserDataSource() datasource.DataSource {
	return &currentUserDataSource{}
}

type currentUserDataSource struct {
	client *conns.Client
}

func (d *currentUserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_current_user"
}

func (d *currentUserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads the Discord user associated with the configured token.",
		Attributes:          userAttributes(false),
	}
}

func (d *currentUserDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProvider(req.ProviderData, &resp.Diagnostics)
}

func (d *currentUserDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	u, err := GetCurrentUser(ctx, d.client)
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("reading current Discord user", "", err))
		return
	}
	var state userModel
	applyUser(&state, u)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ---- discord_user ----

var (
	_ datasource.DataSource              = (*userDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*userDataSource)(nil)
)

// NewUserDataSource is the data source factory.
func NewUserDataSource() datasource.DataSource {
	return &userDataSource{}
}

type userDataSource struct {
	client *conns.Client
}

func (d *userDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *userDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Discord user by ID.",
		Attributes:          userAttributes(true),
	}
}

func (d *userDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProvider(req.ProviderData, &resp.Diagnostics)
}

func (d *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data userModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	u, err := GetUser(ctx, d.client, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("reading Discord user", data.ID.ValueString(), err))
		return
	}
	applyUser(&data, u)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
