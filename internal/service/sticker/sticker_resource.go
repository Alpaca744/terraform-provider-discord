package sticker

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
	"github.com/alpaca744/terraform-provider-discord/internal/diagutil"
	"github.com/alpaca744/terraform-provider-discord/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*stickerResource)(nil)
	_ resource.ResourceWithConfigure   = (*stickerResource)(nil)
	_ resource.ResourceWithImportState = (*stickerResource)(nil)
)

// NewStickerResource is the resource factory.
func NewStickerResource() resource.Resource {
	return &stickerResource{}
}

type stickerResource struct {
	client *conns.Client
}

type stickerModel struct {
	ID                types.String `tfsdk:"id"`
	GuildID           types.String `tfsdk:"guild_id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	Tags              types.String `tfsdk:"tags"`
	Format            types.String `tfsdk:"format"`
	FileContentBase64 types.String `tfsdk:"file_content_base64"`
	FormatType        types.Int64  `tfsdk:"format_type"`
}

func (r *stickerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_guild_sticker"
}

func (r *stickerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	forceNew := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a custom Discord guild sticker. Creation uploads the sticker file via multipart; edits change name/description/tags only.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"guild_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Snowflake ID of the guild. Changing this forces a new sticker.",
				Validators:          []validator.String{validators.Snowflake()},
				PlanModifiers:       forceNew,
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Sticker name (2-30 characters).",
				Validators:          []validator.String{stringvalidator.LengthBetween(2, 30)},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Sticker description (2-100 characters, or empty).",
			},
			"tags": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Autocomplete/suggestion tags for the sticker (max 200 characters).",
				Validators:          []validator.String{stringvalidator.LengthBetween(1, 200)},
			},
			"format": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Sticker file format: `png`, `apng`, `gif`, or `lottie`. Changing this forces a new sticker.",
				Validators:          []validator.String{stringvalidator.OneOf("png", "apng", "gif", "lottie")},
				PlanModifiers:       forceNew,
			},
			"file_content_base64": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "Base64-encoded sticker file content (use `filebase64(...)`). Write-only; changing it forces a new sticker.",
				PlanModifiers:       forceNew,
			},
			"format_type": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Discord sticker format type (1 PNG, 2 APNG, 3 LOTTIE, 4 GIF).",
			},
		},
	}
}

func (r *stickerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *stickerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan stickerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	content, err := base64.StdEncoding.DecodeString(plan.FileContentBase64.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("file_content_base64"),
			"Invalid base64 content", err.Error())
		return
	}

	st, err := create(ctx, r.client, plan.GuildID.ValueString(),
		plan.Name.ValueString(), plan.Description.ValueString(), plan.Tags.ValueString(),
		plan.Format.ValueString(), content, "")
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("creating Discord sticker", plan.GuildID.ValueString(), err))
		return
	}
	r.apply(&plan, st)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *stickerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state stickerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	st, err := get(ctx, r.client, state.GuildID.ValueString(), state.ID.ValueString())
	if err != nil {
		if conns.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(diagutil.APIError("reading Discord sticker", state.GuildID.ValueString(), err))
		return
	}
	r.apply(&state, st)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *stickerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan stickerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	desc := plan.Description.ValueString()
	tags := plan.Tags.ValueString()
	st, err := modify(ctx, r.client, plan.GuildID.ValueString(), plan.ID.ValueString(),
		modifyBody{Name: &name, Description: &desc, Tags: &tags}, "")
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("updating Discord sticker", plan.GuildID.ValueString(), err))
		return
	}
	r.apply(&plan, st)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *stickerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state stickerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := deleteSticker(ctx, r.client, state.GuildID.ValueString(), state.ID.ValueString(), ""); err != nil && !conns.IsNotFound(err) {
		resp.Diagnostics.Append(diagutil.APIError("deleting Discord sticker", state.GuildID.ValueString(), err))
	}
}

// ImportState accepts the "guild_id:sticker_id" format. The file content and
// format cannot be recovered on import.
func (r *stickerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("expected import ID in the form \"guild_id:sticker_id\", got %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("guild_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func (r *stickerResource) apply(m *stickerModel, st *Sticker) {
	m.ID = types.StringValue(st.ID)
	m.Name = types.StringValue(st.Name)
	m.Tags = types.StringValue(st.Tags)
	m.FormatType = types.Int64Value(st.FormatType)
	if st.GuildID != "" {
		m.GuildID = types.StringValue(st.GuildID)
	}
	if st.Description == "" {
		m.Description = types.StringNull()
	} else {
		m.Description = types.StringValue(st.Description)
	}
	// file_content_base64 and format are write-only; left as configured in state.
}
