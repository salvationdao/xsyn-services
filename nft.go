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

// XsynNftMetadata holds xsyn nft metadata, the nfts main game data it stored here to show on opensea
type XsynNftMetadata struct {
	TokenID            uint64                `json:"token_id" db:"token_id"`
	Name               string                `json:"name" db:"name"`
	Collection         string                `json:"collection" db:"collection"`
	GameObject         interface{}           `json:"game_object" db:"game_object"`
	Description        string                `json:"description" db:"description"`
	ExternalUrl        string                `json:"external_url" db:"external_url"`
	Image              string                `json:"image" db:"image"`
	Durability         int                   `json:"durability" db:"durability"`
	Attributes         []*Attribute          `json:"attributes" db:"attributes"`
	AdditionalMetadata []*AdditionalMetadata `json:"additional_metadata" db:"additional_metadata"`
	DeletedAt          *time.Time            `json:"deleted_at" db:"deleted_at"`
	UpdatedAt          time.Time             `json:"updated_at" db:"updated_at"`
	CreatedAt          time.Time             `json:"created_at" db:"created_at"`
}

type AssetType string

const (
	WarMachine AssetType = "War Machine"
	Weapon     AssetType = "Weapon"
	Utility    AssetType = "Utility"
)

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

type WarMachineNFT struct {
	TokenID         uint64    `json:"tokenID"`
	OwnedByID       UserID    `json:"ownedByID"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	ExternalUrl     string    `json:"externalUrl"`
	Image           string    `json:"image"`
	MaxHitPoint     int       `json:"maxHitPoint"`
	RemainHitPoint  int       `json:"remainHitPoint"`
	Speed           int       `json:"speed"`
	Durability      int       `json:"durability"`
	PowerGrid       int       `json:"powerGrid"`
	CPU             int       `json:"cpu"`
	WeaponHardpoint int       `json:"weaponHardpoint"`
	TurretHardpoint int       `json:"turretHardpoint"`
	UtilitySlots    int       `json:"utilitySlots"`
	FactionID       FactionID `json:"factionID"`
}
