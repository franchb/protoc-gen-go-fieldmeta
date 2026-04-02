// Package fieldmetautil provides generic, reflection-based access to fieldmeta
// options at runtime. Unlike the generated per-message helpers, this library
// works on any proto.Message using protoreflect.
package fieldmetautil

import (
	"strings"

	fieldmetav1 "github.com/franchb/protoc-gen-go-fieldmeta/fieldmeta/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// LogFields returns a map of structured logging keys to field values,
// reading the fieldmeta.log extension from each field's options.
func LogFields(msg proto.Message) map[string]any {
	if msg == nil {
		return nil
	}
	m := msg.ProtoReflect()
	fds := m.Descriptor().Fields()
	var result map[string]any
	for i := 0; i < fds.Len(); i++ {
		fd := fds.Get(i)
		logKey := getStringOpt(fd, fieldmetav1.E_Log)
		if logKey == "" {
			continue
		}
		if result == nil {
			result = make(map[string]any)
		}
		result[logKey] = m.Get(fd).Interface()
	}
	return result
}

// SensitiveFieldNames returns the proto field names marked as sensitive.
func SensitiveFieldNames(msg proto.Message) []string {
	if msg == nil {
		return nil
	}
	m := msg.ProtoReflect()
	fds := m.Descriptor().Fields()
	var names []string
	for i := 0; i < fds.Len(); i++ {
		fd := fds.Get(i)
		if getBoolOpt(fd, fieldmetav1.E_Sensitive) {
			names = append(names, string(fd.Name()))
		}
	}
	return names
}

// RedactSensitive returns a deep copy with sensitive fields cleared.
func RedactSensitive(msg proto.Message) proto.Message {
	if msg == nil {
		return nil
	}
	clone := proto.Clone(msg)
	m := clone.ProtoReflect()
	fds := m.Descriptor().Fields()
	for i := 0; i < fds.Len(); i++ {
		fd := fds.Get(i)
		if getBoolOpt(fd, fieldmetav1.E_Sensitive) {
			m.Clear(fd)
		}
	}
	return clone
}

// GetTagValue extracts a single tag value from a field's fieldmeta.tags option.
func GetTagValue(fd protoreflect.FieldDescriptor, key string) string {
	raw := getStringOpt(fd, fieldmetav1.E_Tags)
	if raw == "" {
		return ""
	}
	return parseRawTag(raw, key)
}

// getStringOpt reads a string extension from a field descriptor's options.
// Returns "" if not set or options are nil.
func getStringOpt(fd protoreflect.FieldDescriptor, ext protoreflect.ExtensionType) string {
	opts, ok := fd.Options().(*descriptorpb.FieldOptions)
	if !ok || opts == nil {
		return ""
	}
	if !proto.HasExtension(opts, ext) {
		return ""
	}
	return proto.GetExtension(opts, ext).(string)
}

// getBoolOpt reads a bool extension from a field descriptor's options.
// Returns false if not set or options are nil.
func getBoolOpt(fd protoreflect.FieldDescriptor, ext protoreflect.ExtensionType) bool {
	opts, ok := fd.Options().(*descriptorpb.FieldOptions)
	if !ok || opts == nil {
		return false
	}
	if !proto.HasExtension(opts, ext) {
		return false
	}
	return proto.GetExtension(opts, ext).(bool)
}

// parseRawTag extracts a single tag value from a raw struct tag string.
// For example, given raw=`validate:"required" yaml:"email"` and key="validate",
// it returns "required".
func parseRawTag(raw, key string) string {
	search := key + `:"`
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
