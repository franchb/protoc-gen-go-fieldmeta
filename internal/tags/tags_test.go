package tags

import "testing"

func TestParseValue(t *testing.T) {
	tests := []struct {
		raw, key, want string
	}{
		{`validate:"required" yaml:"email"`, "validate", "required"},
		{`validate:"required" yaml:"email"`, "yaml", "email"},
		{`validate:"required"`, "missing", ""},
		{``, "any", ""},
		{`validate:"required,email"`, "validate", "required,email"},
		{`longvalidate:"wrong" validate:"right"`, "validate", "right"},
		{`a:"1" b:"2" c:"3"`, "b", "2"},
		{`key:"value with spaces"`, "key", "value with spaces"},
	}
	for _, tt := range tests {
		got := ParseValue(tt.raw, tt.key)
		if got != tt.want {
			t.Errorf("ParseValue(%q, %q) = %q, want %q", tt.raw, tt.key, got, tt.want)
		}
	}
}
