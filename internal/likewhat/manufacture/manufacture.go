package manufacture

import "time"

// Manufacture is the domain model identifying the producer of a tobacco product.
// It deliberately does not depend on protobuf.
type Manufacture struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// ListFilter limits and paginates manufacture records returned by a repository.
type ListFilter struct {
	NameLike string
	IDs      []string
	Page     uint64
	PerPage  uint64
}
