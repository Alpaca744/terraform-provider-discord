package guild

import (
	"context"
	"fmt"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
	"github.com/alpaca744/terraform-provider-discord/internal/diagutil"
	"github.com/alpaca744/terraform-provider-discord/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*auditLogDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*auditLogDataSource)(nil)
)

// NewAuditLogDataSource is the data source factory.
func NewAuditLogDataSource() datasource.DataSource {
	return &auditLogDataSource{}
}

type auditLogDataSource struct {
	client *conns.Client
}

type auditLogModel struct {
	GuildID    types.String         `tfsdk:"guild_id"`
	UserID     types.String         `tfsdk:"user_id"`
	ActionType types.Int64          `tfsdk:"action_type"`
	Limit      types.Int64          `tfsdk:"limit"`
	Entries    []auditLogEntryModel `tfsdk:"entries"`
}

type auditLogEntryModel struct {
	ID         types.String `tfsdk:"id"`
	UserID     types.String `tfsdk:"user_id"`
	TargetID   types.String `tfsdk:"target_id"`
	ActionType types.Int64  `tfsdk:"action_type"`
	Reason     types.String `tfsdk:"reason"`
}

func (d *auditLogDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_audit_log"
}

func (d *auditLogDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads entries from a Discord guild's audit log. Requires the `VIEW_AUDIT_LOG` permission.",
		Attributes: map[string]schema.Attribute{
			"guild_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Snowflake ID of the guild.",
				Validators:          []validator.String{validators.Snowflake()},
			},
			"user_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter entries to those performed by this user ID.",
				Validators:          []validator.String{validators.Snowflake()},
			},
			"action_type": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Filter entries to this audit log action type.",
			},
			"limit": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Maximum number of entries to return (1-100).",
				Validators:          []validator.Int64{int64validator.Between(1, 100)},
			},
			"entries": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The matching audit log entries.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.StringAttribute{Computed: true, MarkdownDescription: "Entry snowflake ID."},
						"user_id":     schema.StringAttribute{Computed: true, MarkdownDescription: "ID of the user who performed the action."},
						"target_id":   schema.StringAttribute{Computed: true, MarkdownDescription: "ID of the affected entity."},
						"action_type": schema.Int64Attribute{Computed: true, MarkdownDescription: "The audit log action type."},
						"reason":      schema.StringAttribute{Computed: true, MarkdownDescription: "The reason supplied for the action, if any."},
					},
				},
			},
		},
	}
}

func (d *auditLogDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*conns.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data",
			fmt.Sprintf("expected *conns.Client, got %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *auditLogDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data auditLogModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	q := AuditLogQuery{UserID: data.UserID.ValueString()}
	if !data.ActionType.IsNull() {
		v := data.ActionType.ValueInt64()
		q.ActionType = &v
	}
	if !data.Limit.IsNull() {
		v := data.Limit.ValueInt64()
		q.Limit = &v
	}

	log, err := GetAuditLog(ctx, d.client, data.GuildID.ValueString(), q)
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("reading Discord audit log", data.GuildID.ValueString(), err))
		return
	}

	data.Entries = make([]auditLogEntryModel, 0, len(log.Entries))
	for _, e := range log.Entries {
		data.Entries = append(data.Entries, auditLogEntryModel{
			ID:         types.StringValue(e.ID),
			UserID:     ptrStr(e.UserID),
			TargetID:   ptrStr(e.TargetID),
			ActionType: types.Int64Value(e.ActionType),
			Reason:     optStr(e.Reason),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
