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
	_ datasource.DataSource              = (*channelsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*channelsDataSource)(nil)
)

// NewChannelsDataSource is the data source factory.
func NewChannelsDataSource() datasource.DataSource {
	return &channelsDataSource{}
}

type channelsDataSource struct {
	client *conns.Client
}

type channelsModel struct {
	GuildID  types.String        `tfsdk:"guild_id"`
	Channels []channelEntryModel `tfsdk:"channels"`
}

type channelEntryModel struct {
	ID       types.String `tfsdk:"id"`
	Type     types.Int64  `tfsdk:"type"`
	Name     types.String `tfsdk:"name"`
	Topic    types.String `tfsdk:"topic"`
	Position types.Int64  `tfsdk:"position"`
	NSFW     types.Bool   `tfsdk:"nsfw"`
	ParentID types.String `tfsdk:"parent_id"`
}

func (d *channelsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_channels"
}

func (d *channelsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all channels in a Discord guild.",
		Attributes: map[string]schema.Attribute{
			"guild_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Snowflake ID of the guild.",
				Validators:          []validator.String{validators.Snowflake()},
			},
			"channels": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The guild's channels.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":        schema.StringAttribute{Computed: true, MarkdownDescription: "Channel snowflake ID."},
						"type":      schema.Int64Attribute{Computed: true, MarkdownDescription: "Discord channel type."},
						"name":      schema.StringAttribute{Computed: true, MarkdownDescription: "Channel name."},
						"topic":     schema.StringAttribute{Computed: true, MarkdownDescription: "Channel topic."},
						"position":  schema.Int64Attribute{Computed: true, MarkdownDescription: "Channel position."},
						"nsfw":      schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the channel is NSFW."},
						"parent_id": schema.StringAttribute{Computed: true, MarkdownDescription: "Parent category channel ID."},
					},
				},
			},
		},
	}
}

func (d *channelsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *channelsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data channelsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	channels, err := ListChannels(ctx, d.client, data.GuildID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("listing Discord channels", data.GuildID.ValueString(), err))
		return
	}

	data.Channels = make([]channelEntryModel, 0, len(channels))
	for _, ch := range channels {
		data.Channels = append(data.Channels, channelEntryModel{
			ID:       types.StringValue(ch.ID),
			Type:     types.Int64Value(ch.Type),
			Name:     types.StringValue(ch.Name),
			Topic:    optStr(ch.Topic),
			Position: types.Int64Value(ch.Position),
			NSFW:     types.BoolValue(ch.NSFW),
			ParentID: optStr(ch.ParentID),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
