// Package validators provides Terraform Plugin Framework schema validators that
// enforce documented Discord constraints (snowflake shape, permission names)
// before a request is ever sent.
package validators

import (
	"context"
	"fmt"

	"github.com/alpaca744/terraform-provider-discord/internal/discord"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Snowflake validates that a string attribute is a Discord snowflake.
func Snowflake() validator.String {
	return snowflakeValidator{}
}

type snowflakeValidator struct{}

func (snowflakeValidator) Description(_ context.Context) string {
	return "value must be a Discord snowflake (a numeric ID string)"
}

func (v snowflakeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (snowflakeValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if !discord.IsSnowflake(req.ConfigValue.ValueString()) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Discord snowflake",
			fmt.Sprintf("%q is not a valid Discord snowflake; expected a numeric ID string.", req.ConfigValue.ValueString()),
		)
	}
}

// PermissionName validates that a string is a documented Discord permission.
func PermissionName() validator.String {
	return permissionNameValidator{}
}

type permissionNameValidator struct{}

func (permissionNameValidator) Description(_ context.Context) string {
	return "value must be a documented Discord permission name (e.g. VIEW_CHANNEL)"
}

func (v permissionNameValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (permissionNameValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if !discord.IsPermissionName(req.ConfigValue.ValueString()) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Unknown Discord permission",
			fmt.Sprintf("%q is not a documented Discord permission name.", req.ConfigValue.ValueString()),
		)
	}
}

// ImageDataURI validates that a string is a Discord image data URI.
func ImageDataURI() validator.String {
	return imageDataURIValidator{}
}

type imageDataURIValidator struct{}

func (imageDataURIValidator) Description(_ context.Context) string {
	return "value must be an image data URI (data:image/png;base64,...) using PNG, JPEG, or GIF"
}

func (v imageDataURIValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (imageDataURIValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if !discord.IsImageDataURI(req.ConfigValue.ValueString()) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid image data URI",
			"value must be a data URI of the form data:image/png;base64,... using PNG, JPEG, or GIF.",
		)
	}
}

// AudioDataURI validates that a string is a Discord soundboard audio data URI.
func AudioDataURI() validator.String {
	return audioDataURIValidator{}
}

type audioDataURIValidator struct{}

func (audioDataURIValidator) Description(_ context.Context) string {
	return "value must be an audio data URI (data:audio/mpeg;base64,... or data:audio/ogg;...)"
}

func (v audioDataURIValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (audioDataURIValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if !discord.IsAudioDataURI(req.ConfigValue.ValueString()) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid audio data URI",
			"value must be a data URI of the form data:audio/mpeg;base64,... or data:audio/ogg;base64,...",
		)
	}
}
