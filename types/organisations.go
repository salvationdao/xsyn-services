package types

import "time"

// Organisation is an object representing the database table.
type Organisation struct {
	ID        OrganisationID `json:"id" db:"id"`
	Slug      string         `json:"slug" db:"slug"`
	Name      string         `json:"name" db:"name"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time     `json:"deleted_at" db:"deleted_at"`
}
