package cli

import "strings"

// cutIconPrefix strips the "Interface\Icons\" DBC path prefix case-insensitively.
func cutIconPrefix(s string) string {
	const prefix = `Interface\Icons\`
	if len(s) >= len(prefix) && strings.EqualFold(s[:len(prefix)], prefix) {
		return s[len(prefix):]
	}
	return s
}
