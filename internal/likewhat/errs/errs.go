// Package errs holds sentinel errors shared across all likewhat domains.
package errs

import "errors"

// ErrNotFound is returned when an entity does not exist, regardless of domain.
var ErrNotFound = errors.New("entity not found")
