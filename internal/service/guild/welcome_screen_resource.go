package guild

import (
	"context"
	"fmt"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
	"github.com/alpaca744/terraform-provider-discord/internal/diagutil"
	"github.com/alpaca744/terraform-provider-discord/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = (*welcomeScreenResource)(nil)
	_ resource.ResourceWithConfigure   = (*welcomeScreenResource)(nil)
	_ resource.ResourceWithImportState = (*welcomeScreenResource)(nil)
)

// NewWelcomeScreenResource is the resource factory.
func NewWelcomeScreenResource() resource.Resource {
	return &welcomeScreenResource{}
}

// welcomeScreenResource manages a community guild's welcome screen. The object
// always exists for community guilds, so Delete is a no-op.
type welcomeScreenResource struct {
	client *conns.Client
}

type welcomeScreenModel struct {
	GuildID         types.String          `tfsdk:"guild_id"`
	Enabled         types.Bool            `tfsdk:"enabled"`
	Description     types.String          `tfsdk:"description"`
	WelcomeChannels []welcomeChannelModel `tfsdk:"welcome_channels"`
}

type welcomeChannelModel struct {
	ChannelID   types.String `tfsdk:"channel_id"`
	Description types.String `tfsdk:"description"`
	EmojiID     types.String `tfsdk:"emoji_id"`
	EmojiName   types.String `tfsdk:"emoji_name"`
}

func (r *welcomeScreenResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_guild_welcome_screen"
}

func (r *welcomeScreenResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a community guild's welcome screen. The object always exists for community guilds; deleting this resource leaves it unchanged.",
		Attributes: map[string]schema.Attribute{
			"guild_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Snowflake ID of the guild. Changing this forces a new resource.",
				Validators:          []validator.String{validators.Snowflake()},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether the welcome screen is enabled. Write-only: Discord does not return this on read.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The server description shown on the welcome screen.",
			},
			"welcome_channels": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Up to 5 channels shown on the welcome screen.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"channel_id": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Snowflake ID of the channel.",
							Validators:          []validator.String{validators.Snowflake()},
						},
						"description": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Description shown for the channel.",
						},
						"emoji_id": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Snowflake ID of a custom emoji shown for the channel.",
							Validators:          []validator.String{validators.Snowflake()},
						},
						"emoji_name": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Unicode emoji or custom emoji name shown for the channel.",
						},
					},
				},
			},
		},
	}
}

func (r *welcomeScreenResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *welcomeScreenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.write(ctx, req.Plan, &resp.State, &resp.Diagnostics)
}

func (r *welcomeScreenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.write(ctx, req.Plan, &resp.State, &resp.Diagnostics)
}

func (r *welcomeScreenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state welcomeScreenModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ws, err := GetWelcomeScreen(ctx, r.client, state.GuildID.ValueString())
	if err != nil {
		if conns.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(diagutil.APIError("reading Discord welcome screen", state.GuildID.ValueString(), err))
		return
	}
	r.apply(&state, ws)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete is a no-op: the welcome screen object is not owned by this resource.
func (r *welcomeScreenResource) Delete(ctx context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	tflog.Info(ctx, "discord_guild_welcome_screen deletion is a no-op; the welcome screen is left unchanged")
}

// ImportState accepts a bare guild ID.
func (r *welcomeScreenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("guild_id"), req.ID)...)
}

func (r *welcomeScreenResource) write(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State, diags *diag.Diagnostics) {
	var model welcomeScreenModel
	diags.Append(plan.Get(ctx, &model)...)
	if diags.HasError() {
		return
	}

	channels := make([]WelcomeChannel, 0, len(model.WelcomeChannels))
	for _, ch := range model.WelcomeChannels {
		channels = append(channels, WelcomeChannel{
			ChannelID:   ch.ChannelID.ValueString(),
			Description: ch.Description.ValueString(),
			EmojiID:     strPtr(ch.EmojiID),
			EmojiName:   strPtr(ch.EmojiName),
		})
	}
	body := WelcomeScreenBody{
		Enabled:         boolPtr(model.Enabled),
		Description:     strPtr(model.Description),
		WelcomeChannels: channels,
	}

	ws, err := ModifyWelcomeScreen(ctx, r.client, model.GuildID.ValueString(), body, "")
	if err != nil {
		diags.Append(diagutil.APIError("updating Discord welcome screen", model.GuildID.ValueString(), err))
		return
	}
	r.apply(&model, ws)
	diags.Append(state.Set(ctx, &model)...)
}

func (r *welcomeScreenResource) apply(m *welcomeScreenModel, ws *WelcomeScreen) {
	if ws.Description != nil && *ws.Description != "" {
		m.Description = types.StringValue(*ws.Description)
	} else {
		m.Description = types.StringNull()
	}

	channels := make([]welcomeChannelModel, 0, len(ws.WelcomeChannels))
	for _, ch := range ws.WelcomeChannels {
		channels = append(channels, welcomeChannelModel{
			ChannelID:   types.StringValue(ch.ChannelID),
			Description: types.StringValue(ch.Description),
			EmojiID:     ptrStr(ch.EmojiID),
			EmojiName:   ptrStr(ch.EmojiName),
		})
	}
	if len(channels) == 0 {
		m.WelcomeChannels = nil
	} else {
		m.WelcomeChannels = channels
	}
	// enabled is write-only; left as configured in state.
}
