package models

import "time"

// Tobacco is the domain model. It deliberately does not depend on protobuf.
type Tobacco struct {
	ID          string
	Taste       string
	Proto       string
	Photo       string
	Manufacture Manufacture
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// ListFilter limits tobacco records returned by a repository.
type ListFilter struct {
	Taste          string
	ManufactureIDs []string
}
