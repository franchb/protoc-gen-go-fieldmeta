// Package fieldmetautil provides generic, reflection-based access to fieldmeta
// options at runtime. Unlike the generated per-message helpers, this library
// works on any proto.Message using protoreflect.
package fieldmetautil

import (
	fieldmetav1 "github.com/franchb/protoc-gen-go-fieldmeta/fieldmeta/v1"
	"github.com/franchb/protoc-gen-go-fieldmeta/internal/tags"
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
	if fd == nil {
		return ""
	}
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

// parseRawTag delegates to the shared tags package.
func parseRawTag(raw, key string) string {
	return tags.ParseValue(raw, key)
}
