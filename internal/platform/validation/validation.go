// Package validation provides reusable ozzo-validation rules.
package validation

import (
	"strings"

	ozzo "github.com/go-ozzo/ozzo-validation/v4"
)

// RequiredString validates that a value is a string that is non-empty after
// trimming surrounding whitespace. It rejects nil, non-string values and
// blank-like strings such as "   ".
func RequiredString() ozzo.Rule {
	return ozzo.By(func(value any) error {
		s, ok := value.(string)
		if !ok || strings.TrimSpace(s) == "" {
			return ozzo.ErrRequired
		}
		return nil
	})
}
