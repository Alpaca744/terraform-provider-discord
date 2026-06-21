package discord

import (
	"strconv"
)

// IsSnowflake reports whether s is a valid Discord snowflake: a non-empty string
// of ASCII digits that fits in a uint64.
//
// Discord returns snowflakes as strings in JSON because they are up to 64 bits
// and would overflow some languages' integers
// (https://discord.com/developers/docs/reference#snowflakes). The provider keeps
// them as strings everywhere and only validates shape, never numeric magnitude
// beyond the uint64 bound.
func IsSnowflake(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	// Reject leading-zero forms and values that overflow uint64.
	if len(s) > 1 && s[0] == '0' {
		return false
	}
	if _, err := strconv.ParseUint(s, 10, 64); err != nil {
		return false
	}
	return true
}
