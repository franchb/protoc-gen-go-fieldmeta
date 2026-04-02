// Package tags provides shared struct tag parsing logic used by both the
// generator (internal/generator) and the runtime library (fieldmetautil).
package tags

import "strings"

// ParseValue extracts a single tag value from a raw struct tag string.
// For example, given raw=`validate:"required" yaml:"email"` and key="validate",
// it returns "required".
func ParseValue(raw, key string) string {
	search := key + `:"` //nolint:gocritic
	idx := strings.Index(raw, search)
	if idx < 0 {
		return ""
	}
	// Ensure it's a full key match (not a suffix of another key).
	for idx > 0 && raw[idx-1] != ' ' {
		next := strings.Index(raw[idx+1:], search)
		if next < 0 {
			return ""
		}
		idx = idx + 1 + next
	}
	start := idx + len(search)
	end := strings.Index(raw[start:], `"`)
	if end < 0 {
		return ""
	}
	return raw[start : start+end]
}
