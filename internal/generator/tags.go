package generator

import "strings"

// parseTagValue extracts a single tag value from a raw struct tag string.
// For example, given raw=`validate:"required" yaml:"email"` and key="validate",
// it returns "required".
func parseTagValue(raw, key string) string {
	search := key + `:"`
	idx := strings.Index(raw, search)
	if idx < 0 {
		return ""
	}
	// Ensure it's a full key match (not a suffix of another key).
	if idx > 0 && raw[idx-1] != ' ' {
		// Try finding the next occurrence.
		rest := raw[idx+1:]
		for {
			next := strings.Index(rest, search)
			if next < 0 {
				return ""
			}
			absIdx := idx + 1 + next
			if absIdx == 0 || raw[absIdx-1] == ' ' {
				idx = absIdx
				break
			}
			rest = rest[next+1:]
			idx = absIdx
		}
	}
	start := idx + len(search)
	end := strings.Index(raw[start:], `"`)
	if end < 0 {
		return ""
	}
	return raw[start : start+end]
}

// indexOf returns the index of substr in s, or -1.
func indexOf(s, substr string) int {
	return strings.Index(s, substr)
}
