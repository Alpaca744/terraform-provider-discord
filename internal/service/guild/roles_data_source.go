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

var (
	_ datasource.DataSource              = (*rolesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*rolesDataSource)(nil)
)

// NewRolesDataSource is the data source factory.
func NewRolesDataSource() datasource.DataSource {
	return &rolesDataSource{}
}

type rolesDataSource struct {
	client *conns.Client
}

type rolesModel struct {
	GuildID types.String     `tfsdk:"guild_id"`
	Roles   []roleEntryModel `tfsdk:"roles"`
}

type roleEntryModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Color       types.Int64  `tfsdk:"color"`
	Hoist       types.Bool   `tfsdk:"hoist"`
	Position    types.Int64  `tfsdk:"position"`
	Managed     types.Bool   `tfsdk:"managed"`
	Mentionable types.Bool   `tfsdk:"mentionable"`
	Permissions types.Set    `tfsdk:"permissions"`
}

func (d *rolesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_roles"
}

func (d *rolesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all roles in a Discord guild.",
		Attributes: map[string]schema.Attribute{
			"guild_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Snowflake ID of the guild.",
				Validators:          []validator.String{validators.Snowflake()},
			},
			"roles": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The guild's roles.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.StringAttribute{Computed: true, MarkdownDescription: "Role snowflake ID."},
						"name":        schema.StringAttribute{Computed: true, MarkdownDescription: "Role name."},
						"color":       schema.Int64Attribute{Computed: true, MarkdownDescription: "Integer RGB color."},
						"hoist":       schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the role is hoisted."},
						"position":    schema.Int64Attribute{Computed: true, MarkdownDescription: "Role position."},
						"managed":     schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the role is integration-managed."},
						"mentionable": schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the role is mentionable."},
						"permissions": schema.SetAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Permission names granted to the role."},
					},
				},
			},
		},
	}
}

func (d *rolesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *rolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data rolesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	roles, err := ListRoles(ctx, d.client, data.GuildID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("listing Discord roles", data.GuildID.ValueString(), err))
		return
	}

	data.Roles = make([]roleEntryModel, 0, len(roles))
	for _, role := range roles {
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
		data.Roles = append(data.Roles, roleEntryModel{
			ID:          types.StringValue(role.ID),
			Name:        types.StringValue(role.Name),
			Color:       types.Int64Value(role.Color),
			Hoist:       types.BoolValue(role.Hoist),
			Position:    types.Int64Value(role.Position),
			Managed:     types.BoolValue(role.Managed),
			Mentionable: types.BoolValue(role.Mentionable),
			Permissions: perms,
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
