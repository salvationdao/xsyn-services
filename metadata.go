package passport

import (
	"time"
)

type Collection struct {
	ID        CollectionID `json:"id"`
	Name      string       `json:"name"`
	ImageURL  string       `json:"imageURL"`
	DeletedAt *time.Time   `json:"deleted_at" db:"deleted_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
}

type Attribute struct {
	DisplayType DisplayType `json:"display_type,omitempty"`
	TraitType   string      `json:"trait_type"`
	TokenID     uint64      `json:"token_id,omitempty"`
	Value       interface{} `json:"value"` // string or number only
}

type DisplayType string

const (
	BoostNumber     DisplayType = "boost_number"
	BoostPercentage DisplayType = "boost_percentage"
	Number          DisplayType = "number"
	Date            DisplayType = "date"
)

// StoreItem holds data for a nft that is listed on the marketplace
type StoreItem struct {
	ID              StoreItemID  `json:"ID" db:"id"`
	Name            string       `json:"name" db:"name"`
	FactionID       FactionID    `json:"factionID" db:"faction_id"`
	CollectionID    CollectionID `json:"collectionID" db:"collection_id"`
	Description     string       `json:"description" db:"description"`
	Image           string       `json:"image" db:"image"`
	Attributes      []*Attribute `json:"attributes" db:"attributes"`
	UsdCentCost     int          `json:"usdCentCost" db:"usd_cent_cost"`
	AmountSold      int          `json:"amountSold" db:"amount_sold"`
	AmountAvailable int          `json:"amountAvailable" db:"amount_available"`
	SoldAfter       time.Time    `json:"soldAfter" db:"sold_after"`
	SoldBefore      time.Time    `json:"soldBefore" db:"sold_before"`
	DeletedAt       *time.Time   `json:"deletedAt" db:"deleted_at"`
	CreatedAt       time.Time    `json:"createdAt" db:"created_at"`
	UpdatedAt       time.Time    `json:"updatedAt" db:"updated_at"`
}

// XsynMetadata holds xsyn nft metadata, the nfts main game data it stored here to show on opensea
type XsynMetadata struct {
	UserID             *UserID               `json:"userID" db:"user_id"`
	TokenID            uint64                `json:"tokenID" db:"token_id"`
	Name               string                `json:"name" db:"name"`
	CollectionID       CollectionID          `json:"collectionID" db:"collection_id"`
	Collection         Collection            `json:"collection" db:"collection"`
	GameObject         interface{}           `json:"game_object" db:"game_object"`
	Description        string                `json:"description" db:"description"`
	ExternalUrl        string                `json:"external_url" db:"external_url"`
	Image              string                `json:"image" db:"image"`
	Durability         int                   `json:"durability" db:"durability"`
	Attributes         []*Attribute          `json:"attributes" db:"attributes"`
	AdditionalMetadata []*AdditionalMetadata `json:"additional_metadata" db:"additional_metadata"`
	DeletedAt          *time.Time            `json:"deleted_at" db:"deleted_at"`
	FrozenAt           *time.Time            `json:"frozenAt" db:"frozen_at"`
	UpdatedAt          time.Time             `json:"updatedAt" db:"updated_at"`
	CreatedAt          time.Time             `json:"createdAt" db:"created_at"`
}

type AssetType string

const (
	WarMachine AssetType = "War Machine"
	Weapon     AssetType = "Weapon"
	Utility    AssetType = "Utility"
)

// AdditionalMetadata holds metadata for a nfts non main game
type AdditionalMetadata struct {
	TokenID     uint64       `json:"tokenID"`
	Collection  Collection   `json:"collection" db:"collection"`
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

type WarMachineMetadata struct {
	TokenID         uint64    `json:"tokenID"`
	OwnedByID       UserID    `json:"ownedByID"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	ExternalUrl     string    `json:"externalUrl"`
	Image           string    `json:"image"`
	MaxHealth       int       `json:"maxHealth"`
	Health          int       `json:"health"`
	MaxShield       int       `json:"maxShield"`
	Shield          int       `json:"shield"`
	Speed           int       `json:"speed"`
	Durability      int       `json:"durability"`
	PowerGrid       int       `json:"powerGrid"`
	CPU             int       `json:"cpu"`
	WeaponHardpoint int       `json:"weaponHardpoint"`
	WeaponNames     []string  `json:"weaponNames"`
	TurretHardpoint int       `json:"turretHardpoint"`
	UtilitySlots    int       `json:"utilitySlots"`
	FactionID       FactionID `json:"factionID"`
	Faction         *Faction  `json:"faction"`
}

type WarMachineAttField string

const (
	WarMachineAttFieldMaxHitPoint       WarMachineAttField = "Max Structure Hit Points"
	WarMachineAttFieldMaxShieldHitPoint WarMachineAttField = "Max Shield Hit Points"
	WarMachineAttFieldSpeed             WarMachineAttField = "Speed"
	WarMachineAttFieldPowerGrid         WarMachineAttField = "Power Grid"
	WarMachineAttFieldCPU               WarMachineAttField = "CPU"
	WarMachineAttFieldWeaponHardpoints  WarMachineAttField = "Weapon Hardpoints"
	WarMachineAttFieldTurretHardpoints  WarMachineAttField = "Turret Hardpoints"
	WarMachineAttFieldUtilitySlots      WarMachineAttField = "Utility Slots"
	WarMachineAttFieldWeapon01          WarMachineAttField = "Weapon One"
	WarMachineAttFieldWeapon02          WarMachineAttField = "Weapon Two"
	WarMachineAttFieldTurret01          WarMachineAttField = "Turret One"
	WarMachineAttFieldTurret02          WarMachineAttField = "Turret Two"
)

// ParseWarMachineMetadata convert json attribute to proper struct
func ParseWarMachineMetadata(metadata *XsynMetadata, warMachineMetadata *WarMachineMetadata) {
	warMachineMetadata.TokenID = metadata.TokenID
	warMachineMetadata.Name = metadata.Name
	warMachineMetadata.Description = metadata.Description
	warMachineMetadata.ExternalUrl = metadata.ExternalUrl
	warMachineMetadata.Image = metadata.Image
	warMachineMetadata.Durability = metadata.Durability

	for _, att := range metadata.Attributes {
		switch att.TraitType {
		case string(WarMachineAttFieldMaxHitPoint):
			warMachineMetadata.MaxHealth = int(att.Value.(float64))
			warMachineMetadata.Health = int(att.Value.(float64))
		case string(WarMachineAttFieldMaxShieldHitPoint):
			warMachineMetadata.MaxShield = int(att.Value.(float64))
			warMachineMetadata.Shield = int(att.Value.(float64))
		case string(WarMachineAttFieldSpeed):
			warMachineMetadata.Speed = int(att.Value.(float64))
		case string(WarMachineAttFieldPowerGrid):
			warMachineMetadata.PowerGrid = int(att.Value.(float64))
		case string(WarMachineAttFieldCPU):
			warMachineMetadata.CPU = int(att.Value.(float64))
		case string(WarMachineAttFieldWeaponHardpoints):
			warMachineMetadata.WeaponHardpoint = int(att.Value.(float64))
		case string(WarMachineAttFieldTurretHardpoints):
			warMachineMetadata.TurretHardpoint = int(att.Value.(float64))
		case string(WarMachineAttFieldUtilitySlots):
			warMachineMetadata.UtilitySlots = int(att.Value.(float64))
		case string(WarMachineAttFieldWeapon01),
			string(WarMachineAttFieldWeapon02),
			string(WarMachineAttFieldTurret01),
			string(WarMachineAttFieldTurret02):
			warMachineMetadata.WeaponNames = append(warMachineMetadata.WeaponNames, att.Value.(string))

		}

	}

}
