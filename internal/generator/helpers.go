package generator

import (
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
)

// emitMessageHelpers generates all helper functions for a single annotated message.
func emitMessageHelpers(g *protogen.GeneratedFile, mm messageMeta) {
	msgName := mm.Message.GoIdent.GoName

	logFields := filter(mm.Fields, func(fm fieldMeta) bool { return fm.Log != "" })
	if len(logFields) > 0 {
		emitLogFields(g, msgName, mm.Message, logFields)
	}

	sensitiveFields := filter(mm.Fields, func(fm fieldMeta) bool { return fm.Sensitive })
	if len(sensitiveFields) > 0 {
		emitSensitiveFields(g, msgName, sensitiveFields)
		emitRedactSensitive(g, msgName, mm.Message, sensitiveFields)
	}

	dbFields := filter(mm.Fields, func(fm fieldMeta) bool { return fm.Db != "" })
	if len(dbFields) > 0 {
		emitDbColumns(g, msgName, dbFields)
	}

	tagFields := filter(mm.Fields, func(fm fieldMeta) bool { return fm.Tags != "" })
	if len(tagFields) > 0 {
		emitFieldTags(g, msgName, tagFields)
		emitGetFieldTag(g, msgName, tagFields)
	}
}

// emitLogFields generates LogFields_<Msg>(msg) map[string]any.
func emitLogFields(g *protogen.GeneratedFile, msgName string, msg *protogen.Message, fields []fieldMeta) {
	g.P("// LogFields_", msgName, " returns a map of structured logging keys to field values.")
	g.P("func LogFields_", msgName, "(msg *", msg.GoIdent, ") map[string]any {")
	g.P("	if msg == nil {")
	g.P("		return nil")
	g.P("	}")
	g.P("	m := msg.ProtoReflect()")
	g.P("	fds := m.Descriptor().Fields()")
	g.P("	result := make(map[string]any, ", len(fields), ")")
	for _, fm := range fields {
		protoName := string(fm.Field.Desc.Name())
		g.P(fmt.Sprintf("	result[%q] = m.Get(fds.ByName(%q)).Interface()", fm.Log, protoName))
	}
	g.P("	return result")
	g.P("}")
	g.P()
}

// emitSensitiveFields generates SensitiveFieldNames_<Msg>() []string.
func emitSensitiveFields(g *protogen.GeneratedFile, msgName string, fields []fieldMeta) {
	g.P("// SensitiveFieldNames_", msgName, " returns the proto field names marked as sensitive.")
	g.P("func SensitiveFieldNames_", msgName, "() []string {")
	g.P("	return []string{")
	for _, fm := range fields {
		g.P(fmt.Sprintf("		%q,", fm.Field.Desc.Name()))
	}
	g.P("	}")
	g.P("}")
	g.P()
}

// emitRedactSensitive generates RedactSensitive_<Msg>(msg) *<Msg>.
func emitRedactSensitive(g *protogen.GeneratedFile, msgName string, msg *protogen.Message, fields []fieldMeta) {
	protoClone := g.QualifiedGoIdent(protogen.GoIdent{
		GoName:       "Clone",
		GoImportPath: protoPackage,
	})
	g.P("// RedactSensitive_", msgName, " returns a deep copy with sensitive fields cleared.")
	g.P("func RedactSensitive_", msgName, "(msg *", msg.GoIdent, ") *", msg.GoIdent, " {")
	g.P("	if msg == nil {")
	g.P("		return nil")
	g.P("	}")
	g.P("	clone := ", protoClone, "(msg).(*", msg.GoIdent, ")")
	g.P("	m := clone.ProtoReflect()")
	g.P("	fds := m.Descriptor().Fields()")
	for _, fm := range fields {
		protoName := string(fm.Field.Desc.Name())
		g.P(fmt.Sprintf("	m.Clear(fds.ByName(%q))", protoName))
	}
	g.P("	return clone")
	g.P("}")
	g.P()
}

// emitDbColumns generates FieldDbColumns_<Msg>() map[string]string.
func emitDbColumns(g *protogen.GeneratedFile, msgName string, fields []fieldMeta) {
	g.P("// FieldDbColumns_", msgName, " returns a map of proto field name to DB column name.")
	g.P("func FieldDbColumns_", msgName, "() map[string]string {")
	g.P("	return map[string]string{")
	for _, fm := range fields {
		g.P(fmt.Sprintf("		%q: %q,", fm.Field.Desc.Name(), fm.Db))
	}
	g.P("	}")
	g.P("}")
	g.P()
}

// emitFieldTags generates FieldTags_<Msg>() map[string]string.
func emitFieldTags(g *protogen.GeneratedFile, msgName string, fields []fieldMeta) {
	g.P("// FieldTags_", msgName, " returns a map of proto field name to raw struct tag string.")
	g.P("func FieldTags_", msgName, "() map[string]string {")
	g.P("	return map[string]string{")
	for _, fm := range fields {
		g.P(fmt.Sprintf("		%q: %q,", fm.Field.Desc.Name(), fm.Tags))
	}
	g.P("	}")
	g.P("}")
	g.P()
}

// emitGetFieldTag generates GetFieldTag_<Msg>(protoFieldName, tagKey) string.
func emitGetFieldTag(g *protogen.GeneratedFile, msgName string, fields []fieldMeta) {
	g.P("// GetFieldTag_", msgName, " returns a single struct tag value for the given field and tag key.")
	g.P("func GetFieldTag_", msgName, "(protoFieldName, tagKey string) string {")
	g.P("	tags := FieldTags_", msgName, "()")
	g.P("	raw, ok := tags[protoFieldName]")
	g.P("	if !ok {")
	g.P("		return \"\"")
	g.P("	}")
	g.P("	return parseTagValue_fieldmeta(raw, tagKey)")
	g.P("}")
	g.P()
}

// emitTagParserHelpers emits file-level parseTagValue_fieldmeta and indexOf_fieldmeta helpers.
// These are emitted once per file to avoid redeclaration.
func emitTagParserHelpers(g *protogen.GeneratedFile) {
	g.P("// parseTagValue_fieldmeta extracts a single tag value from a raw struct tag string.")
	g.P("func parseTagValue_fieldmeta(raw, key string) string {")
	g.P(`	search := key + ":\"" `)
	g.P("	idx := indexOf_fieldmeta(raw, search)")
	g.P("	if idx < 0 {")
	g.P("		return \"\"")
	g.P("	}")
	g.P("	// Ensure full key match (not a suffix of another key).")
	g.P("	for idx > 0 && raw[idx-1] != ' ' {")
	g.P("		next := indexOf_fieldmeta(raw[idx+1:], search)")
	g.P("		if next < 0 {")
	g.P("			return \"\"")
	g.P("		}")
	g.P("		idx = idx + 1 + next")
	g.P("	}")
	g.P("	start := idx + len(search)")
	g.P(`	end := indexOf_fieldmeta(raw[start:], "\"")`)
	g.P("	if end < 0 {")
	g.P("		return \"\"")
	g.P("	}")
	g.P("	return raw[start : start+end]")
	g.P("}")
	g.P()
	g.P("// indexOf_fieldmeta returns the index of substr in s, or -1.")
	g.P("func indexOf_fieldmeta(s, substr string) int {")
	g.P("	for i := 0; i+len(substr) <= len(s); i++ {")
	g.P("		if s[i:i+len(substr)] == substr {")
	g.P("			return i")
	g.P("		}")
	g.P("	}")
	g.P("	return -1")
	g.P("}")
	g.P()
}
