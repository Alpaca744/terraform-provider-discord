package automod

import (
	"context"
	"fmt"
	"strings"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
	"github.com/alpaca744/terraform-provider-discord/internal/diagutil"
	"github.com/alpaca744/terraform-provider-discord/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*ruleResource)(nil)
	_ resource.ResourceWithConfigure   = (*ruleResource)(nil)
	_ resource.ResourceWithImportState = (*ruleResource)(nil)
)

// NewRuleResource is the resource factory registered with the provider.
func NewRuleResource() resource.Resource {
	return &ruleResource{}
}

type ruleResource struct {
	client *conns.Client
}

type ruleModel struct {
	ID              types.String          `tfsdk:"id"`
	GuildID         types.String          `tfsdk:"guild_id"`
	Name            types.String          `tfsdk:"name"`
	EventType       types.Int64           `tfsdk:"event_type"`
	TriggerType     types.Int64           `tfsdk:"trigger_type"`
	Enabled         types.Bool            `tfsdk:"enabled"`
	ExemptRoles     types.Set             `tfsdk:"exempt_roles"`
	ExemptChannels  types.Set             `tfsdk:"exempt_channels"`
	TriggerMetadata *triggerMetadataModel `tfsdk:"trigger_metadata"`
	Actions         []actionModel         `tfsdk:"actions"`
}

type triggerMetadataModel struct {
	KeywordFilter                types.Set   `tfsdk:"keyword_filter"`
	RegexPatterns                types.Set   `tfsdk:"regex_patterns"`
	Presets                      types.Set   `tfsdk:"presets"`
	AllowList                    types.Set   `tfsdk:"allow_list"`
	MentionTotalLimit            types.Int64 `tfsdk:"mention_total_limit"`
	MentionRaidProtectionEnabled types.Bool  `tfsdk:"mention_raid_protection_enabled"`
}

type actionModel struct {
	Type            types.Int64  `tfsdk:"type"`
	ChannelID       types.String `tfsdk:"channel_id"`
	DurationSeconds types.Int64  `tfsdk:"duration_seconds"`
	CustomMessage   types.String `tfsdk:"custom_message"`
}

func (r *ruleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_auto_moderation_rule"
}

func (r *ruleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	stringSet := func(desc string) schema.SetAttribute {
		return schema.SetAttribute{Optional: true, ElementType: types.StringType, MarkdownDescription: desc}
	}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Discord auto moderation rule. Requires `MANAGE_GUILD`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"guild_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Snowflake ID of the guild. Changing this forces a new rule.",
				Validators:          []validator.String{validators.Snowflake()},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The rule name.",
			},
			"event_type": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Event context the rule checks: 1 (MESSAGE_SEND) or 2 (MEMBER_UPDATE).",
			},
			"trigger_type": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Trigger type: 1 KEYWORD, 3 SPAM, 4 KEYWORD_PRESET, 5 MENTION_SPAM, 6 MEMBER_PROFILE. Immutable; changing forces a new rule.",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the rule is enabled.",
			},
			"exempt_roles":    stringSet("Snowflake IDs of roles exempt from the rule (max 20)."),
			"exempt_channels": stringSet("Snowflake IDs of channels exempt from the rule (max 50)."),
			"trigger_metadata": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Trigger-type-specific configuration.",
				Attributes: map[string]schema.Attribute{
					"keyword_filter": stringSet("Substrings to match against content (KEYWORD)."),
					"regex_patterns": stringSet("Regular expressions to match against content (KEYWORD)."),
					"allow_list":     stringSet("Substrings exempt from triggering the rule."),
					"presets": schema.SetAttribute{
						Optional:            true,
						ElementType:         types.Int64Type,
						MarkdownDescription: "Preset word lists (KEYWORD_PRESET): 1 PROFANITY, 2 SEXUAL_CONTENT, 3 SLURS.",
					},
					"mention_total_limit": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Total mentions allowed per message (MENTION_SPAM).",
					},
					"mention_raid_protection_enabled": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Whether mention raid protection is enabled (MENTION_SPAM).",
					},
				},
			},
			"actions": schema.ListNestedAttribute{
				Required:            true,
				MarkdownDescription: "Actions taken when the rule is triggered.",
				Validators:          []validator.List{listvalidator.SizeAtLeast(1)},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.Int64Attribute{
							Required:            true,
							MarkdownDescription: "Action type: 1 BLOCK_MESSAGE, 2 SEND_ALERT_MESSAGE, 3 TIMEOUT.",
						},
						"channel_id": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Channel to send alerts to (SEND_ALERT_MESSAGE).",
							Validators:          []validator.String{validators.Snowflake()},
						},
						"duration_seconds": schema.Int64Attribute{
							Optional:            true,
							MarkdownDescription: "Timeout duration in seconds, max 2419200 (TIMEOUT).",
						},
						"custom_message": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Custom message shown when a message is blocked (BLOCK_MESSAGE).",
						},
					},
				},
			},
		},
	}
}

func (r *ruleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ruleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ruleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tm, actions, exemptRoles, exemptChannels, diags := r.expand(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := createBody{
		Name:            plan.Name.ValueString(),
		EventType:       plan.EventType.ValueInt64(),
		TriggerType:     plan.TriggerType.ValueInt64(),
		TriggerMetadata: tm,
		Actions:         actions,
		Enabled:         plan.Enabled.ValueBool(),
		ExemptRoles:     exemptRoles,
		ExemptChannels:  exemptChannels,
	}
	rule, err := create(ctx, r.client, plan.GuildID.ValueString(), body, "")
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("creating Discord auto moderation rule", plan.GuildID.ValueString(), err))
		return
	}

	resp.Diagnostics.Append(r.flatten(ctx, &plan, rule)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ruleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ruleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := get(ctx, r.client, state.GuildID.ValueString(), state.ID.ValueString())
	if err != nil {
		if conns.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(diagutil.APIError("reading Discord auto moderation rule", state.GuildID.ValueString(), err))
		return
	}

	resp.Diagnostics.Append(r.flatten(ctx, &state, rule)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ruleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ruleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tm, actions, exemptRoles, exemptChannels, diags := r.expand(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := modifyBody{
		Name:            plan.Name.ValueString(),
		EventType:       plan.EventType.ValueInt64(),
		TriggerMetadata: tm,
		Actions:         actions,
		Enabled:         plan.Enabled.ValueBool(),
		ExemptRoles:     exemptRoles,
		ExemptChannels:  exemptChannels,
	}
	rule, err := modify(ctx, r.client, plan.GuildID.ValueString(), plan.ID.ValueString(), body, "")
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("updating Discord auto moderation rule", plan.GuildID.ValueString(), err))
		return
	}

	resp.Diagnostics.Append(r.flatten(ctx, &plan, rule)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ruleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ruleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := deleteRule(ctx, r.client, state.GuildID.ValueString(), state.ID.ValueString(), ""); err != nil && !conns.IsNotFound(err) {
		resp.Diagnostics.Append(diagutil.APIError("deleting Discord auto moderation rule", state.GuildID.ValueString(), err))
	}
}

// ImportState accepts the "guild_id:rule_id" format.
func (r *ruleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("expected import ID in the form \"guild_id:rule_id\", got %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("guild_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
