package guild

import (
	"context"
	"fmt"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
	"github.com/alpaca744/terraform-provider-discord/internal/diagutil"
	"github.com/alpaca744/terraform-provider-discord/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = (*widgetResource)(nil)
	_ resource.ResourceWithConfigure   = (*widgetResource)(nil)
	_ resource.ResourceWithImportState = (*widgetResource)(nil)
)

// NewWidgetResource is the resource factory.
func NewWidgetResource() resource.Resource {
	return &widgetResource{}
}

// widgetResource manages a guild's widget settings. The settings object always
// exists, so this resource only reads and modifies it; Delete is a no-op.
type widgetResource struct {
	client *conns.Client
}

type widgetModel struct {
	GuildID   types.String `tfsdk:"guild_id"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	ChannelID types.String `tfsdk:"channel_id"`
}

func (r *widgetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_guild_widget"
}

func (r *widgetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Discord guild's widget settings. The settings object always exists; deleting this resource leaves the widget configuration unchanged.",
		Attributes: map[string]schema.Attribute{
			"guild_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Snowflake ID of the guild. Changing this forces a new resource.",
				Validators:          []validator.String{validators.Snowflake()},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the widget is enabled.",
			},
			"channel_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Snowflake ID of the channel the widget invite points to (null for none).",
				Validators:          []validator.String{validators.Snowflake()},
			},
		},
	}
}

func (r *widgetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*conns.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data",
			fmt.Sprintf("expected *conns.Client, got %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *widgetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan widgetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ws, err := ModifyWidgetSettings(ctx, r.client, plan.GuildID.ValueString(), r.body(plan), "")
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("applying Discord guild widget", plan.GuildID.ValueString(), err))
		return
	}
	r.apply(&plan, ws)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *widgetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state widgetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ws, err := GetWidgetSettings(ctx, r.client, state.GuildID.ValueString())
	if err != nil {
		if conns.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(diagutil.APIError("reading Discord guild widget", state.GuildID.ValueString(), err))
		return
	}
	r.apply(&state, ws)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *widgetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan widgetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ws, err := ModifyWidgetSettings(ctx, r.client, plan.GuildID.ValueString(), r.body(plan), "")
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("updating Discord guild widget", plan.GuildID.ValueString(), err))
		return
	}
	r.apply(&plan, ws)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete is a no-op: the widget settings object is not owned by this resource.
func (r *widgetResource) Delete(ctx context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	tflog.Info(ctx, "discord_guild_widget deletion is a no-op; widget settings are left unchanged")
}

// ImportState accepts a bare guild ID.
func (r *widgetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("guild_id"), req.ID)...)
}

func (r *widgetResource) body(m widgetModel) WidgetSettingsBody {
	return WidgetSettingsBody{
		Enabled:   boolPtr(m.Enabled),
		ChannelID: strPtr(m.ChannelID),
	}
}

func (r *widgetResource) apply(m *widgetModel, ws *WidgetSettings) {
	m.Enabled = types.BoolValue(ws.Enabled)
	m.ChannelID = ptrStr(ws.ChannelID)
}

func ptrStr(s *string) types.String {
	if s == nil || *s == "" {
		return types.StringNull()
	}
	return types.StringValue(*s)
}
