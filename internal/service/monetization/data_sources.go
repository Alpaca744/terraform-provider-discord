package monetization

import (
	"context"
	"fmt"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
	"github.com/alpaca744/terraform-provider-discord/internal/diagutil"
	"github.com/alpaca744/terraform-provider-discord/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func configure(providerData any, addError func(string, string)) *conns.Client {
	if providerData == nil {
		return nil
	}
	client, ok := providerData.(*conns.Client)
	if !ok {
		addError("Unexpected provider data", fmt.Sprintf("expected *conns.Client, got %T", providerData))
		return nil
	}
	return client
}

func optStr(s *string) types.String {
	if s == nil || *s == "" {
		return types.StringNull()
	}
	return types.StringValue(*s)
}

// ---- discord_sku ----

var (
	_ datasource.DataSource              = (*skuDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*skuDataSource)(nil)
)

// NewSKUDataSource is the data source factory.
func NewSKUDataSource() datasource.DataSource { return &skuDataSource{} }

type skuDataSource struct{ client *conns.Client }

type skuModel struct {
	ApplicationID types.String    `tfsdk:"application_id"`
	SKUs          []skuEntryModel `tfsdk:"skus"`
}

type skuEntryModel struct {
	ID    types.String `tfsdk:"id"`
	Type  types.Int64  `tfsdk:"type"`
	Name  types.String `tfsdk:"name"`
	Slug  types.String `tfsdk:"slug"`
	Flags types.Int64  `tfsdk:"flags"`
}

func (d *skuDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sku"
}

func (d *skuDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists the SKUs (premium offerings) for a Discord application.",
		Attributes: map[string]schema.Attribute{
			"application_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Snowflake ID of the application.",
				Validators:          []validator.String{validators.Snowflake()},
			},
			"skus": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The application's SKUs.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":    schema.StringAttribute{Computed: true, MarkdownDescription: "SKU snowflake ID."},
						"type":  schema.Int64Attribute{Computed: true, MarkdownDescription: "SKU type."},
						"name":  schema.StringAttribute{Computed: true, MarkdownDescription: "Customer-facing SKU name."},
						"slug":  schema.StringAttribute{Computed: true, MarkdownDescription: "URL slug for the SKU."},
						"flags": schema.Int64Attribute{Computed: true, MarkdownDescription: "SKU flags bitfield."},
					},
				},
			},
		},
	}
}

func (d *skuDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configure(req.ProviderData, resp.Diagnostics.AddError)
}

func (d *skuDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data skuModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	skus, err := ListSKUs(ctx, d.client, data.ApplicationID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("listing Discord SKUs", data.ApplicationID.ValueString(), err))
		return
	}
	data.SKUs = make([]skuEntryModel, 0, len(skus))
	for _, s := range skus {
		data.SKUs = append(data.SKUs, skuEntryModel{
			ID:    types.StringValue(s.ID),
			Type:  types.Int64Value(s.Type),
			Name:  types.StringValue(s.Name),
			Slug:  optStr(&s.Slug),
			Flags: types.Int64Value(s.Flags),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---- discord_entitlement ----

var (
	_ datasource.DataSource              = (*entitlementDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*entitlementDataSource)(nil)
)

// NewEntitlementDataSource is the data source factory.
func NewEntitlementDataSource() datasource.DataSource { return &entitlementDataSource{} }

type entitlementDataSource struct{ client *conns.Client }

type entitlementModel struct {
	ApplicationID types.String            `tfsdk:"application_id"`
	UserID        types.String            `tfsdk:"user_id"`
	SKUID         types.String            `tfsdk:"sku_id"`
	Entitlements  []entitlementEntryModel `tfsdk:"entitlements"`
}

type entitlementEntryModel struct {
	ID       types.String `tfsdk:"id"`
	SKUID    types.String `tfsdk:"sku_id"`
	UserID   types.String `tfsdk:"user_id"`
	GuildID  types.String `tfsdk:"guild_id"`
	Type     types.Int64  `tfsdk:"type"`
	Deleted  types.Bool   `tfsdk:"deleted"`
	StartsAt types.String `tfsdk:"starts_at"`
	EndsAt   types.String `tfsdk:"ends_at"`
}

func (d *entitlementDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_entitlement"
}

func (d *entitlementDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists entitlements (granted SKUs) for a Discord application, optionally filtered by user or SKU.",
		Attributes: map[string]schema.Attribute{
			"application_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Snowflake ID of the application.",
				Validators:          []validator.String{validators.Snowflake()},
			},
			"user_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter to entitlements for this user ID.",
				Validators:          []validator.String{validators.Snowflake()},
			},
			"sku_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter to entitlements for this SKU ID.",
				Validators:          []validator.String{validators.Snowflake()},
			},
			"entitlements": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The matching entitlements.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":        schema.StringAttribute{Computed: true, MarkdownDescription: "Entitlement snowflake ID."},
						"sku_id":    schema.StringAttribute{Computed: true, MarkdownDescription: "SKU the entitlement grants."},
						"user_id":   schema.StringAttribute{Computed: true, MarkdownDescription: "User the entitlement is for."},
						"guild_id":  schema.StringAttribute{Computed: true, MarkdownDescription: "Guild the entitlement is for."},
						"type":      schema.Int64Attribute{Computed: true, MarkdownDescription: "Entitlement type."},
						"deleted":   schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the entitlement was deleted."},
						"starts_at": schema.StringAttribute{Computed: true, MarkdownDescription: "Start of the entitlement validity."},
						"ends_at":   schema.StringAttribute{Computed: true, MarkdownDescription: "End of the entitlement validity."},
					},
				},
			},
		},
	}
}

func (d *entitlementDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configure(req.ProviderData, resp.Diagnostics.AddError)
}

func (d *entitlementDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data entitlementModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ents, err := ListEntitlements(ctx, d.client, data.ApplicationID.ValueString(), data.UserID.ValueString(), data.SKUID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("listing Discord entitlements", data.ApplicationID.ValueString(), err))
		return
	}
	data.Entitlements = make([]entitlementEntryModel, 0, len(ents))
	for _, e := range ents {
		data.Entitlements = append(data.Entitlements, entitlementEntryModel{
			ID:       types.StringValue(e.ID),
			SKUID:    types.StringValue(e.SKUID),
			UserID:   optStr(e.UserID),
			GuildID:  optStr(e.GuildID),
			Type:     types.Int64Value(e.Type),
			Deleted:  types.BoolValue(e.Deleted),
			StartsAt: optStr(e.StartsAt),
			EndsAt:   optStr(e.EndsAt),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---- discord_subscription ----

var (
	_ datasource.DataSource              = (*subscriptionDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*subscriptionDataSource)(nil)
)

// NewSubscriptionDataSource is the data source factory.
func NewSubscriptionDataSource() datasource.DataSource { return &subscriptionDataSource{} }

type subscriptionDataSource struct{ client *conns.Client }

type subscriptionModel struct {
	SKUID         types.String             `tfsdk:"sku_id"`
	UserID        types.String             `tfsdk:"user_id"`
	Subscriptions []subscriptionEntryModel `tfsdk:"subscriptions"`
}

type subscriptionEntryModel struct {
	ID     types.String `tfsdk:"id"`
	UserID types.String `tfsdk:"user_id"`
	Status types.Int64  `tfsdk:"status"`
	SKUIDs types.List   `tfsdk:"sku_ids"`
}

func (d *subscriptionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subscription"
}

func (d *subscriptionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists subscriptions containing a SKU, filtered by user. Discord requires `user_id` unless querying with an OAuth token.",
		Attributes: map[string]schema.Attribute{
			"sku_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Snowflake ID of the SKU.",
				Validators:          []validator.String{validators.Snowflake()},
			},
			"user_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter to subscriptions for this user ID (required with bot/bearer auth).",
				Validators:          []validator.String{validators.Snowflake()},
			},
			"subscriptions": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The matching subscriptions.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":      schema.StringAttribute{Computed: true, MarkdownDescription: "Subscription snowflake ID."},
						"user_id": schema.StringAttribute{Computed: true, MarkdownDescription: "User the subscription belongs to."},
						"status":  schema.Int64Attribute{Computed: true, MarkdownDescription: "Subscription status."},
						"sku_ids": schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "SKU IDs included in the subscription."},
					},
				},
			},
		},
	}
}

func (d *subscriptionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configure(req.ProviderData, resp.Diagnostics.AddError)
}

func (d *subscriptionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data subscriptionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	subs, err := ListSubscriptions(ctx, d.client, data.SKUID.ValueString(), data.UserID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(diagutil.APIError("listing Discord subscriptions", data.SKUID.ValueString(), err))
		return
	}
	data.Subscriptions = make([]subscriptionEntryModel, 0, len(subs))
	for _, s := range subs {
		ids, diags := types.ListValueFrom(ctx, types.StringType, s.SKUIDs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Subscriptions = append(data.Subscriptions, subscriptionEntryModel{
			ID:     types.StringValue(s.ID),
			UserID: types.StringValue(s.UserID),
			Status: types.Int64Value(s.Status),
			SKUIDs: ids,
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
