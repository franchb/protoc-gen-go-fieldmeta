package generator

import "github.com/franchb/protoc-gen-go-fieldmeta/internal/tags"

// parseTagValue delegates to the shared tags package.
func parseTagValue(raw, key string) string {
	return tags.ParseValue(raw, key)
}
