package application

import (
	"context"
	"fmt"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
	"github.com/alpaca744/terraform-provider-discord/internal/diagutil"
	"github.com/alpaca744/terraform-provider-discord/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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
	_ resource.Resource                = (*roleConnectionMetadataResource)(nil)
	_ resource.ResourceWithConfigure   = (*roleConnectionMetadataResource)(nil)
	_ resource.ResourceWithImportState = (*roleConnectionMetadataResource)(nil)
)

// NewRoleConnectionMetadataResource is the resource factory.
func NewRoleConnectionMetadataResource() resource.Resource {
	return &roleConnectionMetadataResource{}
}

// roleConnectionMetadataResource manages the full set of role connection
// metadata records for an application. Discord exposes only a replace-all PUT
// (no per-record CRUD), so the entire `records` list is modeled as one resource.
type roleConnectionMetadataResource struct {
	client *conns.Client
}

type roleConnectionMetadataModel struct {
	ApplicationID types.String  `tfsdk:"application_id"`
	Records       []recordModel `tfsdk:"records"`
}

type recordModel struct {
	Type        types.Int64  `tfsdk:"type"`
	Key         types.String `tfsdk:"key"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func (r *roleConnectionMetadataResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_role_connection_metadata"
}

func (r *roleConnectionMetadataResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the full set of role connection metadata records for an application. Discord replaces all records on every write, so this resource owns the entire list.",
		Attributes: map[string]schema.Attribute{
			"application_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Snowflake ID of the application. Changing this forces a new resource.",
				Validators:          []validator.String{validators.Snowflake()},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"records": schema.ListNestedAttribute{
				Required:            true,
				MarkdownDescription: "Up to 5 metadata records.",
				Validators:          []validator.List{listvalidator.SizeAtMost(5)},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.Int64Attribute{
							Required:            true,
							MarkdownDescription: "Metadata type (1-8): comparison operator applied to the value.",
							Validators:          []validator.Int64{int64validator.Between(1, 8)},
						},
						"key": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Dictionary key for the metadata field (a-z, 0-9, _; max 50).",
						},
						"name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Display name of the metadata field (max 100).",
						},
						"description": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Description of the metadata field (max 200).",
						},
					},
				},
			},
		},
	}
}

func (r *roleConnectionMetadataResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *roleConnectionMetadataResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(r.write(ctx, req.Plan, &resp.State)...)
}

func (r *roleConnectionMetadataResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(r.write(ctx, req.Plan, &resp.State)...)
}

func (r *roleConnectionMetadataResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state roleConnectionMetadataModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	records, err := GetRoleConnectionMetadata(ctx, r.client, state.ApplicationID.ValueString())
	if err != nil {
		if conns.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(diagutil.APIError("reading Discord application role connection metadata", state.ApplicationID.ValueString(), err))
		return
	}
	state.Records = flattenRecords(records)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete clears all metadata records by writing an empty list.
func (r *roleConnectionMetadataResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state roleConnectionMetadataModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := PutRoleConnectionMetadata(ctx, r.client, state.ApplicationID.ValueString(), []RoleConnectionMetadata{})
	if err != nil && !conns.IsNotFound(err) {
		resp.Diagnostics.Append(diagutil.APIError("clearing Discord application role connection metadata", state.ApplicationID.ValueString(), err))
	}
}

// ImportState accepts a bare application ID.
func (r *roleConnectionMetadataResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("application_id"), req.ID)...)
}

// write applies the planned records via the replace-all PUT and stores the
// API's echo back into state. Shared by Create and Update, which are identical
// for a replace-all resource.
func (r *roleConnectionMetadataResource) write(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State) diag.Diagnostics {
	var diags diag.Diagnostics
	var model roleConnectionMetadataModel
	diags.Append(plan.Get(ctx, &model)...)
	if diags.HasError() {
		return diags
	}

	records, err := PutRoleConnectionMetadata(ctx, r.client, model.ApplicationID.ValueString(), expandRecords(model.Records))
	if err != nil {
		diags.Append(diagutil.APIError("updating Discord application role connection metadata", model.ApplicationID.ValueString(), err))
		return diags
	}
	model.Records = flattenRecords(records)
	diags.Append(state.Set(ctx, &model)...)
	return diags
}

func expandRecords(in []recordModel) []RoleConnectionMetadata {
	out := make([]RoleConnectionMetadata, 0, len(in))
	for _, m := range in {
		out = append(out, RoleConnectionMetadata{
			Type:        m.Type.ValueInt64(),
			Key:         m.Key.ValueString(),
			Name:        m.Name.ValueString(),
			Description: m.Description.ValueString(),
		})
	}
	return out
}

func flattenRecords(in []RoleConnectionMetadata) []recordModel {
	out := make([]recordModel, 0, len(in))
	for _, rec := range in {
		out = append(out, recordModel{
			Type:        types.Int64Value(rec.Type),
			Key:         types.StringValue(rec.Key),
			Name:        types.StringValue(rec.Name),
			Description: types.StringValue(rec.Description),
		})
	}
	return out
}
