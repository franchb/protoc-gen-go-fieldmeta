package fieldmetautil

import (
	"slices"
	"testing"

	"google.golang.org/protobuf/proto"

	fieldmetav1 "github.com/franchb/protoc-gen-go-fieldmeta/fieldmeta/v1"
	"github.com/franchb/protoc-gen-go-fieldmeta/testdata/testpb"
)

// ---------- LogFields ----------

func TestLogFields_User(t *testing.T) {
	u := &testpb.User{}
	u.Email = proto.String("a@b.com")
	u.DisplayName = proto.String("Alice")
	u.PasswordHash = proto.String("hash")
	u.Ssn = proto.String("123")
	u.Role = proto.String("admin")

	got := LogFields(u)
	if got == nil {
		t.Fatal("LogFields returned nil for non-nil message")
	}

	want := map[string]any{
		"user_email": "a@b.com",
		"name":       "Alice",
	}
	if len(got) != len(want) {
		t.Fatalf("LogFields returned %d entries, want %d: %v", len(got), len(want), got)
	}
	for k, wv := range want {
		gv, ok := got[k]
		if !ok {
			t.Errorf("missing key %q in LogFields result", k)
			continue
		}
		if gv != wv {
			t.Errorf("LogFields[%q] = %v, want %v", k, gv, wv)
		}
	}
}

func TestLogFields_Event(t *testing.T) {
	e := &testpb.Event{}
	e.EventId = proto.String("evt-1")
	e.Payload = &testpb.Event_Text{Text: "hello"}

	got := LogFields(e)
	if got == nil {
		t.Fatal("LogFields returned nil for non-nil Event")
	}
	if got["event_id"] != "evt-1" {
		t.Errorf("LogFields[event_id] = %v, want evt-1", got["event_id"])
	}
	if got["text_payload"] != "hello" {
		t.Errorf("LogFields[text_payload] = %v, want hello", got["text_payload"])
	}
}

func TestLogFields_LegacyUser(t *testing.T) {
	lu := &testpb.LegacyUser{}
	lu.Email = proto.String("x@y.com")
	lu.Age = proto.Int32(30)
	lu.Login = proto.String("mylogin")

	got := LogFields(lu)
	if got == nil {
		t.Fatal("LogFields returned nil for non-nil LegacyUser")
	}
	// Only login has (fieldmeta.v1.log) = "user_login"
	if len(got) != 1 {
		t.Fatalf("LogFields returned %d entries, want 1: %v", len(got), got)
	}
	if got["user_login"] != "mylogin" {
		t.Errorf("LogFields[user_login] = %v, want mylogin", got["user_login"])
	}
}

func TestLogFields_Nil(t *testing.T) {
	got := LogFields(nil)
	if got != nil {
		t.Errorf("LogFields(nil) = %v, want nil", got)
	}
}

func TestLogFields_NoOptions(t *testing.T) {
	p := &testpb.Plain{}
	p.Id = proto.String("1")
	p.Name = proto.String("test")

	got := LogFields(p)
	if got != nil {
		t.Errorf("LogFields on Plain (no options) = %v, want nil", got)
	}
}

func TestLogFields_Empty(t *testing.T) {
	e := &testpb.Empty{}
	got := LogFields(e)
	if got != nil {
		t.Errorf("LogFields on Empty = %v, want nil", got)
	}
}

// ---------- SensitiveFieldNames ----------

func TestSensitiveFieldNames_User(t *testing.T) {
	u := &testpb.User{}
	got := SensitiveFieldNames(u)
	want := []string{"password_hash", "ssn"}
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("SensitiveFieldNames(User) = %v, want %v", got, want)
	}
}

func TestSensitiveFieldNames_Event(t *testing.T) {
	e := &testpb.Event{}
	got := SensitiveFieldNames(e)
	want := []string{"binary"}
	if !slices.Equal(got, want) {
		t.Errorf("SensitiveFieldNames(Event) = %v, want %v", got, want)
	}
}

func TestSensitiveFieldNames_LegacyUser(t *testing.T) {
	lu := &testpb.LegacyUser{}
	got := SensitiveFieldNames(lu)
	want := []string{"login"}
	if !slices.Equal(got, want) {
		t.Errorf("SensitiveFieldNames(LegacyUser) = %v, want %v", got, want)
	}
}

func TestSensitiveFieldNames_Nil(t *testing.T) {
	got := SensitiveFieldNames(nil)
	if got != nil {
		t.Errorf("SensitiveFieldNames(nil) = %v, want nil", got)
	}
}

func TestSensitiveFieldNames_NoOptions(t *testing.T) {
	p := &testpb.Plain{}
	got := SensitiveFieldNames(p)
	if got != nil {
		t.Errorf("SensitiveFieldNames(Plain) = %v, want nil", got)
	}
}

// ---------- RedactSensitive ----------

func TestRedactSensitive_User(t *testing.T) {
	u := &testpb.User{}
	u.Email = proto.String("a@b.com")
	u.DisplayName = proto.String("Alice")
	u.PasswordHash = proto.String("secret-hash")
	u.Ssn = proto.String("123-45-6789")
	u.Role = proto.String("admin")

	result := RedactSensitive(u)
	if result == nil {
		t.Fatal("RedactSensitive returned nil for non-nil message")
	}

	redacted := result.(*testpb.User)

	// Non-sensitive fields preserved.
	if redacted.GetEmail() != "a@b.com" {
		t.Errorf("email = %q, want a@b.com", redacted.GetEmail())
	}
	if redacted.GetDisplayName() != "Alice" {
		t.Errorf("display_name = %q, want Alice", redacted.GetDisplayName())
	}
	if redacted.GetRole() != "admin" {
		t.Errorf("role = %q, want admin", redacted.GetRole())
	}

	// Sensitive fields cleared.
	if redacted.GetPasswordHash() != "" {
		t.Errorf("password_hash = %q, want empty", redacted.GetPasswordHash())
	}
	if redacted.GetSsn() != "" {
		t.Errorf("ssn = %q, want empty", redacted.GetSsn())
	}

	// Original unchanged.
	if u.GetPasswordHash() != "secret-hash" {
		t.Error("original password_hash was mutated")
	}
	if u.GetSsn() != "123-45-6789" {
		t.Error("original ssn was mutated")
	}
}

func TestRedactSensitive_Event(t *testing.T) {
	e := &testpb.Event{}
	e.EventId = proto.String("evt-1")
	e.Payload = &testpb.Event_Binary{Binary: []byte("sensitive-data")}

	result := RedactSensitive(e)
	redacted := result.(*testpb.Event)

	if redacted.GetEventId() != "evt-1" {
		t.Errorf("event_id = %q, want evt-1", redacted.GetEventId())
	}
	if len(redacted.GetBinary()) != 0 {
		t.Errorf("binary = %v, want empty", redacted.GetBinary())
	}
	// Original unchanged.
	if len(e.GetBinary()) == 0 {
		t.Error("original binary was mutated")
	}
}

func TestRedactSensitive_Nil(t *testing.T) {
	got := RedactSensitive(nil)
	if got != nil {
		t.Errorf("RedactSensitive(nil) = %v, want nil", got)
	}
}

func TestRedactSensitive_NoSensitiveFields(t *testing.T) {
	p := &testpb.Plain{}
	p.Id = proto.String("1")
	p.Name = proto.String("test")

	result := RedactSensitive(p)
	redacted := result.(*testpb.Plain)

	if redacted.GetId() != "1" {
		t.Errorf("id = %q, want 1", redacted.GetId())
	}
	if redacted.GetName() != "test" {
		t.Errorf("name = %q, want test", redacted.GetName())
	}
}

// ---------- GetTagValue ----------

func TestGetTagValue_LegacyUser(t *testing.T) {
	lu := &testpb.LegacyUser{}
	fds := lu.ProtoReflect().Descriptor().Fields()

	emailFd := fds.ByName("email")
	if v := GetTagValue(emailFd, "validate"); v != "required,email" {
		t.Errorf("GetTagValue(email, validate) = %q, want %q", v, "required,email")
	}
	if v := GetTagValue(emailFd, "yaml"); v != "email" {
		t.Errorf("GetTagValue(email, yaml) = %q, want %q", v, "email")
	}
	if v := GetTagValue(emailFd, "missing"); v != "" {
		t.Errorf("GetTagValue(email, missing) = %q, want empty", v)
	}

	ageFd := fds.ByName("age")
	if v := GetTagValue(ageFd, "validate"); v != "gte=0,lte=150" {
		t.Errorf("GetTagValue(age, validate) = %q, want %q", v, "gte=0,lte=150")
	}

	// login has no tags extension.
	loginFd := fds.ByName("login")
	if v := GetTagValue(loginFd, "validate"); v != "" {
		t.Errorf("GetTagValue(login, validate) = %q, want empty", v)
	}
}

func TestGetTagValue_NoOptions(t *testing.T) {
	p := &testpb.Plain{}
	fds := p.ProtoReflect().Descriptor().Fields()
	idFd := fds.ByName("id")
	if v := GetTagValue(idFd, "anything"); v != "" {
		t.Errorf("GetTagValue on Plain.id = %q, want empty", v)
	}
}

// ---------- parseRawTag ----------

func TestParseRawTag(t *testing.T) {
	tests := []struct {
		raw, key, want string
	}{
		{`validate:"required" yaml:"email"`, "validate", "required"},
		{`validate:"required" yaml:"email"`, "yaml", "email"},
		{`validate:"required"`, "missing", ""},
		{``, "any", ""},
		{`validate:"required,email"`, "validate", "required,email"},
		{`longvalidate:"wrong" validate:"right"`, "validate", "right"},
	}
	for _, tt := range tests {
		got := parseRawTag(tt.raw, tt.key)
		if got != tt.want {
			t.Errorf("parseRawTag(%q, %q) = %q, want %q", tt.raw, tt.key, got, tt.want)
		}
	}
}

// ---------- internal helpers ----------

func TestGetStringOpt_NilOptions(t *testing.T) {
	// Plain.id has no fieldmeta options at all.
	p := &testpb.Plain{}
	fds := p.ProtoReflect().Descriptor().Fields()
	idFd := fds.ByName("id")

	if v := getStringOpt(idFd, fieldmetav1.E_Log); v != "" {
		t.Errorf("getStringOpt on Plain.id for log = %q, want empty", v)
	}
}

func TestGetBoolOpt_NilOptions(t *testing.T) {
	p := &testpb.Plain{}
	fds := p.ProtoReflect().Descriptor().Fields()
	idFd := fds.ByName("id")

	if v := getBoolOpt(idFd, fieldmetav1.E_Sensitive); v {
		t.Error("getBoolOpt on Plain.id for sensitive = true, want false")
	}
}
