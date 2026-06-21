package channel

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
	_ datasource.DataSource              = (*channelDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*channelDataSource)(nil)
)

// NewChannelDataSource is the data source factory.
func NewChannelDataSource() datasource.DataSource { return &channelDataSource{} }

type channelDataSource struct{ client *conns.Client }

type channelDataModel struct {
	ID       types.String `tfsdk:"id"`
	GuildID  types.String `tfsdk:"guild_id"`
	Type     types.Int64  `tfsdk:"type"`
	Name     types.String `tfsdk:"name"`
	Topic    types.String `tfsdk:"topic"`
	Position types.Int64  `tfsdk:"position"`
	NSFW     types.Bool   `tfsdk:"nsfw"`
	ParentID types.String `tfsdk:"parent_id"`
}

func (d *channelDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_channel"
}

func (d *channelDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a single Discord channel by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Snowflake ID of the channel.",
				Validators:          []validator.String{validators.Snowflake()},
			},
			"guild_id":  schema.StringAttribute{Computed: true, MarkdownDescription: "Snowflake ID of the guild."},
			"type":      schema.Int64Attribute{Computed: true, MarkdownDescription: "Discord channel type."},
			"name":      schema.StringAttribute{Computed: true, MarkdownDescription: "Channel name."},
			"topic":     schema.StringAttribute{Computed: true, MarkdownDescription: "Channel topic."},
			"position":  schema.Int64Attribute{Computed: true, MarkdownDescription: "Channel position."},
			"nsfw":      schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the channel is NSFW."},
			"parent_id": schema.StringAttribute{Computed: true, MarkdownDescription: "Parent category channel ID."},
		},
	}
}

func (d *channelDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *channelDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data channelDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ch, err := get(ctx, d.client, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("reading Discord channel", data.ID.ValueString(), err))
		return
	}

	data.GuildID = optStr(ch.GuildID)
	data.Type = types.Int64Value(ch.Type)
	data.Name = types.StringValue(ch.Name)
	data.Topic = optStr(ch.Topic)
	data.Position = types.Int64Value(ch.Position)
	data.NSFW = types.BoolValue(ch.NSFW)
	data.ParentID = optStr(ch.ParentID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
