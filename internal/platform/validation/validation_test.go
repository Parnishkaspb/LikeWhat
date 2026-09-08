package validation

import (
	"testing"

	ozzo "github.com/go-ozzo/ozzo-validation/v4"
)

func TestRequiredString(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  bool // want error?
	}{
		{name: "non-empty", value: "Ozon", want: false},
		{name: "empty", value: "", want: true},
		{name: "whitespace only", value: "   \t\n ", want: true},
		{name: "surrounded by spaces", value: "  Ozon  ", want: false},
		{name: "nil", value: nil, want: true},
		{name: "non-string", value: 42, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ozzo.Validate(tt.value, RequiredString())
			if got := err != nil; got != tt.want {
				t.Fatalf("RequiredString(%v) error = %v, want error: %t", tt.value, err, tt.want)
			}
		})
	}
}
