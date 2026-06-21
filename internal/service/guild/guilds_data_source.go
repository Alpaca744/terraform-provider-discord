package guild

import (
	"context"
	"fmt"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
	"github.com/alpaca744/terraform-provider-discord/internal/diagutil"
	"github.com/alpaca744/terraform-provider-discord/internal/discord"
	"github.com/alpaca744/terraform-provider-discord/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ---- discord_guilds ----

var (
	_ datasource.DataSource              = (*guildsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*guildsDataSource)(nil)
)

// NewGuildsDataSource is the data source factory.
func NewGuildsDataSource() datasource.DataSource { return &guildsDataSource{} }

type guildsDataSource struct{ client *conns.Client }

type guildsModel struct {
	Guilds []partialGuildModel `tfsdk:"guilds"`
}

type partialGuildModel struct {
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Owner types.Bool   `tfsdk:"owner"`
	Icon  types.String `tfsdk:"icon"`
}

func (d *guildsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_guilds"
}

func (d *guildsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists the guilds the configured bot/user is a member of.",
		Attributes: map[string]schema.Attribute{
			"guilds": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The guilds the current user is in.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":    schema.StringAttribute{Computed: true, MarkdownDescription: "Guild snowflake ID."},
						"name":  schema.StringAttribute{Computed: true, MarkdownDescription: "Guild name."},
						"owner": schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the current user owns the guild."},
						"icon":  schema.StringAttribute{Computed: true, MarkdownDescription: "Guild icon hash."},
					},
				},
			},
		},
	}
}

func (d *guildsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientOrError(req.ProviderData, &resp.Diagnostics)
}

func (d *guildsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	guilds, err := ListCurrentUserGuilds(ctx, d.client)
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("listing Discord guilds", "", err))
		return
	}
	var state guildsModel
	state.Guilds = make([]partialGuildModel, 0, len(guilds))
	for _, g := range guilds {
		state.Guilds = append(state.Guilds, partialGuildModel{
			ID:    types.StringValue(g.ID),
			Name:  types.StringValue(g.Name),
			Owner: types.BoolValue(g.Owner),
			Icon:  optStr(g.Icon),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ---- discord_role (single) ----

var (
	_ datasource.DataSource              = (*roleDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*roleDataSource)(nil)
)

// NewRoleDataSource is the data source factory.
func NewRoleDataSource() datasource.DataSource { return &roleDataSource{} }

type roleDataSource struct{ client *conns.Client }

type roleDataModel struct {
	GuildID     types.String `tfsdk:"guild_id"`
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Color       types.Int64  `tfsdk:"color"`
	Hoist       types.Bool   `tfsdk:"hoist"`
	Position    types.Int64  `tfsdk:"position"`
	Managed     types.Bool   `tfsdk:"managed"`
	Mentionable types.Bool   `tfsdk:"mentionable"`
	Permissions types.Set    `tfsdk:"permissions"`
}

func (d *roleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (d *roleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a single Discord role by guild and role ID.",
		Attributes: map[string]schema.Attribute{
			"guild_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Snowflake ID of the guild.",
				Validators:          []validator.String{validators.Snowflake()},
			},
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Snowflake ID of the role.",
				Validators:          []validator.String{validators.Snowflake()},
			},
			"name":        schema.StringAttribute{Computed: true, MarkdownDescription: "Role name."},
			"color":       schema.Int64Attribute{Computed: true, MarkdownDescription: "Integer RGB color."},
			"hoist":       schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the role is hoisted."},
			"position":    schema.Int64Attribute{Computed: true, MarkdownDescription: "Role position."},
			"managed":     schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the role is integration-managed."},
			"mentionable": schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the role is mentionable."},
			"permissions": schema.SetAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Permission names granted to the role."},
		},
	}
}

func (d *roleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientOrError(req.ProviderData, &resp.Diagnostics)
}

func (d *roleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data roleDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	role, err := GetRole(ctx, d.client, data.GuildID.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("reading Discord role", data.GuildID.ValueString(), err))
		return
	}

	names, err := discord.BitfieldToPermissions(role.Permissions)
	if err != nil {
		resp.Diagnostics.AddError("Invalid permissions from API", err.Error())
		return
	}
	perms, diags := types.SetValueFrom(ctx, types.StringType, names)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Name = types.StringValue(role.Name)
	data.Color = types.Int64Value(role.Color)
	data.Hoist = types.BoolValue(role.Hoist)
	data.Position = types.Int64Value(role.Position)
	data.Managed = types.BoolValue(role.Managed)
	data.Mentionable = types.BoolValue(role.Mentionable)
	data.Permissions = perms
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// clientOrError extracts the shared client from provider data.
func clientOrError(providerData any, diags interface{ AddError(string, string) }) *conns.Client {
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
