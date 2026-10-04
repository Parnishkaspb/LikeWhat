package models

import "time"

// Table and column names of the tobaccos table. They live next to the domain
// model so repositories share a single source of truth for SQL fragments.
const (
	TableTobaccos = "tobaccos"

	ColTobaccoID            = "id"
	ColTobaccoTaste         = "taste"
	ColTobaccoPhoto         = "photo"
	ColTobaccoManufactureID = "manufacture_id"
	ColTobaccoCreatedAt     = "created_at"
	ColTobaccoUpdatedAt     = "updated_at"
	ColTobaccoDeletedAt     = "deleted_at"
)

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

// TobaccoFilter limits tobacco records returned by a repository.
type TobaccoFilter struct {
	Taste          string
	ManufactureIDs []string
}
