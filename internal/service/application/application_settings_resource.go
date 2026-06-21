package application

import (
	"context"
	"fmt"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
	"github.com/alpaca744/terraform-provider-discord/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = (*settingsResource)(nil)
	_ resource.ResourceWithConfigure   = (*settingsResource)(nil)
	_ resource.ResourceWithImportState = (*settingsResource)(nil)
)

// NewSettingsResource is the resource factory registered with the provider.
func NewSettingsResource() resource.Resource {
	return &settingsResource{}
}

// settingsResource manages the current application's settings via
// PATCH /applications/@me. The application itself is not created or destroyed by
// this resource, so Delete is a no-op.
type settingsResource struct {
	client *conns.Client
}

type settingsModel struct {
	ID                             types.String `tfsdk:"id"`
	Description                    types.String `tfsdk:"description"`
	InteractionsEndpointURL        types.String `tfsdk:"interactions_endpoint_url"`
	RoleConnectionsVerificationURL types.String `tfsdk:"role_connections_verification_url"`
	CustomInstallURL               types.String `tfsdk:"custom_install_url"`
	Tags                           types.Set    `tfsdk:"tags"`
}

func (r *settingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_settings"
}

func (r *settingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages settings on the current Discord application (the app behind the configured bot token). The application is not created or destroyed by this resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The application snowflake ID.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Application description.",
			},
			"interactions_endpoint_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "HTTP interactions endpoint URL.",
			},
			"role_connections_verification_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Role connection verification URL for linked roles.",
			},
			"custom_install_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Custom authorization (install) URL.",
			},
			"tags": schema.SetAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Up to 5 descriptive tags for the application.",
			},
		},
	}
}

func (r *settingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *settingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan settingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := r.body(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := ModifyCurrentApplication(ctx, r.client, body)
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("applying Discord application settings", "", err))
		return
	}
	r.apply(&plan, app)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *settingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state settingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := GetCurrentApplication(ctx, r.client)
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("reading Discord application settings", "", err))
		return
	}
	r.apply(&state, app)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *settingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan settingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := r.body(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := ModifyCurrentApplication(ctx, r.client, body)
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("updating Discord application settings", "", err))
		return
	}
	r.apply(&plan, app)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete is a no-op: the application is not owned by this resource.
func (r *settingsResource) Delete(ctx context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	tflog.Info(ctx, "discord_application_settings deletion is a no-op; the application configuration is left unchanged")
}

// ImportState reads the current application; the import ID is ignored because
// the endpoint is always the token's own application.
func (r *settingsResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "")...)
}

func (r *settingsResource) body(ctx context.Context, plan settingsModel) (ApplicationSettingsBody, diag.Diagnostics) {
	body := ApplicationSettingsBody{
		Description:                    strPtr(plan.Description),
		InteractionsEndpointURL:        strPtr(plan.InteractionsEndpointURL),
		RoleConnectionsVerificationURL: strPtr(plan.RoleConnectionsVerificationURL),
		CustomInstallURL:               strPtr(plan.CustomInstallURL),
	}
	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		var tags []string
		if d := plan.Tags.ElementsAs(ctx, &tags, false); d.HasError() {
			return body, d
		}
		body.Tags = tags
	}
	return body, nil
}

func (r *settingsResource) apply(m *settingsModel, app *Application) {
	m.ID = types.StringValue(app.ID)
	m.Description = types.StringValue(app.Description)
	m.InteractionsEndpointURL = optStr(app.InteractionsEndpointURL)
	m.RoleConnectionsVerificationURL = optStr(app.RoleConnectionsVerificationURL)
	m.CustomInstallURL = optStr(app.CustomInstallURL)
	// Tags is left as configured; the API echoes it back and round-trips via the
	// plan, so it is not re-derived here to avoid ordering churn on a set.
}
