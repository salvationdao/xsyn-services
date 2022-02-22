package passport

import (
	"math/big"
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
	ID                 StoreItemID         `json:"ID" db:"id"`
	Name               string              `json:"name" db:"name"`
	Restriction        string              `json:"restriction" db:"restriction"`
	FactionID          FactionID           `json:"factionID" db:"faction_id"`
	Faction            *Faction            `json:"faction" db:"faction"`
	CollectionID       CollectionID        `json:"collectionID" db:"collection_id"`
	Collection         Collection          `json:"collection" db:"collection"`
	Description        string              `json:"description" db:"description"`
	Image              string              `json:"image" db:"image"`
	AnimationURL       string              `json:"animation_url" db:"animation_url"`
	Attributes         []*Attribute        `json:"attributes" db:"attributes"`
	AdditionalMetadata *AdditionalMetadata `json:"additionalMetadata" db:"additional_metadata"`
	UsdCentCost        int                 `json:"usdCentCost" db:"usd_cent_cost"`
	AmountSold         int                 `json:"amountSold" db:"amount_sold"`
	AmountAvailable    int                 `json:"amountAvailable" db:"amount_available"`
	SoldAfter          time.Time           `json:"soldAfter" db:"sold_after"`
	SoldBefore         time.Time           `json:"soldBefore" db:"sold_before"`
	DeletedAt          *time.Time          `json:"deletedAt" db:"deleted_at"`
	CreatedAt          time.Time           `json:"createdAt" db:"created_at"`
	UpdatedAt          time.Time           `json:"updatedAt" db:"updated_at"`
	SupCost            string              `json:"supCost"`
}

// XsynMetadata holds xsyn nft metadata, the nfts main game data it stored here to show on opensea
type XsynMetadata struct {
	UserID             *UserID               `json:"userID" db:"user_id"`
	Username           *string               `json:"username" db:"username"`
	TokenID            uint64                `json:"tokenID" db:"token_id"`
	Name               string                `json:"name" db:"name"`
	CollectionID       CollectionID          `json:"collectionID" db:"collection_id"`
	Collection         Collection            `json:"collection" db:"collection"`
	GameObject         interface{}           `json:"game_object" db:"game_object"`
	Description        string                `json:"description" db:"description"`
	ExternalUrl        string                `json:"external_url" db:"external_url"`
	Image              string                `json:"image" db:"image"`
	AnimationURL       string                `json:"animation_url" db:"animation_url"`
	Durability         int                   `json:"durability" db:"durability"`
	Attributes         []*Attribute          `json:"attributes" db:"attributes"`
	AdditionalMetadata []*AdditionalMetadata `json:"additional_metadata" db:"additional_metadata"`
	DeletedAt          *time.Time            `json:"deleted_at" db:"deleted_at"`
	FrozenAt           *time.Time            `json:"frozenAt" db:"frozen_at"`
	LockedByID         *UserID               `json:"lockedByID" db:"locked_by_id"`
	MintingSignature   string                `json:"mintingSignature" db:"minting_signature"`
	UpdatedAt          time.Time             `json:"updatedAt" db:"updated_at"`
	CreatedAt          time.Time             `json:"createdAt" db:"created_at"`
	TxHistory          interface{}           `json:"txHistory" db:"tx_history"`
}

type AssetType string

const (
	WarMachine AssetType = "War Machine"
	Weapon     AssetType = "Weapon"
	Utility    AssetType = "Utility"
	Ability    AssetType = "Ability"
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
	TokenID         uint64             `json:"tokenID"`
	OwnedByID       UserID             `json:"ownedByID"`
	Name            string             `json:"name"`
	Description     string             `json:"description"`
	ExternalUrl     string             `json:"externalUrl"`
	Image           string             `json:"image"`
	Model           string             `json:"model"`
	Skin            string             `json:"skin"`
	MaxHealth       int                `json:"maxHealth"`
	Health          int                `json:"health"`
	MaxShield       int                `json:"maxShield"`
	Shield          int                `json:"shield"`
	Speed           int                `json:"speed"`
	Durability      int                `json:"durability"`
	PowerGrid       int                `json:"powerGrid"`
	CPU             int                `json:"cpu"`
	WeaponHardpoint int                `json:"weaponHardpoint"`
	WeaponNames     []string           `json:"weaponNames"`
	TurretHardpoint int                `json:"turretHardpoint"`
	UtilitySlots    int                `json:"utilitySlots"`
	FactionID       FactionID          `json:"factionID"`
	Faction         *Faction           `json:"faction"`
	Abilities       []*AbilityMetadata `json:"abilities"`

	ContractReward big.Int `json:"contractReward"`
	IsInsured      bool    `json:"isInsured"`
}

type WarMachineAttField string

const (
	WarMachineAttName                   WarMachineAttField = "Name"
	WarMachineAttModel                  WarMachineAttField = "Model"
	WarMachineAttSkin                   WarMachineAttField = "Skin"
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
	WarMachineAttFieldAbility01         WarMachineAttField = "Ability One"
	WarMachineAttFieldAbility02         WarMachineAttField = "Ability Two"
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
		case string(WarMachineAttName):
			warMachineMetadata.Name = att.Value.(string)
		case string(WarMachineAttModel):
			warMachineMetadata.Model = att.Value.(string)
		case string(WarMachineAttSkin):
			warMachineMetadata.Skin = att.Value.(string)
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
		case string(WarMachineAttFieldAbility01),
			string(WarMachineAttFieldAbility02):
			if att.TokenID == 0 {
				continue
			}
			warMachineMetadata.Abilities = append(warMachineMetadata.Abilities, &AbilityMetadata{
				TokenID: uint64(att.TokenID),
			})
		}

	}

}

type AbilityMetadata struct {
	TokenID           uint64 `json:"tokenID"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	ExternalUrl       string `json:"externalUrl"`
	Image             string `json:"image"`
	SupsCost          string `json:"supsCost"`
	GameClientID      int    `json:"gameClientID"`
	RequiredSlot      string `json:"requiredSlot"`
	RequiredPowerGrid int    `json:"requiredPowerGrid"`
	RequiredCPU       int    `json:"requiredCPU"`
}

type AbilityAttField string

const (
	AbilityAttFieldAbilityCost       AbilityAttField = "Ability Cost"
	AbilityAttFieldAbilityID         AbilityAttField = "Ability ID"
	AbilityAttFieldRequiredSlot      AbilityAttField = "Required Slot"
	AbilityAttFieldRequiredPowerGrid AbilityAttField = "Required Power Grid"
	AbilityAttFieldRequiredCPU       AbilityAttField = "Required CPU"
)

// ParseAbilityMetadata convert json attribute to proper struct
func ParseAbilityMetadata(metadata *XsynMetadata, abilityMetadata *AbilityMetadata) {
	abilityMetadata.TokenID = metadata.TokenID
	abilityMetadata.Name = metadata.Name
	abilityMetadata.Description = metadata.Description
	abilityMetadata.ExternalUrl = metadata.ExternalUrl
	abilityMetadata.Image = metadata.Image

	for _, att := range metadata.Attributes {
		switch att.TraitType {
		case string(AbilityAttFieldAbilityCost):
			abilityMetadata.SupsCost = att.Value.(string)
		case string(AbilityAttFieldAbilityID):
			abilityMetadata.GameClientID = int(att.Value.(float64))
		case string(AbilityAttFieldRequiredSlot):
			abilityMetadata.RequiredSlot = att.Value.(string)
		case string(AbilityAttFieldRequiredPowerGrid):
			abilityMetadata.RequiredPowerGrid = int(att.Value.(float64))
		case string(AbilityAttFieldRequiredCPU):
			abilityMetadata.RequiredCPU = int(att.Value.(float64))
		}
	}
}

// IsUsable returns true if the asset isn't locked in any way
func (xsmd *XsynMetadata) IsUsable() bool {
	if xsmd.LockedByID != nil && !xsmd.LockedByID.IsNil() {
		return false
	}
	if xsmd.FrozenAt != nil && !xsmd.FrozenAt.IsZero() {
		return false
	}
	if xsmd.DeletedAt != nil && !xsmd.DeletedAt.IsZero() {
		return false
	}
	if xsmd.MintingSignature != "" {
		return false
	}

	return true
}
