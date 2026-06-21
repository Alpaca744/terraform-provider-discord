package emoji

import (
	"context"
	"fmt"
	"strings"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
	"github.com/alpaca744/terraform-provider-discord/internal/diagutil"
	"github.com/alpaca744/terraform-provider-discord/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*emojiResource)(nil)
	_ resource.ResourceWithConfigure   = (*emojiResource)(nil)
	_ resource.ResourceWithImportState = (*emojiResource)(nil)
)

// NewEmojiResource is the resource factory.
func NewEmojiResource() resource.Resource {
	return &emojiResource{}
}

type emojiResource struct {
	client *conns.Client
}

type emojiModel struct {
	ID       types.String `tfsdk:"id"`
	GuildID  types.String `tfsdk:"guild_id"`
	Name     types.String `tfsdk:"name"`
	Image    types.String `tfsdk:"image"`
	Roles    types.Set    `tfsdk:"roles"`
	Animated types.Bool   `tfsdk:"animated"`
}

func (r *emojiResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_guild_emoji"
}

func (r *emojiResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a custom Discord guild emoji. Requires `CREATE_GUILD_EXPRESSIONS` / `MANAGE_GUILD_EXPRESSIONS`. Note: Discord documents that emoji routes do not follow normal rate-limit conventions.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"guild_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Snowflake ID of the guild. Changing this forces a new emoji.",
				Validators:          []validator.String{validators.Snowflake()},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Emoji name (2-32 characters, alphanumerics and underscores).",
			},
			"image": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Emoji image as a data URI (`data:image/png;base64,...`). Discord does not return the image on read, so it is write-only; changing it forces a new emoji.",
				Validators:          []validator.String{validators.ImageDataURI()},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"roles": schema.SetAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Snowflake IDs of roles allowed to use this emoji (empty means everyone).",
				Validators:          []validator.Set{setvalidator.ValueStringsAre(validators.Snowflake())},
			},
			"animated": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the emoji is animated.",
			},
		},
	}
}

func (r *emojiResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *emojiResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan emojiModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	roles, diags := r.roles(ctx, plan.Roles)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	em, err := create(ctx, r.client, plan.GuildID.ValueString(),
		createBody{Name: plan.Name.ValueString(), Image: plan.Image.ValueString(), Roles: roles}, "")
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("creating Discord emoji", plan.GuildID.ValueString(), err))
		return
	}
	resp.Diagnostics.Append(r.apply(ctx, &plan, em)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *emojiResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state emojiModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	em, err := get(ctx, r.client, state.GuildID.ValueString(), state.ID.ValueString())
	if err != nil {
		if conns.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(diagutil.APIError("reading Discord emoji", state.GuildID.ValueString(), err))
		return
	}
	resp.Diagnostics.Append(r.apply(ctx, &state, em)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *emojiResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan emojiModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	roles, diags := r.roles(ctx, plan.Roles)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	em, err := modify(ctx, r.client, plan.GuildID.ValueString(), plan.ID.ValueString(),
		modifyBody{Name: &name, Roles: roles}, "")
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("updating Discord emoji", plan.GuildID.ValueString(), err))
		return
	}
	resp.Diagnostics.Append(r.apply(ctx, &plan, em)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *emojiResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state emojiModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := deleteEmoji(ctx, r.client, state.GuildID.ValueString(), state.ID.ValueString(), ""); err != nil && !conns.IsNotFound(err) {
		resp.Diagnostics.Append(diagutil.APIError("deleting Discord emoji", state.GuildID.ValueString(), err))
	}
}

// ImportState accepts the "guild_id:emoji_id" format. The image cannot be
// recovered on import (Discord does not return it).
func (r *emojiResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("expected import ID in the form \"guild_id:emoji_id\", got %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("guild_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func (r *emojiResource) roles(ctx context.Context, set types.Set) ([]string, diag.Diagnostics) {
	if set.IsNull() || set.IsUnknown() {
		return nil, nil
	}
	var roles []string
	d := set.ElementsAs(ctx, &roles, false)
	return roles, d
}

func (r *emojiResource) apply(ctx context.Context, m *emojiModel, em *Emoji) diag.Diagnostics {
	var diags diag.Diagnostics
	m.ID = types.StringValue(em.ID)
	m.Name = types.StringValue(em.Name)
	m.Animated = types.BoolValue(em.Animated)
	// image is write-only; left as configured in state.

	if len(em.Roles) == 0 {
		// Discord cannot distinguish a null roles set from an empty one: both mean
		// "every member may use the emoji". Preserve the value already in the model
		// (the plan on create/update, prior state on read) so a configured empty set
		// does not flip to null and trigger an "inconsistent result after apply"
		// error. Only fall back to null when the model has no concrete value yet
		// (e.g. an unknown value, or a fresh import that has not set roles).
		if m.Roles.IsUnknown() {
			m.Roles = types.SetNull(types.StringType)
		}
		return diags
	}
	set, d := types.SetValueFrom(ctx, types.StringType, em.Roles)
	diags.Append(d...)
	m.Roles = set
	return diags
}
