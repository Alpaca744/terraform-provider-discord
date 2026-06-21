// Package voice implements the discord_voice_regions data source.
package voice

import (
	"context"
	"fmt"
	"net/http"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
	"github.com/alpaca744/terraform-provider-discord/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Region mirrors the Discord voice region object.
type Region struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Optimal    bool   `json:"optimal"`
	Deprecated bool   `json:"deprecated"`
	Custom     bool   `json:"custom"`
}

func getRegions(ctx context.Context, c *conns.Client) ([]Region, error) {
	var out []Region
	err := c.Do(ctx, "reading Discord voice regions", http.MethodGet,
		"/voice/regions", conns.RequestOptions{Out: &out})
	if err != nil {
		return nil, err
	}
	return out, nil
}

var (
	_ datasource.DataSource              = (*regionsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*regionsDataSource)(nil)
)

// NewVoiceRegionsDataSource is the data source factory.
func NewVoiceRegionsDataSource() datasource.DataSource {
	return &regionsDataSource{}
}

type regionsDataSource struct {
	client *conns.Client
}

type regionsModel struct {
	Regions []regionModel `tfsdk:"regions"`
}

type regionModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Optimal    types.Bool   `tfsdk:"optimal"`
	Deprecated types.Bool   `tfsdk:"deprecated"`
	Custom     types.Bool   `tfsdk:"custom"`
}

func (d *regionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_voice_regions"
}

func (d *regionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists the Discord voice regions usable for voice and stage channels.",
		Attributes: map[string]schema.Attribute{
			"regions": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The available voice regions.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":         schema.StringAttribute{Computed: true, MarkdownDescription: "Unique region ID."},
						"name":       schema.StringAttribute{Computed: true, MarkdownDescription: "Human-readable region name."},
						"optimal":    schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this is the closest region to the current user."},
						"deprecated": schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the region is deprecated."},
						"custom":     schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the region is custom (e.g. for events)."},
					},
				},
			},
		},
	}
}

func (d *regionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *regionsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	regions, err := getRegions(ctx, d.client)
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("reading Discord voice regions", "", err))
		return
	}

	state := regionsModel{Regions: make([]regionModel, 0, len(regions))}
	for _, rg := range regions {
		state.Regions = append(state.Regions, regionModel{
			ID:         types.StringValue(rg.ID),
			Name:       types.StringValue(rg.Name),
			Optimal:    types.BoolValue(rg.Optimal),
			Deprecated: types.BoolValue(rg.Deprecated),
			Custom:     types.BoolValue(rg.Custom),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
