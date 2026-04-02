package testpb

import (
	"sort"
	"testing"

	"google.golang.org/protobuf/proto"
)

// ---------------------------------------------------------------------------
// User: LogFields
// ---------------------------------------------------------------------------

func TestLogFields_User(t *testing.T) {
	msg := &User{
		Email:       proto.String("alice@example.com"),
		DisplayName: proto.String("Alice"),
		Role:        proto.String("admin"),
	}

	got := LogFields_User(msg)

	if len(got) != 2 {
		t.Fatalf("expected 2 keys, got %d: %v", len(got), got)
	}
	if v, ok := got["user_email"]; !ok || v != "alice@example.com" {
		t.Errorf("user_email = %v, want alice@example.com", v)
	}
	if v, ok := got["name"]; !ok || v != "Alice" {
		t.Errorf("name = %v, want Alice", v)
	}
	// role should NOT be in log fields
	if _, ok := got["role"]; ok {
		t.Error("role should not be in LogFields")
	}
}

func TestLogFields_User_Nil(t *testing.T) {
	got := LogFields_User(nil)
	if got != nil {
		t.Errorf("expected nil for nil message, got %v", got)
	}
}

func TestLogFields_User_ZeroValue(t *testing.T) {
	msg := &User{}
	got := LogFields_User(msg)
	if len(got) != 2 {
		t.Fatalf("expected 2 keys for zero-value message, got %d", len(got))
	}
	// Zero-value strings should still appear with their default ("")
	if v, ok := got["user_email"]; !ok {
		t.Error("missing user_email key")
	} else if v != "" {
		t.Errorf("user_email = %v, want empty string", v)
	}
}

// ---------------------------------------------------------------------------
// User: SensitiveFieldNames
// ---------------------------------------------------------------------------

func TestSensitiveFieldNames_User(t *testing.T) {
	got := SensitiveFieldNames_User()
	want := []string{"password_hash", "ssn"}

	if len(got) != len(want) {
		t.Fatalf("SensitiveFieldNames_User() = %v, want %v", got, want)
	}

	sorted := make([]string, len(got))
	copy(sorted, got)
	sort.Strings(sorted)
	sort.Strings(want)
	for i := range want {
		if sorted[i] != want[i] {
			t.Errorf("sorted[%d] = %q, want %q", i, sorted[i], want[i])
		}
	}
}

// ---------------------------------------------------------------------------
// User: RedactSensitive
// ---------------------------------------------------------------------------

func TestRedactSensitive_User(t *testing.T) {
	original := &User{
		Email:        proto.String("alice@example.com"),
		DisplayName:  proto.String("Alice"),
		PasswordHash: proto.String("secret-hash"),
		Ssn:          proto.String("123-45-6789"),
		Role:         proto.String("admin"),
	}

	redacted := RedactSensitive_User(original)

	// Sensitive fields must be cleared
	if redacted.GetPasswordHash() != "" {
		t.Errorf("password_hash not redacted: %q", redacted.GetPasswordHash())
	}
	if redacted.GetSsn() != "" {
		t.Errorf("ssn not redacted: %q", redacted.GetSsn())
	}

	// Non-sensitive fields must be preserved
	if redacted.GetEmail() != "alice@example.com" {
		t.Errorf("email = %q, want alice@example.com", redacted.GetEmail())
	}
	if redacted.GetDisplayName() != "Alice" {
		t.Errorf("display_name = %q, want Alice", redacted.GetDisplayName())
	}
	if redacted.GetRole() != "admin" {
		t.Errorf("role = %q, want admin", redacted.GetRole())
	}

	// Original must be untouched
	if original.GetPasswordHash() != "secret-hash" {
		t.Fatal("original password_hash was modified")
	}
	if original.GetSsn() != "123-45-6789" {
		t.Fatal("original ssn was modified")
	}
}

func TestRedactSensitive_User_Nil(t *testing.T) {
	got := RedactSensitive_User(nil)
	if got != nil {
		t.Errorf("expected nil for nil message, got %v", got)
	}
}

func TestRedactSensitive_User_NoSensitiveSet(t *testing.T) {
	original := &User{
		Email: proto.String("bob@example.com"),
	}
	redacted := RedactSensitive_User(original)
	if redacted.GetEmail() != "bob@example.com" {
		t.Errorf("email = %q, want bob@example.com", redacted.GetEmail())
	}
	// Sensitive fields were never set, redact should still succeed
	if redacted.GetPasswordHash() != "" {
		t.Errorf("password_hash should be empty, got %q", redacted.GetPasswordHash())
	}
}

// ---------------------------------------------------------------------------
// User: FieldDbColumns
// ---------------------------------------------------------------------------

func TestFieldDbColumns_User(t *testing.T) {
	got := FieldDbColumns_User()

	if len(got) != 1 {
		t.Fatalf("expected 1 mapping, got %d: %v", len(got), got)
	}
	if col, ok := got["email"]; !ok || col != "email_address" {
		t.Errorf("email mapping = %q, want email_address", col)
	}
}

// ---------------------------------------------------------------------------
// Outer: LogFields (nested message)
// ---------------------------------------------------------------------------

func TestLogFields_Outer(t *testing.T) {
	msg := &Outer{Name: proto.String("test-outer")}
	got := LogFields_Outer(msg)

	if len(got) != 1 {
		t.Fatalf("expected 1 key, got %d: %v", len(got), got)
	}
	if v, ok := got["outer_name"]; !ok || v != "test-outer" {
		t.Errorf("outer_name = %v, want test-outer", v)
	}
}

func TestLogFields_Outer_Nil(t *testing.T) {
	got := LogFields_Outer(nil)
	if got != nil {
		t.Errorf("expected nil for nil message, got %v", got)
	}
}

// ---------------------------------------------------------------------------
// Outer_Inner: LogFields (nested message type)
// ---------------------------------------------------------------------------

func TestLogFields_Outer_Inner(t *testing.T) {
	msg := &Outer_Inner{Value: proto.Int32(42)}
	got := LogFields_Outer_Inner(msg)

	if len(got) != 1 {
		t.Fatalf("expected 1 key, got %d: %v", len(got), got)
	}
	if v, ok := got["inner_val"]; !ok {
		t.Error("missing inner_val key")
	} else {
		// ProtoReflect returns int32 values; reflect interface comparison
		if v != int32(42) {
			t.Errorf("inner_val = %v (%T), want int32(42)", v, v)
		}
	}
}

func TestLogFields_Outer_Inner_Nil(t *testing.T) {
	got := LogFields_Outer_Inner(nil)
	if got != nil {
		t.Errorf("expected nil for nil message, got %v", got)
	}
}

func TestLogFields_Outer_Inner_ZeroValue(t *testing.T) {
	msg := &Outer_Inner{}
	got := LogFields_Outer_Inner(msg)
	if len(got) != 1 {
		t.Fatalf("expected 1 key, got %d", len(got))
	}
	if v := got["inner_val"]; v != int32(0) {
		t.Errorf("inner_val = %v, want int32(0)", v)
	}
}

// ---------------------------------------------------------------------------
// LegacyUser: LogFields
// ---------------------------------------------------------------------------

func TestLogFields_LegacyUser(t *testing.T) {
	msg := &LegacyUser{
		Email: proto.String("legacy@example.com"),
		Age:   proto.Int32(30),
		Login: proto.String("leguser"),
	}
	got := LogFields_LegacyUser(msg)

	if len(got) != 1 {
		t.Fatalf("expected 1 key, got %d: %v", len(got), got)
	}
	if v, ok := got["user_login"]; !ok || v != "leguser" {
		t.Errorf("user_login = %v, want leguser", v)
	}
}

func TestLogFields_LegacyUser_Nil(t *testing.T) {
	if got := LogFields_LegacyUser(nil); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

// ---------------------------------------------------------------------------
// LegacyUser: SensitiveFieldNames
// ---------------------------------------------------------------------------

func TestSensitiveFieldNames_LegacyUser(t *testing.T) {
	got := SensitiveFieldNames_LegacyUser()
	if len(got) != 1 {
		t.Fatalf("expected 1 sensitive field, got %d: %v", len(got), got)
	}
	if got[0] != "login" {
		t.Errorf("got %q, want login", got[0])
	}
}

// ---------------------------------------------------------------------------
// LegacyUser: RedactSensitive
// ---------------------------------------------------------------------------

func TestRedactSensitive_LegacyUser(t *testing.T) {
	original := &LegacyUser{
		Email: proto.String("legacy@example.com"),
		Age:   proto.Int32(25),
		Login: proto.String("secret-login"),
	}

	redacted := RedactSensitive_LegacyUser(original)

	if redacted.GetLogin() != "" {
		t.Errorf("login not redacted: %q", redacted.GetLogin())
	}
	if redacted.GetEmail() != "legacy@example.com" {
		t.Errorf("email = %q, want legacy@example.com", redacted.GetEmail())
	}
	if redacted.GetAge() != 25 {
		t.Errorf("age = %d, want 25", redacted.GetAge())
	}

	// Original untouched
	if original.GetLogin() != "secret-login" {
		t.Fatal("original login was modified")
	}
}

func TestRedactSensitive_LegacyUser_Nil(t *testing.T) {
	if got := RedactSensitive_LegacyUser(nil); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

// ---------------------------------------------------------------------------
// LegacyUser: FieldTags
// ---------------------------------------------------------------------------

func TestFieldTags_LegacyUser(t *testing.T) {
	got := FieldTags_LegacyUser()

	if len(got) != 2 {
		t.Fatalf("expected 2 tag entries, got %d: %v", len(got), got)
	}

	wantEmail := `validate:"required,email" yaml:"email"`
	if got["email"] != wantEmail {
		t.Errorf("email tag = %q, want %q", got["email"], wantEmail)
	}

	wantAge := `validate:"gte=0,lte=150"`
	if got["age"] != wantAge {
		t.Errorf("age tag = %q, want %q", got["age"], wantAge)
	}

	// login uses structured options, not tags -- should not appear
	if _, ok := got["login"]; ok {
		t.Error("login should not appear in FieldTags (uses structured options)")
	}
}

// ---------------------------------------------------------------------------
// LegacyUser: GetFieldTag
// ---------------------------------------------------------------------------

func TestGetFieldTag_LegacyUser(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		tagKey    string
		want      string
	}{
		{
			name:   "validate tag on email",
			field:  "email",
			tagKey: "validate",
			want:   "required,email",
		},
		{
			name:   "yaml tag on email",
			field:  "email",
			tagKey: "yaml",
			want:   "email",
		},
		{
			name:   "validate tag on age",
			field:  "age",
			tagKey: "validate",
			want:   "gte=0,lte=150",
		},
		{
			name:   "missing tag key on email",
			field:  "email",
			tagKey: "json",
			want:   "",
		},
		{
			name:   "missing field entirely",
			field:  "nonexistent",
			tagKey: "validate",
			want:   "",
		},
		{
			name:   "missing tag key on age",
			field:  "age",
			tagKey: "yaml",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetFieldTag_LegacyUser(tt.field, tt.tagKey)
			if got != tt.want {
				t.Errorf("GetFieldTag_LegacyUser(%q, %q) = %q, want %q",
					tt.field, tt.tagKey, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Event (oneof): LogFields
// ---------------------------------------------------------------------------

func TestLogFields_Event_WithText(t *testing.T) {
	msg := &Event{
		EventId: proto.String("evt-1"),
		Payload: &Event_Text{Text: "hello"},
	}

	got := LogFields_Event(msg)

	if len(got) != 2 {
		t.Fatalf("expected 2 keys, got %d: %v", len(got), got)
	}
	if v := got["event_id"]; v != "evt-1" {
		t.Errorf("event_id = %v, want evt-1", v)
	}
	if v := got["text_payload"]; v != "hello" {
		t.Errorf("text_payload = %v, want hello", v)
	}
}

func TestLogFields_Event_WithBinary(t *testing.T) {
	msg := &Event{
		EventId: proto.String("evt-2"),
		Payload: &Event_Binary{Binary: []byte{0xDE, 0xAD}},
	}

	got := LogFields_Event(msg)

	if len(got) != 2 {
		t.Fatalf("expected 2 keys, got %d: %v", len(got), got)
	}
	if v := got["event_id"]; v != "evt-2" {
		t.Errorf("event_id = %v, want evt-2", v)
	}
	// text_payload key is still present because LogFields emits all annotated fields.
	// When text is not the active oneof, the proto reflection returns the default ("").
	if v := got["text_payload"]; v != "" {
		t.Errorf("text_payload = %v, want empty string (text not set)", v)
	}
}

func TestLogFields_Event_NoPayload(t *testing.T) {
	msg := &Event{
		EventId: proto.String("evt-3"),
	}

	got := LogFields_Event(msg)

	if len(got) != 2 {
		t.Fatalf("expected 2 keys, got %d: %v", len(got), got)
	}
	if v := got["event_id"]; v != "evt-3" {
		t.Errorf("event_id = %v, want evt-3", v)
	}
}

func TestLogFields_Event_Nil(t *testing.T) {
	if got := LogFields_Event(nil); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

// ---------------------------------------------------------------------------
// Event (oneof): SensitiveFieldNames
// ---------------------------------------------------------------------------

func TestSensitiveFieldNames_Event(t *testing.T) {
	got := SensitiveFieldNames_Event()
	if len(got) != 1 {
		t.Fatalf("expected 1, got %d: %v", len(got), got)
	}
	if got[0] != "binary" {
		t.Errorf("got %q, want binary", got[0])
	}
}

// ---------------------------------------------------------------------------
// Event (oneof): RedactSensitive
// ---------------------------------------------------------------------------

func TestRedactSensitive_Event_WithBinary(t *testing.T) {
	original := &Event{
		EventId: proto.String("evt-redact"),
		Payload: &Event_Binary{Binary: []byte("secret-data")},
	}

	redacted := RedactSensitive_Event(original)

	// binary is cleared
	if len(redacted.GetBinary()) != 0 {
		t.Errorf("binary not redacted: %v", redacted.GetBinary())
	}
	// event_id preserved
	if redacted.GetEventId() != "evt-redact" {
		t.Errorf("event_id = %q, want evt-redact", redacted.GetEventId())
	}

	// Original untouched
	if len(original.GetBinary()) == 0 {
		t.Fatal("original binary was modified")
	}
}

func TestRedactSensitive_Event_WithText(t *testing.T) {
	original := &Event{
		EventId: proto.String("evt-safe"),
		Payload: &Event_Text{Text: "public-data"},
	}

	redacted := RedactSensitive_Event(original)

	// text is not sensitive, should be preserved
	if redacted.GetText() != "public-data" {
		t.Errorf("text = %q, want public-data", redacted.GetText())
	}
	if redacted.GetEventId() != "evt-safe" {
		t.Errorf("event_id = %q, want evt-safe", redacted.GetEventId())
	}
}

func TestRedactSensitive_Event_Nil(t *testing.T) {
	if got := RedactSensitive_Event(nil); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

// ---------------------------------------------------------------------------
// RedactSensitive_User is a deep copy (not same pointer)
// ---------------------------------------------------------------------------

func TestRedactSensitive_User_IsCopy(t *testing.T) {
	original := &User{
		Email:        proto.String("test@test.com"),
		PasswordHash: proto.String("hash"),
	}
	redacted := RedactSensitive_User(original)

	if redacted == original {
		t.Error("redacted must be a distinct object (deep copy)")
	}

	// Mutating the redacted copy must not affect the original
	redacted.Email = proto.String("mutated@test.com")
	if original.GetEmail() != "test@test.com" {
		t.Error("mutating redacted copy affected the original")
	}
}

// ---------------------------------------------------------------------------
// FieldDbColumns: ensure non-annotated fields are absent
// ---------------------------------------------------------------------------

func TestFieldDbColumns_User_OnlyAnnotated(t *testing.T) {
	got := FieldDbColumns_User()

	// Fields without db_column annotation should not appear
	for _, absent := range []string{"display_name", "password_hash", "ssn", "role"} {
		if _, ok := got[absent]; ok {
			t.Errorf("field %q should not have a db_column mapping", absent)
		}
	}
}

// ---------------------------------------------------------------------------
// FieldTags stability: calling multiple times returns same data
// ---------------------------------------------------------------------------

func TestFieldTags_LegacyUser_Idempotent(t *testing.T) {
	a := FieldTags_LegacyUser()
	b := FieldTags_LegacyUser()

	if len(a) != len(b) {
		t.Fatalf("length mismatch: %d vs %d", len(a), len(b))
	}
	for k, va := range a {
		if vb, ok := b[k]; !ok || va != vb {
			t.Errorf("mismatch for key %q: %q vs %q", k, va, vb)
		}
	}
}

// ---------------------------------------------------------------------------
// SensitiveFieldNames stability: returns new slice each time
// ---------------------------------------------------------------------------

func TestSensitiveFieldNames_User_ReturnsNewSlice(t *testing.T) {
	a := SensitiveFieldNames_User()
	b := SensitiveFieldNames_User()

	// Mutating one should not affect the other
	a[0] = "MUTATED"
	if b[0] == "MUTATED" {
		t.Error("SensitiveFieldNames should return a fresh slice each call")
	}
}
