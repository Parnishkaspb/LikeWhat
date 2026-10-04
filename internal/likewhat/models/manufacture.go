package models

import "time"

// Table and column names of the manufactures table. They live next to the
// domain model so repositories share a single source of truth for SQL fragments.
const (
	TableManufactures = "manufactures"

	ColManufactureID        = "id"
	ColManufactureName      = "name"
	ColManufactureCreatedAt = "created_at"
	ColManufactureUpdatedAt = "updated_at"
	ColManufactureDeletedAt = "deleted_at"
)

// Manufacture is the domain model identifying the producer of a tobacco product.
// It deliberately does not depend on protobuf.
type Manufacture struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// ManufactureFilter limits and paginates manufacture records returned by a repository.
type ManufactureFilter struct {
	NameLike string
	IDs      []string
	Page     uint64
	PerPage  uint64
}
