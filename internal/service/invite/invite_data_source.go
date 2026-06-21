package invite

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
	_ datasource.DataSource              = (*inviteDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*inviteDataSource)(nil)
)

// NewInviteDataSource is the data source factory.
func NewInviteDataSource() datasource.DataSource { return &inviteDataSource{} }

type inviteDataSource struct{ client *conns.Client }

type inviteDataModel struct {
	Code      types.String `tfsdk:"code"`
	ChannelID types.String `tfsdk:"channel_id"`
	GuildID   types.String `tfsdk:"guild_id"`
	URL       types.String `tfsdk:"url"`
}

func (d *inviteDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_invite"
}

func (d *inviteDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Discord invite by code.",
		Attributes: map[string]schema.Attribute{
			"code": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The invite code to read.",
			},
			"channel_id": schema.StringAttribute{Computed: true, MarkdownDescription: "Snowflake ID of the invite's channel."},
			"guild_id":   schema.StringAttribute{Computed: true, MarkdownDescription: "Snowflake ID of the invite's guild."},
			"url":        schema.StringAttribute{Computed: true, MarkdownDescription: "The full invite URL."},
		},
	}
}

func (d *inviteDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *inviteDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data inviteDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inv, err := get(ctx, d.client, data.Code.ValueString())
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("reading Discord invite", data.Code.ValueString(), err))
		return
	}

	data.ChannelID = optStr(inv.Channel.ID)
	data.GuildID = optStr(inv.Guild.ID)
	data.URL = types.StringValue("https://discord.gg/" + inv.Code)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
