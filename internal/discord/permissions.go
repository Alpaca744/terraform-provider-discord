// Package discord contains Discord API domain primitives that are independent of
// the Terraform plugin framework: permission bitfields, snowflake helpers, and
// image-data encoding. Keeping them here lets both the API client and the
// framework type layer reuse the same conversions.
package discord

import (
	"fmt"
	"math/bits"
	"sort"
	"strconv"
)

// Permission is a single Discord permission flag.
//
// Discord serializes the combined permission set as a decimal string in API v8+
// (see https://discord.com/developers/docs/topics/permissions). Each flag is a
// distinct bit; the names below match Discord's documented constants exactly so
// they can be surfaced directly in Terraform configuration.
type Permission uint64

// Permission flag values. Source: Discord permissions reference.
const (
	PermCreateInstantInvite     Permission = 1 << 0
	PermKickMembers             Permission = 1 << 1
	PermBanMembers              Permission = 1 << 2
	PermAdministrator           Permission = 1 << 3
	PermManageChannels          Permission = 1 << 4
	PermManageGuild             Permission = 1 << 5
	PermAddReactions            Permission = 1 << 6
	PermViewAuditLog            Permission = 1 << 7
	PermPrioritySpeaker         Permission = 1 << 8
	PermStream                  Permission = 1 << 9
	PermViewChannel             Permission = 1 << 10
	PermSendMessages            Permission = 1 << 11
	PermSendTTSMessages         Permission = 1 << 12
	PermManageMessages          Permission = 1 << 13
	PermEmbedLinks              Permission = 1 << 14
	PermAttachFiles             Permission = 1 << 15
	PermReadMessageHistory      Permission = 1 << 16
	PermMentionEveryone         Permission = 1 << 17
	PermUseExternalEmojis       Permission = 1 << 18
	PermViewGuildInsights       Permission = 1 << 19
	PermConnect                 Permission = 1 << 20
	PermSpeak                   Permission = 1 << 21
	PermMuteMembers             Permission = 1 << 22
	PermDeafenMembers           Permission = 1 << 23
	PermMoveMembers             Permission = 1 << 24
	PermUseVAD                  Permission = 1 << 25
	PermChangeNickname          Permission = 1 << 26
	PermManageNicknames         Permission = 1 << 27
	PermManageRoles             Permission = 1 << 28
	PermManageWebhooks          Permission = 1 << 29
	PermManageGuildExpressions  Permission = 1 << 30
	PermUseApplicationCommands  Permission = 1 << 31
	PermRequestToSpeak          Permission = 1 << 32
	PermManageEvents            Permission = 1 << 33
	PermManageThreads           Permission = 1 << 34
	PermCreatePublicThreads     Permission = 1 << 35
	PermCreatePrivateThreads    Permission = 1 << 36
	PermUseExternalStickers     Permission = 1 << 37
	PermSendMessagesInThreads   Permission = 1 << 38
	PermUseEmbeddedActivities   Permission = 1 << 39
	PermModerateMembers         Permission = 1 << 40
	PermViewCreatorMonetization Permission = 1 << 41
	PermUseSoundboard           Permission = 1 << 42
	PermCreateGuildExpressions  Permission = 1 << 43
	PermCreateEvents            Permission = 1 << 44
	PermUseExternalSounds       Permission = 1 << 45
	PermSendVoiceMessages       Permission = 1 << 46
	PermSendPolls               Permission = 1 << 49
	PermUseExternalApps         Permission = 1 << 50
)

// permissionNames maps each documented Discord permission constant to its flag.
var permissionNames = map[string]Permission{
	"CREATE_INSTANT_INVITE":               PermCreateInstantInvite,
	"KICK_MEMBERS":                        PermKickMembers,
	"BAN_MEMBERS":                         PermBanMembers,
	"ADMINISTRATOR":                       PermAdministrator,
	"MANAGE_CHANNELS":                     PermManageChannels,
	"MANAGE_GUILD":                        PermManageGuild,
	"ADD_REACTIONS":                       PermAddReactions,
	"VIEW_AUDIT_LOG":                      PermViewAuditLog,
	"PRIORITY_SPEAKER":                    PermPrioritySpeaker,
	"STREAM":                              PermStream,
	"VIEW_CHANNEL":                        PermViewChannel,
	"SEND_MESSAGES":                       PermSendMessages,
	"SEND_TTS_MESSAGES":                   PermSendTTSMessages,
	"MANAGE_MESSAGES":                     PermManageMessages,
	"EMBED_LINKS":                         PermEmbedLinks,
	"ATTACH_FILES":                        PermAttachFiles,
	"READ_MESSAGE_HISTORY":                PermReadMessageHistory,
	"MENTION_EVERYONE":                    PermMentionEveryone,
	"USE_EXTERNAL_EMOJIS":                 PermUseExternalEmojis,
	"VIEW_GUILD_INSIGHTS":                 PermViewGuildInsights,
	"CONNECT":                             PermConnect,
	"SPEAK":                               PermSpeak,
	"MUTE_MEMBERS":                        PermMuteMembers,
	"DEAFEN_MEMBERS":                      PermDeafenMembers,
	"MOVE_MEMBERS":                        PermMoveMembers,
	"USE_VAD":                             PermUseVAD,
	"CHANGE_NICKNAME":                     PermChangeNickname,
	"MANAGE_NICKNAMES":                    PermManageNicknames,
	"MANAGE_ROLES":                        PermManageRoles,
	"MANAGE_WEBHOOKS":                     PermManageWebhooks,
	"MANAGE_GUILD_EXPRESSIONS":            PermManageGuildExpressions,
	"USE_APPLICATION_COMMANDS":            PermUseApplicationCommands,
	"REQUEST_TO_SPEAK":                    PermRequestToSpeak,
	"MANAGE_EVENTS":                       PermManageEvents,
	"MANAGE_THREADS":                      PermManageThreads,
	"CREATE_PUBLIC_THREADS":               PermCreatePublicThreads,
	"CREATE_PRIVATE_THREADS":              PermCreatePrivateThreads,
	"USE_EXTERNAL_STICKERS":               PermUseExternalStickers,
	"SEND_MESSAGES_IN_THREADS":            PermSendMessagesInThreads,
	"USE_EMBEDDED_ACTIVITIES":             PermUseEmbeddedActivities,
	"MODERATE_MEMBERS":                    PermModerateMembers,
	"VIEW_CREATOR_MONETIZATION_ANALYTICS": PermViewCreatorMonetization,
	"USE_SOUNDBOARD":                      PermUseSoundboard,
	"CREATE_GUILD_EXPRESSIONS":            PermCreateGuildExpressions,
	"CREATE_EVENTS":                       PermCreateEvents,
	"USE_EXTERNAL_SOUNDS":                 PermUseExternalSounds,
	"SEND_VOICE_MESSAGES":                 PermSendVoiceMessages,
	"SEND_POLLS":                          PermSendPolls,
	"USE_EXTERNAL_APPS":                   PermUseExternalApps,
}

// nameByPermission is the reverse lookup, built once at init.
var nameByPermission = func() map[Permission]string {
	m := make(map[Permission]string, len(permissionNames))
	for name, flag := range permissionNames {
		m[flag] = name
	}
	return m
}()

// PermissionNames returns the sorted list of every known Discord permission name.
func PermissionNames() []string {
	names := make([]string, 0, len(permissionNames))
	for name := range permissionNames {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// IsPermissionName reports whether name is a documented Discord permission.
func IsPermissionName(name string) bool {
	_, ok := permissionNames[name]
	return ok
}

// PermissionsToBitfield converts human-readable permission names into the decimal
// string bitfield Discord expects. Unknown names return an error naming the
// offending value so diagnostics can be specific.
func PermissionsToBitfield(names []string) (string, error) {
	var bitfield Permission
	for _, name := range names {
		flag, ok := permissionNames[name]
		if !ok {
			return "", fmt.Errorf("unknown Discord permission %q", name)
		}
		bitfield |= flag
	}
	return strconv.FormatUint(uint64(bitfield), 10), nil
}

// BitfieldToPermissions converts a Discord decimal permission string into a
// sorted list of known permission names. Bits that do not map to a documented
// permission are ignored rather than erroring, so newly added Discord flags do
// not break reads.
func BitfieldToPermissions(bitfield string) ([]string, error) {
	if bitfield == "" {
		return []string{}, nil
	}
	value, err := strconv.ParseUint(bitfield, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid Discord permission bitfield %q: %w", bitfield, err)
	}
	names := make([]string, 0, bits.OnesCount64(value))
	for flag := Permission(1); flag != 0; flag <<= 1 {
		if value&uint64(flag) == 0 {
			continue
		}
		if name, ok := nameByPermission[flag]; ok {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names, nil
}
