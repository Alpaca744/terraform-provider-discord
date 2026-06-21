package automod

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// expand turns the Terraform plan into the API representation of trigger
// metadata, actions, and the exempt-ID slices.
func (r *ruleResource) expand(ctx context.Context, plan *ruleModel) (*TriggerMetadata, []Action, []string, []string, diag.Diagnostics) {
	var diags diag.Diagnostics

	exemptRoles, d := expandStringSet(ctx, plan.ExemptRoles)
	diags.Append(d...)
	exemptChannels, d := expandStringSet(ctx, plan.ExemptChannels)
	diags.Append(d...)

	var tm *TriggerMetadata
	if plan.TriggerMetadata != nil {
		m := plan.TriggerMetadata
		tm = &TriggerMetadata{
			MentionTotalLimit:            m.MentionTotalLimit.ValueInt64(),
			MentionRaidProtectionEnabled: m.MentionRaidProtectionEnabled.ValueBool(),
		}
		tm.KeywordFilter, d = expandStringSet(ctx, m.KeywordFilter)
		diags.Append(d...)
		tm.RegexPatterns, d = expandStringSet(ctx, m.RegexPatterns)
		diags.Append(d...)
		tm.AllowList, d = expandStringSet(ctx, m.AllowList)
		diags.Append(d...)
		tm.Presets, d = expandInt64Set(ctx, m.Presets)
		diags.Append(d...)
	}

	actions := make([]Action, 0, len(plan.Actions))
	for _, a := range plan.Actions {
		action := Action{Type: a.Type.ValueInt64()}
		if !a.ChannelID.IsNull() || !a.DurationSeconds.IsNull() || !a.CustomMessage.IsNull() {
			action.Metadata = &ActionMetadata{
				ChannelID:       a.ChannelID.ValueString(),
				DurationSeconds: a.DurationSeconds.ValueInt64(),
				CustomMessage:   a.CustomMessage.ValueString(),
			}
		}
		actions = append(actions, action)
	}

	return tm, actions, exemptRoles, exemptChannels, diags
}

// flatten writes the API rule back into the Terraform model.
func (r *ruleResource) flatten(ctx context.Context, m *ruleModel, rule *Rule) diag.Diagnostics {
	var diags diag.Diagnostics

	m.ID = types.StringValue(rule.ID)
	m.GuildID = types.StringValue(rule.GuildID)
	m.Name = types.StringValue(rule.Name)
	m.EventType = types.Int64Value(rule.EventType)
	m.TriggerType = types.Int64Value(rule.TriggerType)
	m.Enabled = types.BoolValue(rule.Enabled)

	roles, d := flattenStringSlice(ctx, rule.ExemptRoles)
	diags.Append(d...)
	channels, d := flattenStringSlice(ctx, rule.ExemptChannels)
	diags.Append(d...)
	m.ExemptRoles = roles
	m.ExemptChannels = channels

	// Only populate trigger_metadata when the rule defines it, so KEYWORD-less
	// trigger types keep the attribute null rather than an empty object.
	if rule.TriggerMetadata != nil && !triggerMetadataEmpty(rule.TriggerMetadata) {
		tm := &triggerMetadataModel{
			MentionTotalLimit:            types.Int64Value(rule.TriggerMetadata.MentionTotalLimit),
			MentionRaidProtectionEnabled: types.BoolValue(rule.TriggerMetadata.MentionRaidProtectionEnabled),
		}
		tm.KeywordFilter, d = optStringSet(ctx, rule.TriggerMetadata.KeywordFilter)
		diags.Append(d...)
		tm.RegexPatterns, d = optStringSet(ctx, rule.TriggerMetadata.RegexPatterns)
		diags.Append(d...)
		tm.AllowList, d = optStringSet(ctx, rule.TriggerMetadata.AllowList)
		diags.Append(d...)
		tm.Presets, d = optInt64Set(ctx, rule.TriggerMetadata.Presets)
		diags.Append(d...)
		m.TriggerMetadata = tm
	} else {
		m.TriggerMetadata = nil
	}

	m.Actions = make([]actionModel, 0, len(rule.Actions))
	for _, a := range rule.Actions {
		am := actionModel{
			Type:            types.Int64Value(a.Type),
			ChannelID:       types.StringNull(),
			DurationSeconds: types.Int64Null(),
			CustomMessage:   types.StringNull(),
		}
		if a.Metadata != nil {
			if a.Metadata.ChannelID != "" {
				am.ChannelID = types.StringValue(a.Metadata.ChannelID)
			}
			if a.Metadata.DurationSeconds != 0 {
				am.DurationSeconds = types.Int64Value(a.Metadata.DurationSeconds)
			}
			if a.Metadata.CustomMessage != "" {
				am.CustomMessage = types.StringValue(a.Metadata.CustomMessage)
			}
		}
		m.Actions = append(m.Actions, am)
	}

	return diags
}

func triggerMetadataEmpty(tm *TriggerMetadata) bool {
	return len(tm.KeywordFilter) == 0 && len(tm.RegexPatterns) == 0 &&
		len(tm.Presets) == 0 && len(tm.AllowList) == 0 &&
		tm.MentionTotalLimit == 0 && !tm.MentionRaidProtectionEnabled
}

func expandStringSet(ctx context.Context, set types.Set) ([]string, diag.Diagnostics) {
	if set.IsNull() || set.IsUnknown() {
		return nil, nil
	}
	var out []string
	d := set.ElementsAs(ctx, &out, false)
	return out, d
}

func expandInt64Set(ctx context.Context, set types.Set) ([]int64, diag.Diagnostics) {
	if set.IsNull() || set.IsUnknown() {
		return nil, nil
	}
	var out []int64
	d := set.ElementsAs(ctx, &out, false)
	return out, d
}

func flattenStringSlice(ctx context.Context, in []string) (types.Set, diag.Diagnostics) {
	if len(in) == 0 {
		return types.SetNull(types.StringType), nil
	}
	return types.SetValueFrom(ctx, types.StringType, in)
}

func optStringSet(ctx context.Context, in []string) (types.Set, diag.Diagnostics) {
	return flattenStringSlice(ctx, in)
}

func optInt64Set(ctx context.Context, in []int64) (types.Set, diag.Diagnostics) {
	if len(in) == 0 {
		return types.SetNull(types.Int64Type), nil
	}
	return types.SetValueFrom(ctx, types.Int64Type, in)
}
