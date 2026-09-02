package models

import "time"

// ManufactureTable .
var ManufactureTable string = "manufactures"

// Manufacture identifies the producer of a tobacco product.
type Manufacture struct {
	ID        string     `db:"id"`
	Name      string     `db:"name"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}
