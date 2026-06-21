package channel

import (
	"context"
	"fmt"
	"strings"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
	"github.com/alpaca744/terraform-provider-discord/internal/diagutil"
	"github.com/alpaca744/terraform-provider-discord/internal/discord"
	"github.com/alpaca744/terraform-provider-discord/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	_ resource.Resource                = (*overwriteResource)(nil)
	_ resource.ResourceWithConfigure   = (*overwriteResource)(nil)
	_ resource.ResourceWithImportState = (*overwriteResource)(nil)
)

// overwrite type values per Discord.
const (
	overwriteTypeRole   = 0
	overwriteTypeMember = 1
)

// NewOverwriteResource is the resource factory registered with the provider.
func NewOverwriteResource() resource.Resource {
	return &overwriteResource{}
}

type overwriteResource struct {
	client *conns.Client
}

type overwriteResourceModel struct {
	ID          types.String `tfsdk:"id"`
	ChannelID   types.String `tfsdk:"channel_id"`
	OverwriteID types.String `tfsdk:"overwrite_id"`
	Type        types.String `tfsdk:"type"`
	Allow       types.Set    `tfsdk:"allow"`
	Deny        types.Set    `tfsdk:"deny"`
}

func (r *overwriteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_channel_permission_overwrite"
}

func (r *overwriteResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	forceNew := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	permSet := schema.SetAttribute{
		Optional:    true,
		Computed:    true,
		ElementType: types.StringType,
		Validators:  []validator.Set{setvalidator.ValueStringsAre(validators.PermissionName())},
	}
	allow := permSet
	allow.MarkdownDescription = "Set of Discord permission names explicitly allowed by this overwrite."
	deny := permSet
	deny.MarkdownDescription = "Set of Discord permission names explicitly denied by this overwrite."

	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a single permission overwrite on a Discord channel for a role or member. Read through the parent channel, since Discord exposes no standalone overwrite endpoint.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Synthetic ID in the form `channel_id:overwrite_id`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"channel_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Snowflake ID of the channel. Changing this forces a new overwrite.",
				Validators:          []validator.String{validators.Snowflake()},
				PlanModifiers:       forceNew,
			},
			"overwrite_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Snowflake ID of the role or member the overwrite applies to. Changing this forces a new overwrite.",
				Validators:          []validator.String{validators.Snowflake()},
				PlanModifiers:       forceNew,
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Overwrite target type: `role` or `member`. Changing this forces a new overwrite.",
				Validators:          []validator.String{stringvalidator.OneOf("role", "member")},
				PlanModifiers:       forceNew,
			},
			"allow": allow,
			"deny":  deny,
		},
	}
}

func (r *overwriteResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *overwriteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan overwriteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(r.put(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(r.refresh(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *overwriteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state overwriteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ow, found, err := findOverwrite(ctx, r.client, state.ChannelID.ValueString(), state.OverwriteID.ValueString())
	if err != nil {
		if conns.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(diagutil.APIError("reading Discord channel permission overwrite", state.ChannelID.ValueString(), err))
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(r.applyOverwrite(ctx, &state, ow)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *overwriteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan overwriteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(r.put(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(r.refresh(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *overwriteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state overwriteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := deleteOverwrite(ctx, r.client, state.ChannelID.ValueString(), state.OverwriteID.ValueString(), "")
	if err != nil && !conns.IsNotFound(err) {
		resp.Diagnostics.Append(diagutil.APIError("deleting Discord channel permission overwrite", state.ChannelID.ValueString(), err))
	}
}

// ImportState accepts the "channel_id:overwrite_id" format. type is recovered by
// the subsequent Read from the channel object.
func (r *overwriteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("expected import ID in the form \"channel_id:overwrite_id\", got %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("channel_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("overwrite_id"), parts[1])...)
}

// put writes the overwrite from the plan via the Discord PUT endpoint.
func (r *overwriteResource) put(ctx context.Context, plan *overwriteResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	allow, d := setToBitfield(ctx, plan.Allow)
	diags.Append(d...)
	deny, d := setToBitfield(ctx, plan.Deny)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	body := overwriteBody{Type: typeNameToInt(plan.Type.ValueString()), Allow: allow, Deny: deny}
	if err := putOverwrite(ctx, r.client, plan.ChannelID.ValueString(), plan.OverwriteID.ValueString(), body, ""); err != nil {
		diags.Append(diagutil.APIError("setting Discord channel permission overwrite", plan.ChannelID.ValueString(), err))
	}
	return diags
}

// refresh reads the overwrite back so computed allow/deny sets are populated.
func (r *overwriteResource) refresh(ctx context.Context, plan *overwriteResourceModel) diag.Diagnostics {
	ow, found, err := findOverwrite(ctx, r.client, plan.ChannelID.ValueString(), plan.OverwriteID.ValueString())
	if err != nil {
		return diag.Diagnostics{diagutil.APIError("reading Discord channel permission overwrite", plan.ChannelID.ValueString(), err)}
	}
	if !found {
		var diags diag.Diagnostics
		diags.AddError("Overwrite not found after write",
			"Discord did not return the overwrite immediately after it was set.")
		return diags
	}
	return r.applyOverwrite(ctx, plan, ow)
}

func (r *overwriteResource) applyOverwrite(ctx context.Context, m *overwriteResourceModel, ow *Overwrite) diag.Diagnostics {
	var diags diag.Diagnostics
	m.OverwriteID = types.StringValue(ow.ID)
	m.Type = types.StringValue(typeIntToName(ow.Type))
	m.ID = types.StringValue(fmt.Sprintf("%s:%s", m.ChannelID.ValueString(), ow.ID))

	allow, d := bitfieldToSet(ctx, ow.Allow)
	diags.Append(d...)
	deny, d := bitfieldToSet(ctx, ow.Deny)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	m.Allow = allow
	m.Deny = deny
	return diags
}

func typeNameToInt(name string) int64 {
	if name == "member" {
		return overwriteTypeMember
	}
	return overwriteTypeRole
}

func typeIntToName(t int64) string {
	if t == overwriteTypeMember {
		return "member"
	}
	return "role"
}

func setToBitfield(ctx context.Context, set types.Set) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if set.IsNull() || set.IsUnknown() {
		return "0", diags
	}
	var names []string
	diags.Append(set.ElementsAs(ctx, &names, false)...)
	if diags.HasError() {
		return "0", diags
	}
	bf, err := discord.PermissionsToBitfield(names)
	if err != nil {
		diags.AddError("Invalid permissions", err.Error())
		return "0", diags
	}
	return bf, diags
}

func bitfieldToSet(ctx context.Context, bitfield string) (types.Set, diag.Diagnostics) {
	names, err := discord.BitfieldToPermissions(bitfield)
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("Invalid permissions from API", err.Error())
		return types.SetNull(types.StringType), diags
	}
	set, d := types.SetValueFrom(ctx, types.StringType, names)
	return set, d
}
