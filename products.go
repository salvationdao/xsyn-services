package passport

import "time"

// Product is an object representing the database table.
type Product struct {
	ID          ProductID  `json:"id" db:"id"`
	Slug        string     `json:"slug" db:"slug"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description" db:"description"`
	ImageID     *BlobID    `json:"imageID" db:"image_id"`
	CreatedAt   time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time  `json:"updatedAt" db:"updated_at"`
	DeletedAt   *time.Time `json:"deletedAt" db:"deleted_at"`
}
