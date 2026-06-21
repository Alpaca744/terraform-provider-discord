package guild

import (
	"context"
	"fmt"
	"strings"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
	"github.com/alpaca744/terraform-provider-discord/internal/diagutil"
	"github.com/alpaca744/terraform-provider-discord/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*templateResource)(nil)
	_ resource.ResourceWithConfigure   = (*templateResource)(nil)
	_ resource.ResourceWithImportState = (*templateResource)(nil)
)

// NewTemplateResource is the resource factory.
func NewTemplateResource() resource.Resource {
	return &templateResource{}
}

type templateResource struct {
	client *conns.Client
}

type templateModel struct {
	Code          types.String `tfsdk:"code"`
	GuildID       types.String `tfsdk:"guild_id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	UsageCount    types.Int64  `tfsdk:"usage_count"`
	SourceGuildID types.String `tfsdk:"source_guild_id"`
}

func (r *templateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_guild_template"
}

func (r *templateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Discord guild template, a snapshot of a guild's settings that can be used to create new guilds.",
		Attributes: map[string]schema.Attribute{
			"code": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique template code (also the resource ID).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"guild_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Snowflake ID of the source guild. Changing this forces a new template.",
				Validators:          []validator.String{validators.Snowflake()},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Template name (1-100 characters).",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Template description (0-120 characters).",
			},
			"usage_count": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of times the template has been used.",
			},
			"source_guild_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Snowflake ID of the guild the template is based on.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *templateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *templateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan templateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	t, err := CreateTemplate(ctx, r.client, plan.GuildID.ValueString(), r.body(plan), "")
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("creating Discord guild template", plan.GuildID.ValueString(), err))
		return
	}
	r.apply(&plan, t)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *templateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state templateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	t, err := GetTemplate(ctx, r.client, state.Code.ValueString())
	if err != nil {
		if conns.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(diagutil.APIError("reading Discord guild template", state.GuildID.ValueString(), err))
		return
	}
	r.apply(&state, t)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *templateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan templateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	t, err := ModifyTemplate(ctx, r.client, plan.GuildID.ValueString(), plan.Code.ValueString(), r.body(plan), "")
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("updating Discord guild template", plan.GuildID.ValueString(), err))
		return
	}
	r.apply(&plan, t)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *templateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state templateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := DeleteTemplate(ctx, r.client, state.GuildID.ValueString(), state.Code.ValueString(), ""); err != nil && !conns.IsNotFound(err) {
		resp.Diagnostics.Append(diagutil.APIError("deleting Discord guild template", state.GuildID.ValueString(), err))
	}
}

// ImportState accepts the "guild_id:code" format.
func (r *templateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("expected import ID in the form \"guild_id:code\", got %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("guild_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("code"), parts[1])...)
}

func (r *templateResource) body(m templateModel) templateWriteBody {
	return templateWriteBody{Name: strPtr(m.Name), Description: strPtr(m.Description)}
}

func (r *templateResource) apply(m *templateModel, t *Template) {
	m.Code = types.StringValue(t.Code)
	m.Name = types.StringValue(t.Name)
	m.UsageCount = types.Int64Value(t.UsageCount)
	if t.SourceGuildID != "" {
		m.SourceGuildID = types.StringValue(t.SourceGuildID)
		m.GuildID = types.StringValue(t.SourceGuildID)
	}
	m.Description = optStr(t.Description)
}
