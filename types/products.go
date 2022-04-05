package types

import "time"

// Product is an object representing the database table.
type Product struct {
	ID          ProductID  `json:"id" db:"id"`
	Slug        string     `json:"slug" db:"slug"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description" db:"description"`
	ImageID     *BlobID    `json:"image_id" db:"image_id"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at" db:"deleted_at"`
}
