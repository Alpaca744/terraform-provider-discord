package webhook

import (
	"context"
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
	_ resource.Resource                = (*webhookResource)(nil)
	_ resource.ResourceWithConfigure   = (*webhookResource)(nil)
	_ resource.ResourceWithImportState = (*webhookResource)(nil)
)

// NewWebhookResource is the resource factory registered with the provider.
func NewWebhookResource() resource.Resource {
	return &webhookResource{}
}

type webhookResource struct {
	client *conns.Client
}

type webhookResourceModel struct {
	ID        types.String `tfsdk:"id"`
	ChannelID types.String `tfsdk:"channel_id"`
	Name      types.String `tfsdk:"name"`
	Avatar    types.String `tfsdk:"avatar"`
	GuildID   types.String `tfsdk:"guild_id"`
	Token     types.String `tfsdk:"token"`
	URL       types.String `tfsdk:"url"`
}

func (r *webhookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

func (r *webhookResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Discord channel webhook. Requires the bot to have `MANAGE_WEBHOOKS`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"channel_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Snowflake ID of the channel the webhook posts to. May be changed to move the webhook.",
				Validators:          []validator.String{validators.Snowflake()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Webhook name (1-80 characters). Must not contain the substring `clyde`.",
				Validators:          []validator.String{stringvalidator.LengthBetween(1, 80)},
			},
			"avatar": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Webhook avatar as a Discord image data URI (`data:image/png;base64,...`).",
			},
			"guild_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Snowflake ID of the guild the webhook belongs to.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"token": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The secure token of the webhook.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"url": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The full webhook URL (includes the token).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *webhookResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *webhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan webhookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := createBody{Name: plan.Name.ValueString(), Avatar: strPtr(plan.Avatar)}
	wh, err := create(ctx, r.client, plan.ChannelID.ValueString(), body, "")
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("creating Discord webhook", plan.ChannelID.ValueString(), err))
		return
	}

	r.apply(&plan, wh)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *webhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state webhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wh, err := get(ctx, r.client, state.ID.ValueString())
	if err != nil {
		if conns.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(diagutil.APIError("reading Discord webhook", state.ID.ValueString(), err))
		return
	}

	// The token is only returned on create/get for app-owned webhooks; preserve
	// the prior value if the read omits it so the sensitive field is stable.
	prevToken := state.Token
	r.apply(&state, wh)
	if wh.Token == "" {
		state.Token = prevToken
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *webhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan webhookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := modifyBody{
		Name:      strPtr(plan.Name),
		Avatar:    strPtr(plan.Avatar),
		ChannelID: strPtr(plan.ChannelID),
	}
	wh, err := modify(ctx, r.client, plan.ID.ValueString(), body, "")
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("updating Discord webhook", plan.ID.ValueString(), err))
		return
	}

	r.apply(&plan, wh)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *webhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state webhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := deleteWebhook(ctx, r.client, state.ID.ValueString(), ""); err != nil && !conns.IsNotFound(err) {
		resp.Diagnostics.Append(diagutil.APIError("deleting Discord webhook", state.ID.ValueString(), err))
	}
}

// ImportState accepts a bare webhook ID. The token cannot be recovered on
// import unless the read returns it, so downstream uses of `token`/`url` may be
// empty after import.
func (r *webhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id := strings.TrimSpace(req.ID)
	if id == "" {
		resp.Diagnostics.AddError("Invalid import ID", "expected a webhook snowflake ID")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func (r *webhookResource) apply(m *webhookResourceModel, wh *Webhook) {
	m.ID = types.StringValue(wh.ID)
	m.ChannelID = types.StringValue(wh.ChannelID)
	m.Name = types.StringValue(wh.Name)
	m.GuildID = optStr(wh.GuildID)
	m.Avatar = optStr(wh.Avatar)
	m.Token = types.StringValue(wh.Token)
	if wh.URL != "" {
		m.URL = types.StringValue(wh.URL)
	} else if wh.Token != "" {
		m.URL = types.StringValue(fmt.Sprintf("https://discord.com/api/webhooks/%s/%s", wh.ID, wh.Token))
	} else {
		m.URL = types.StringNull()
	}
}

func strPtr(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	return &s
}

func optStr(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}
