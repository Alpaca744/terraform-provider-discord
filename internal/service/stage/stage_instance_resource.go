package stage

import (
	"context"
	"fmt"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
	"github.com/alpaca744/terraform-provider-discord/internal/diagutil"
	"github.com/alpaca744/terraform-provider-discord/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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
	_ resource.Resource                = (*stageInstanceResource)(nil)
	_ resource.ResourceWithConfigure   = (*stageInstanceResource)(nil)
	_ resource.ResourceWithImportState = (*stageInstanceResource)(nil)
)

// NewStageInstanceResource is the resource factory.
func NewStageInstanceResource() resource.Resource {
	return &stageInstanceResource{}
}

type stageInstanceResource struct {
	client *conns.Client
}

type stageInstanceModel struct {
	ID           types.String `tfsdk:"id"`
	ChannelID    types.String `tfsdk:"channel_id"`
	GuildID      types.String `tfsdk:"guild_id"`
	Topic        types.String `tfsdk:"topic"`
	PrivacyLevel types.Int64  `tfsdk:"privacy_level"`
}

func (r *stageInstanceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stage_instance"
}

func (r *stageInstanceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Discord stage instance (a live stage in a stage channel). Keyed by the stage channel ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The stage instance snowflake ID.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"channel_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Snowflake ID of the stage channel. Changing this forces a new stage instance.",
				Validators:          []validator.String{validators.Snowflake()},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"guild_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Snowflake ID of the guild.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"topic": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The topic of the stage instance (1-120 characters).",
				Validators:          []validator.String{stringvalidator.LengthBetween(1, 120)},
			},
			"privacy_level": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Privacy level: 2 GUILD_ONLY (default). 1 (PUBLIC) is deprecated.",
				Validators:          []validator.Int64{int64validator.OneOf(1, 2)},
			},
		},
	}
}

func (r *stageInstanceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *stageInstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan stageInstanceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := createBody{ChannelID: plan.ChannelID.ValueString(), Topic: plan.Topic.ValueString()}
	if !plan.PrivacyLevel.IsNull() && !plan.PrivacyLevel.IsUnknown() {
		v := plan.PrivacyLevel.ValueInt64()
		body.PrivacyLevel = &v
	}

	si, err := create(ctx, r.client, body, "")
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("creating Discord stage instance", plan.ChannelID.ValueString(), err))
		return
	}
	r.apply(&plan, si)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *stageInstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state stageInstanceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	si, err := get(ctx, r.client, state.ChannelID.ValueString())
	if err != nil {
		if conns.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(diagutil.APIError("reading Discord stage instance", state.ChannelID.ValueString(), err))
		return
	}
	r.apply(&state, si)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *stageInstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan stageInstanceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	topic := plan.Topic.ValueString()
	body := modifyBody{Topic: &topic}
	if !plan.PrivacyLevel.IsNull() && !plan.PrivacyLevel.IsUnknown() {
		v := plan.PrivacyLevel.ValueInt64()
		body.PrivacyLevel = &v
	}

	si, err := modify(ctx, r.client, plan.ChannelID.ValueString(), body, "")
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("updating Discord stage instance", plan.ChannelID.ValueString(), err))
		return
	}
	r.apply(&plan, si)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *stageInstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state stageInstanceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := deleteInstance(ctx, r.client, state.ChannelID.ValueString(), ""); err != nil && !conns.IsNotFound(err) {
		resp.Diagnostics.Append(diagutil.APIError("deleting Discord stage instance", state.ChannelID.ValueString(), err))
	}
}

// ImportState accepts the stage channel ID.
func (r *stageInstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("channel_id"), req.ID)...)
}

func (r *stageInstanceResource) apply(m *stageInstanceModel, si *StageInstance) {
	m.ID = types.StringValue(si.ID)
	m.ChannelID = types.StringValue(si.ChannelID)
	m.GuildID = types.StringValue(si.GuildID)
	m.Topic = types.StringValue(si.Topic)
	m.PrivacyLevel = types.Int64Value(si.PrivacyLevel)
}
