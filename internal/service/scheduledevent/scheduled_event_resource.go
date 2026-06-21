package scheduledevent

import (
	"context"
	"fmt"
	"strings"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
	"github.com/alpaca744/terraform-provider-discord/internal/diagutil"
	"github.com/alpaca744/terraform-provider-discord/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*eventResource)(nil)
	_ resource.ResourceWithConfigure   = (*eventResource)(nil)
	_ resource.ResourceWithImportState = (*eventResource)(nil)
)

// NewEventResource is the resource factory.
func NewEventResource() resource.Resource {
	return &eventResource{}
}

type eventResource struct {
	client *conns.Client
}

type eventModel struct {
	ID                 types.String `tfsdk:"id"`
	GuildID            types.String `tfsdk:"guild_id"`
	ChannelID          types.String `tfsdk:"channel_id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	ScheduledStartTime types.String `tfsdk:"scheduled_start_time"`
	ScheduledEndTime   types.String `tfsdk:"scheduled_end_time"`
	PrivacyLevel       types.Int64  `tfsdk:"privacy_level"`
	Status             types.Int64  `tfsdk:"status"`
	EntityType         types.Int64  `tfsdk:"entity_type"`
	Location           types.String `tfsdk:"location"`
	Image              types.String `tfsdk:"image"`
}

func (r *eventResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_guild_scheduled_event"
}

func (r *eventResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Discord guild scheduled event. Requires `MANAGE_EVENTS`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"guild_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Snowflake ID of the guild. Changing this forces a new event.",
				Validators:          []validator.String{validators.Snowflake()},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Event name (1-100 characters).",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Event description (1-1000 characters).",
			},
			"scheduled_start_time": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ISO 8601 timestamp for when the event starts.",
			},
			"scheduled_end_time": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "ISO 8601 timestamp for when the event ends. Required for EXTERNAL events.",
			},
			"privacy_level": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Privacy level: 2 GUILD_ONLY (the only supported value).",
				Validators:          []validator.Int64{int64validator.OneOf(2)},
			},
			"status": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Event status: 1 SCHEDULED, 2 ACTIVE, 3 COMPLETED, 4 CANCELED.",
				Validators:          []validator.Int64{int64validator.Between(1, 4)},
			},
			"entity_type": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Entity type: 1 STAGE_INSTANCE, 2 VOICE, 3 EXTERNAL.",
				Validators:          []validator.Int64{int64validator.Between(1, 3)},
			},
			"channel_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Stage or voice channel ID. Required for STAGE_INSTANCE and VOICE events; must be omitted for EXTERNAL.",
				Validators:          []validator.String{validators.Snowflake()},
			},
			"location": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Location of an EXTERNAL event (entity_metadata.location). Required for EXTERNAL events.",
			},
			"image": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Cover image as a data URI. Write-only; Discord does not return it on read.",
				Validators:          []validator.String{validators.ImageDataURI()},
			},
		},
	}
}

func (r *eventResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *eventResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan eventModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := createBody{
		ChannelID:          strPtr(plan.ChannelID),
		Name:               plan.Name.ValueString(),
		Description:        strPtr(plan.Description),
		ScheduledStartTime: plan.ScheduledStartTime.ValueString(),
		ScheduledEndTime:   strPtr(plan.ScheduledEndTime),
		EntityType:         plan.EntityType.ValueInt64(),
		EntityMetadata:     locationMeta(plan.Location),
		Image:              strPtr(plan.Image),
		PrivacyLevel:       2,
	}
	if !plan.PrivacyLevel.IsNull() && !plan.PrivacyLevel.IsUnknown() {
		body.PrivacyLevel = plan.PrivacyLevel.ValueInt64()
	}

	ev, err := create(ctx, r.client, plan.GuildID.ValueString(), body, "")
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("creating Discord scheduled event", plan.GuildID.ValueString(), err))
		return
	}
	r.apply(&plan, ev)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *eventResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state eventModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ev, err := get(ctx, r.client, state.GuildID.ValueString(), state.ID.ValueString())
	if err != nil {
		if conns.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(diagutil.APIError("reading Discord scheduled event", state.GuildID.ValueString(), err))
		return
	}
	r.apply(&state, ev)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *eventResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan eventModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	start := plan.ScheduledStartTime.ValueString()
	entityType := plan.EntityType.ValueInt64()
	body := modifyBody{
		ChannelID:          strPtr(plan.ChannelID),
		Name:               &name,
		Description:        strPtr(plan.Description),
		ScheduledStartTime: &start,
		ScheduledEndTime:   strPtr(plan.ScheduledEndTime),
		EntityType:         &entityType,
		EntityMetadata:     locationMeta(plan.Location),
		Image:              strPtr(plan.Image),
	}
	if !plan.PrivacyLevel.IsNull() && !plan.PrivacyLevel.IsUnknown() {
		v := plan.PrivacyLevel.ValueInt64()
		body.PrivacyLevel = &v
	}
	if !plan.Status.IsNull() && !plan.Status.IsUnknown() {
		v := plan.Status.ValueInt64()
		body.Status = &v
	}

	ev, err := modify(ctx, r.client, plan.GuildID.ValueString(), plan.ID.ValueString(), body, "")
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("updating Discord scheduled event", plan.GuildID.ValueString(), err))
		return
	}
	r.apply(&plan, ev)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *eventResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state eventModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := deleteEvent(ctx, r.client, state.GuildID.ValueString(), state.ID.ValueString(), ""); err != nil && !conns.IsNotFound(err) {
		resp.Diagnostics.Append(diagutil.APIError("deleting Discord scheduled event", state.GuildID.ValueString(), err))
	}
}

// ImportState accepts the "guild_id:event_id" format.
func (r *eventResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("expected import ID in the form \"guild_id:event_id\", got %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("guild_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func (r *eventResource) apply(m *eventModel, ev *Event) {
	m.ID = types.StringValue(ev.ID)
	m.GuildID = types.StringValue(ev.GuildID)
	m.Name = types.StringValue(ev.Name)
	m.ScheduledStartTime = types.StringValue(normalizeTime(ev.ScheduledStartTime))
	m.PrivacyLevel = types.Int64Value(ev.PrivacyLevel)
	m.Status = types.Int64Value(ev.Status)
	m.EntityType = types.Int64Value(ev.EntityType)

	m.ChannelID = ptrStr(ev.ChannelID)
	if ev.ScheduledEndTime != nil {
		t := normalizeTime(*ev.ScheduledEndTime)
		m.ScheduledEndTime = types.StringValue(t)
	} else {
		m.ScheduledEndTime = types.StringNull()
	}
	m.Description = optStr(ev.Description)

	if ev.EntityMetadata != nil && ev.EntityMetadata.Location != "" {
		m.Location = types.StringValue(ev.EntityMetadata.Location)
	} else {
		m.Location = types.StringNull()
	}
	// image is write-only; left as configured in state.
}

// locationMeta wraps a location into entity_metadata, or nil when unset.
func locationMeta(loc types.String) *EntityMetadata {
	if loc.IsNull() || loc.IsUnknown() || loc.ValueString() == "" {
		return nil
	}
	return &EntityMetadata{Location: loc.ValueString()}
}

func strPtr(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	return &s
}

func ptrStr(s *string) types.String {
	if s == nil || *s == "" {
		return types.StringNull()
	}
	return types.StringValue(*s)
}

func optStr(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

// normalizeTime converts Discord's +00:00 suffix to the canonical Z form so
// state values stay consistent regardless of which format the API returns.
func normalizeTime(t string) string {
	if strings.HasSuffix(t, "+00:00") {
		return t[:len(t)-6] + "Z"
	}
	return t
}
