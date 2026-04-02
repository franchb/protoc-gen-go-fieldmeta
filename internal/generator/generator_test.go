package generator

import "testing"

func TestParseTagValue(t *testing.T) {
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
		got := parseTagValue(tt.raw, tt.key)
		if got != tt.want {
			t.Errorf("parseTagValue(%q, %q) = %q, want %q", tt.raw, tt.key, got, tt.want)
		}
	}
}

func TestIndexOf(t *testing.T) {
	if got := indexOf("hello world", "world"); got != 6 {
		t.Errorf("indexOf(hello world, world) = %d, want 6", got)
	}
	if got := indexOf("hello", "xyz"); got != -1 {
		t.Errorf("indexOf(hello, xyz) = %d, want -1", got)
	}
}
