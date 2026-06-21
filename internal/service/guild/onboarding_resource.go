package guild

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
	"github.com/alpaca744/terraform-provider-discord/internal/diagutil"
	"github.com/alpaca744/terraform-provider-discord/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*onboardingResource)(nil)
	_ resource.ResourceWithConfigure   = (*onboardingResource)(nil)
	_ resource.ResourceWithImportState = (*onboardingResource)(nil)
)

// NewOnboardingResource is the resource factory.
func NewOnboardingResource() resource.Resource {
	return &onboardingResource{}
}

// onboardingResource manages a guild's onboarding configuration. Discord exposes
// only a replace-all PUT, so this resource owns the whole configuration.
type onboardingResource struct {
	client *conns.Client
}

type onboardingModel struct {
	GuildID           types.String         `tfsdk:"guild_id"`
	Enabled           types.Bool           `tfsdk:"enabled"`
	Mode              types.Int64          `tfsdk:"mode"`
	DefaultChannelIDs types.Set            `tfsdk:"default_channel_ids"`
	Prompts           jsontypes.Normalized `tfsdk:"prompts"`
}

func (r *onboardingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_guild_onboarding"
}

func (r *onboardingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a guild's onboarding configuration. Discord replaces the whole configuration on every write, so this resource owns it entirely.",
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
				MarkdownDescription: "Whether onboarding is enabled.",
			},
			"mode": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Onboarding mode: 0 ONBOARDING_DEFAULT, 1 ONBOARDING_ADVANCED.",
				Validators:          []validator.Int64{int64validator.OneOf(0, 1)},
			},
			"default_channel_ids": schema.SetAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Channel IDs members are opted into by default.",
				Validators:          []validator.Set{setvalidator.ValueStringsAre(validators.Snowflake())},
			},
			"prompts": schema.StringAttribute{
				Optional:            true,
				CustomType:          jsontypes.NormalizedType{},
				MarkdownDescription: "Onboarding prompts as a JSON array. Carried as JSON to support the deeply nested prompt/option structure. Compared semantically: Discord assigns its own prompt/option `id` values and adds default fields, so those are ignored when detecting drift and your configured prompts remain authoritative. Use `null` (or omit) for no prompts.",
			},
		},
	}
}

func (r *onboardingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *onboardingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.write(ctx, req.Plan, &resp.State, &resp.Diagnostics)
}

func (r *onboardingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.write(ctx, req.Plan, &resp.State, &resp.Diagnostics)
}

func (r *onboardingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state onboardingModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ob, err := GetOnboarding(ctx, r.client, state.GuildID.ValueString())
	if err != nil {
		if conns.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(diagutil.APIError("reading Discord guild onboarding", state.GuildID.ValueString(), err))
		return
	}
	resp.Diagnostics.Append(r.apply(ctx, &state, ob)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete is a no-op: onboarding cannot be deleted, only reconfigured.
func (r *onboardingResource) Delete(ctx context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

// ImportState accepts a bare guild ID.
func (r *onboardingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("guild_id"), req.ID)...)
}

func (r *onboardingResource) write(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State, diags *diag.Diagnostics) {
	var model onboardingModel
	diags.Append(plan.Get(ctx, &model)...)
	if diags.HasError() {
		return
	}

	var channels []string
	if !model.DefaultChannelIDs.IsNull() && !model.DefaultChannelIDs.IsUnknown() {
		diags.Append(model.DefaultChannelIDs.ElementsAs(ctx, &channels, false)...)
		if diags.HasError() {
			return
		}
	}

	prompts := json.RawMessage("[]")
	if !model.Prompts.IsNull() && !model.Prompts.IsUnknown() && model.Prompts.ValueString() != "" {
		prompts = json.RawMessage(model.Prompts.ValueString())
	}

	body := OnboardingBody{
		Prompts:           prompts,
		DefaultChannelIDs: channels,
		Enabled:           model.Enabled.ValueBool(),
		Mode:              model.Mode.ValueInt64(),
	}
	ob, err := PutOnboarding(ctx, r.client, model.GuildID.ValueString(), body, "")
	if err != nil {
		diags.Append(diagutil.APIError("updating Discord guild onboarding", model.GuildID.ValueString(), err))
		return
	}
	diags.Append(r.apply(ctx, &model, ob)...)
	diags.Append(state.Set(ctx, &model)...)
}

func (r *onboardingResource) apply(ctx context.Context, m *onboardingModel, ob *Onboarding) diag.Diagnostics {
	var diags diag.Diagnostics
	m.Enabled = types.BoolValue(ob.Enabled)
	m.Mode = types.Int64Value(ob.Mode)

	if len(ob.DefaultChannelIDs) == 0 {
		m.DefaultChannelIDs = types.SetNull(types.StringType)
	} else {
		set, d := types.SetValueFrom(ctx, types.StringType, ob.DefaultChannelIDs)
		diags.Append(d...)
		m.DefaultChannelIDs = set
	}

	// Discord reassigns every prompt/option id, reorders object keys, and enriches
	// options with default fields (empty role_ids, emoji defaults, and so on), so
	// the response never byte- or jsontypes-equals the configured prompts. Keep
	// the configured value (plan on write, prior state on read) while the response
	// still agrees with everything the user specified; only adopt the API value on
	// a genuine difference. This keeps an empty "[]" from collapsing to null and
	// stops server-assigned ids from causing perpetual drift.
	if !promptsAgree(m.Prompts, ob.Prompts) {
		m.Prompts = flattenPrompts(ob.Prompts)
	}
	return diags
}

// flattenPrompts maps the API prompts into the normalized attribute, using null
// for an empty list so an unset configuration does not show drift.
func flattenPrompts(raw json.RawMessage) jsontypes.Normalized {
	if promptsEmpty(raw) {
		return jsontypes.NewNormalizedNull()
	}
	return jsontypes.NewNormalizedValue(string(raw))
}

// promptsEmpty reports whether a prompts JSON value carries no prompts.
func promptsEmpty(raw json.RawMessage) bool {
	s := strings.TrimSpace(string(raw))
	return s == "" || s == "null" || s == "[]"
}

// promptsAgree reports whether the prompts Discord returned still match what the
// user configured. Empty on both sides agrees; empty on only one side does not.
// Otherwise the configured value is authoritative as long as the response is a
// superset of it, ignoring server-assigned ids (see jsonSubset).
func promptsAgree(configured jsontypes.Normalized, apiPrompts json.RawMessage) bool {
	cfgEmpty := configured.IsNull() || configured.IsUnknown() ||
		promptsEmpty(json.RawMessage(configured.ValueString()))
	apiEmpty := promptsEmpty(apiPrompts)
	if cfgEmpty || apiEmpty {
		return cfgEmpty == apiEmpty
	}

	var want, got any
	if err := json.Unmarshal([]byte(configured.ValueString()), &want); err != nil {
		return false
	}
	if err := json.Unmarshal(apiPrompts, &got); err != nil {
		return false
	}
	return jsonSubset(want, got)
}

// jsonSubset reports whether got contains everything want specifies. Object keys
// named "id" in want are ignored because Discord assigns its own prompt/option
// ids; extra keys in got are ignored. Arrays must match length and agree
// element-wise, and scalars must be equal.
func jsonSubset(want, got any) bool {
	switch w := want.(type) {
	case map[string]any:
		g, ok := got.(map[string]any)
		if !ok {
			return false
		}
		for k, wv := range w {
			if k == "id" {
				continue
			}
			gv, ok := g[k]
			if !ok || !jsonSubset(wv, gv) {
				return false
			}
		}
		return true
	case []any:
		g, ok := got.([]any)
		if !ok || len(g) != len(w) {
			return false
		}
		for i := range w {
			if !jsonSubset(w[i], g[i]) {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(want, got)
	}
}
