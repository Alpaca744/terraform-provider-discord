package soundboard

import (
	"context"
	"fmt"
	"strings"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
	"github.com/alpaca744/terraform-provider-discord/internal/diagutil"
	"github.com/alpaca744/terraform-provider-discord/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*soundResource)(nil)
	_ resource.ResourceWithConfigure   = (*soundResource)(nil)
	_ resource.ResourceWithImportState = (*soundResource)(nil)
)

// NewSoundResource is the resource factory.
func NewSoundResource() resource.Resource {
	return &soundResource{}
}

type soundResource struct {
	client *conns.Client
}

type soundModel struct {
	SoundID   types.String  `tfsdk:"sound_id"`
	GuildID   types.String  `tfsdk:"guild_id"`
	Name      types.String  `tfsdk:"name"`
	Sound     types.String  `tfsdk:"sound"`
	Volume    types.Float64 `tfsdk:"volume"`
	EmojiID   types.String  `tfsdk:"emoji_id"`
	EmojiName types.String  `tfsdk:"emoji_name"`
}

func (r *soundResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_soundboard_sound"
}

func (r *soundResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Discord guild soundboard sound. Requires `CREATE_GUILD_EXPRESSIONS` / `MANAGE_GUILD_EXPRESSIONS`.",
		Attributes: map[string]schema.Attribute{
			"sound_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The soundboard sound snowflake ID.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"guild_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Snowflake ID of the guild. Changing this forces a new sound.",
				Validators:          []validator.String{validators.Snowflake()},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Sound name (2-32 characters).",
			},
			"sound": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Audio as a data URI (`data:audio/mpeg;base64,...` or `data:audio/ogg;...`). Write-only; changing it forces a new sound.",
				Validators:          []validator.String{validators.AudioDataURI()},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"volume": schema.Float64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             float64default.StaticFloat64(1),
				MarkdownDescription: "Playback volume from 0 to 1 (default 1).",
				Validators:          []validator.Float64{float64validator.Between(0, 1)},
			},
			"emoji_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Snowflake ID of a custom emoji associated with the sound.",
				Validators:          []validator.String{validators.Snowflake()},
			},
			"emoji_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Unicode emoji associated with the sound.",
			},
		},
	}
}

func (r *soundResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *soundResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan soundModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := createBody{
		Name:      plan.Name.ValueString(),
		Sound:     plan.Sound.ValueString(),
		EmojiID:   strPtr(plan.EmojiID),
		EmojiName: strPtr(plan.EmojiName),
	}
	if !plan.Volume.IsNull() && !plan.Volume.IsUnknown() {
		v := plan.Volume.ValueFloat64()
		body.Volume = &v
	}

	s, err := create(ctx, r.client, plan.GuildID.ValueString(), body, "")
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("creating Discord soundboard sound", plan.GuildID.ValueString(), err))
		return
	}
	r.apply(&plan, s)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *soundResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state soundModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := get(ctx, r.client, state.GuildID.ValueString(), state.SoundID.ValueString())
	if err != nil {
		if conns.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(diagutil.APIError("reading Discord soundboard sound", state.GuildID.ValueString(), err))
		return
	}
	r.apply(&state, s)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *soundResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan soundModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	body := modifyBody{Name: &name, EmojiID: strPtr(plan.EmojiID), EmojiName: strPtr(plan.EmojiName)}
	if !plan.Volume.IsNull() && !plan.Volume.IsUnknown() {
		v := plan.Volume.ValueFloat64()
		body.Volume = &v
	}

	s, err := modify(ctx, r.client, plan.GuildID.ValueString(), plan.SoundID.ValueString(), body, "")
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("updating Discord soundboard sound", plan.GuildID.ValueString(), err))
		return
	}
	r.apply(&plan, s)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *soundResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state soundModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := deleteSound(ctx, r.client, state.GuildID.ValueString(), state.SoundID.ValueString(), ""); err != nil && !conns.IsNotFound(err) {
		resp.Diagnostics.Append(diagutil.APIError("deleting Discord soundboard sound", state.GuildID.ValueString(), err))
	}
}

// ImportState accepts the "guild_id:sound_id" format. The audio data cannot be
// recovered on import.
func (r *soundResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("expected import ID in the form \"guild_id:sound_id\", got %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("guild_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("sound_id"), parts[1])...)
}

func (r *soundResource) apply(m *soundModel, s *Sound) {
	m.SoundID = types.StringValue(s.SoundID)
	m.GuildID = types.StringValue(s.GuildID)
	m.Name = types.StringValue(s.Name)
	m.Volume = types.Float64Value(s.Volume)
	m.EmojiID = optStr(s.EmojiID)
	m.EmojiName = optStr(s.EmojiName)
	// sound (audio) is write-only; left as configured in state.
}

func strPtr(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	return &s
}

func optStr(s *string) types.String {
	if s == nil || *s == "" {
		return types.StringNull()
	}
	return types.StringValue(*s)
}
