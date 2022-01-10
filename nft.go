package passport

import (
	"time"
)

type Attribute struct {
	DisplayType DisplayType `json:"display_type,omitempty"`
	TraitType   string      `json:"trait_type"`
	Value       interface{} `json:"value"` // string or number only
}

type DisplayType string

const (
	BoostNumber     DisplayType = "boost_number"
	BoostPercentage DisplayType = "boost_percentage"
	Number          DisplayType = "number"
	Date            DisplayType = "date"
)

// NsynNftMetadata holds nsyn nft metadata, the nfts main game data it stored here to show on opensea
type NsynNftMetadata struct {
	TokenID            uint64                `json:"token_id" db:"token_id"`
	Name               string                `json:"name" db:"name"`
	Game               string                `json:"game" db:"game"`
	GameObject         interface{}           `json:"game_object" db:"game_object"`
	Description        string                `json:"description" db:"description"`
	ExternalUrl        string                `json:"external_url" db:"external_url"`
	Image              string                `json:"image" db:"image"`
	Attributes         []*Attribute          `json:"attributes" db:"attributes"`
	AdditionalMetadata []*AdditionalMetadata `json:"additional_metadata" db:"additional_metadata"`
	DeletedAt          *time.Time            `json:"deleted_at" db:"deleted_at"`
	UpdatedAt          time.Time             `json:"updated_at" db:"updated_at"`
	CreatedAt          time.Time             `json:"created_at" db:"created_at"`
}

// AdditionalMetadata holds metadata for a nfts non main game
type AdditionalMetadata struct {
	Game        string       `json:"game" db:"game"`
	GameObject  interface{}  `json:"game_object" db:"game_object"`
	Name        string       `json:"name" db:"name"`
	Description string       `json:"description" db:"description"`
	ExternalUrl string       `json:"external_url" db:"external_url"`
	Image       string       `json:"image" db:"image"`
	Attributes  []*Attribute `json:"attributes" db:"attributes"`
	DeletedAt   *time.Time   `json:"deleted_at" db:"deleted_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
}
