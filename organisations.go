package passport

import "time"

// Organisation is an object representing the database table.
type Organisation struct {
	ID        OrganisationID `json:"id" db:"id"`
	Slug      string         `json:"slug" db:"slug"`
	Name      string         `json:"name" db:"name"`
	CreatedAt time.Time      `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time      `json:"updatedAt" db:"updated_at"`
	DeletedAt *time.Time     `json:"deletedAt" db:"deleted_at"`
}
