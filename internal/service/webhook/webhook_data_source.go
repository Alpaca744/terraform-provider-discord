package webhook

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
	_ datasource.DataSource              = (*webhookDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*webhookDataSource)(nil)
)

// NewWebhookDataSource is the data source factory.
func NewWebhookDataSource() datasource.DataSource { return &webhookDataSource{} }

type webhookDataSource struct{ client *conns.Client }

type webhookDataModel struct {
	ID        types.String `tfsdk:"id"`
	Type      types.Int64  `tfsdk:"type"`
	GuildID   types.String `tfsdk:"guild_id"`
	ChannelID types.String `tfsdk:"channel_id"`
	Name      types.String `tfsdk:"name"`
	Avatar    types.String `tfsdk:"avatar"`
}

func (d *webhookDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

func (d *webhookDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a single Discord webhook by ID. The token is not exposed by this data source.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Snowflake ID of the webhook.",
				Validators:          []validator.String{validators.Snowflake()},
			},
			"type":       schema.Int64Attribute{Computed: true, MarkdownDescription: "Webhook type."},
			"guild_id":   schema.StringAttribute{Computed: true, MarkdownDescription: "Snowflake ID of the guild."},
			"channel_id": schema.StringAttribute{Computed: true, MarkdownDescription: "Snowflake ID of the channel."},
			"name":       schema.StringAttribute{Computed: true, MarkdownDescription: "Webhook name."},
			"avatar":     schema.StringAttribute{Computed: true, MarkdownDescription: "Webhook avatar hash."},
		},
	}
}

func (d *webhookDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *webhookDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data webhookDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wh, err := get(ctx, d.client, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("reading Discord webhook", data.ID.ValueString(), err))
		return
	}

	data.Type = types.Int64Value(wh.Type)
	data.GuildID = optStr(wh.GuildID)
	data.ChannelID = types.StringValue(wh.ChannelID)
	data.Name = types.StringValue(wh.Name)
	data.Avatar = optStr(wh.Avatar)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
