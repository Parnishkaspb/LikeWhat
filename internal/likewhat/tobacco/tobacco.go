package tobacco

import "time"

// Manufacture is a minimal sub-object of a tobacco product identifying its producer.
type Manufacture struct {
	ID   string
	Name string
}

// Tobacco is the domain model. It deliberately does not depend on protobuf.
type Tobacco struct {
	ID          string
	Taste       string
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
