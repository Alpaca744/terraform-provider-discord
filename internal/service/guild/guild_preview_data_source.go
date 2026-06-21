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
	_ datasource.DataSource              = (*guildPreviewDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*guildPreviewDataSource)(nil)
)

// NewGuildPreviewDataSource is the data source factory.
func NewGuildPreviewDataSource() datasource.DataSource {
	return &guildPreviewDataSource{}
}

type guildPreviewDataSource struct {
	client *conns.Client
}

type guildPreviewModel struct {
	ID                       types.String `tfsdk:"id"`
	Name                     types.String `tfsdk:"name"`
	Description              types.String `tfsdk:"description"`
	Icon                     types.String `tfsdk:"icon"`
	Features                 types.List   `tfsdk:"features"`
	ApproximateMemberCount   types.Int64  `tfsdk:"approximate_member_count"`
	ApproximatePresenceCount types.Int64  `tfsdk:"approximate_presence_count"`
}

func (d *guildPreviewDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_guild_preview"
}

func (d *guildPreviewDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads the public preview of a discoverable Discord guild.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Snowflake ID of the guild.",
				Validators:          []validator.String{validators.Snowflake()},
			},
			"name":                       schema.StringAttribute{Computed: true, MarkdownDescription: "Guild name."},
			"description":                schema.StringAttribute{Computed: true, MarkdownDescription: "Guild description."},
			"icon":                       schema.StringAttribute{Computed: true, MarkdownDescription: "Guild icon hash."},
			"features":                   schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Enabled guild feature flags."},
			"approximate_member_count":   schema.Int64Attribute{Computed: true, MarkdownDescription: "Approximate number of members."},
			"approximate_presence_count": schema.Int64Attribute{Computed: true, MarkdownDescription: "Approximate number of online members."},
		},
	}
}

func (d *guildPreviewDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *guildPreviewDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data guildPreviewModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	p, err := GetGuildPreview(ctx, d.client, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("reading Discord guild preview", data.ID.ValueString(), err))
		return
	}

	features, diags := types.ListValueFrom(ctx, types.StringType, p.Features)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Name = types.StringValue(p.Name)
	data.Description = optStr(p.Description)
	data.Icon = optStr(p.Icon)
	data.Features = features
	data.ApproximateMemberCount = types.Int64Value(p.ApproximateMemberCount)
	data.ApproximatePresenceCount = types.Int64Value(p.ApproximatePresenceCount)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
