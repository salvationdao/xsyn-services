package passport

import (
	"time"
)

// Asset is a single xsyn_asset on the platform
type Asset struct {
	TokenID     int        `json:"token_id" db:"token_id"`
	Name        string     `json:"name" db:"name"`
	Collection  string     `json:"collection" db:"collection"`
	Description string     `json:"description" db:"description"`
	ExternalUrl string     `json:"externalURL" db:"external_url"`
	Image       string     `json:"image" db:"image"`
	Attributes  string     `json:"attributes" db:"attributes"`
	CreatedAt   time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time  `json:"updatedAt" db:"updated_at"`
	DeletedAt   *time.Time `json:"deletedAt" db:"deleted_at"`
}
