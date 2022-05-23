package types

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"
)

type Collection struct {
	ID            CollectionID `json:"id"`
	Name          string       `json:"name"`
	Slug          string       `json:"slug"`
	ImageURL      string       `json:"image_url"`
	DeletedAt     *time.Time   `json:"deleted_at" db:"deleted_at"`
	UpdatedAt     time.Time    `json:"updated_at" db:"updated_at"`
	CreatedAt     time.Time    `json:"created_at" db:"created_at"`
	MintContract  string       `json:"mint_contract" db:"mint_contract"`
	StakeContract string       `json:"stake_contract" db:"stake_contract"`
}

type AttributeOld struct {
	DisplayType DisplayType `json:"display_type,omitempty"`
	TraitType   string      `json:"trait_type"`
	TokenID     uint64      `json:"token_id,omitempty"`
	Value       interface{} `json:"value"` // string or number only
}



// StoreItem holds data for a nft that is listed on the marketplace
type StoreItem struct {
	ID                   StoreItemID         `json:"id" db:"id"`
	RestrictionGroup     string              `json:"restriction_group" db:"restriction_group"`
	Name                 string              `json:"name" db:"name"`
	IsDefault            bool                `json:"is_default" db:"is_default"`
	Tier                 string              `json:"tier" db:"tier"`
	WhitelistedAddresses []string            `json:"white_listed_addresses" db:"white_listed_addresses"`
	Restriction          string              `json:"restriction" db:"restriction"`
	FactionID            FactionID           `json:"faction_id" db:"faction_id"`
	Faction              *Faction            `json:"faction" db:"faction"`
	CollectionID         CollectionID        `json:"collection_id" db:"collection_id"`
	Collection           *Collection         `json:"collection" db:"collection"`
	Description          string              `json:"description" db:"description"`
	Image                string              `json:"image" db:"image"`
	ImageAvatar          string              `json:"image_avatar" db:"image_avatar"`
	AnimationURL         string              `json:"animation_url" db:"animation_url"`
	Attributes           []*AttributeOld     `json:"attributes" db:"attributes"`
	AdditionalMetadata   *AdditionalMetadata `json:"additional_metadata" db:"additional_metadata"`
	UsdCentCost          int                 `json:"usd_cent_cost" db:"usd_cent_cost"`
	AmountSold           int                 `json:"amount_sold" db:"amount_sold"`
	AmountAvailable      int                 `json:"amount_available" db:"amount_available"`
	SoldAfter            time.Time           `json:"sold_after" db:"sold_after"`
	SoldBefore           time.Time           `json:"sold_before" db:"sold_before"`
	DeletedAt            *time.Time          `json:"deleted_at" db:"deleted_at"`
	CreatedAt            time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time           `json:"updated_at" db:"updated_at"`
	SupCost              string              `json:"sup_cost"`
}

// XsynMetadata holds xsyn nft metadata, the nfts main game data it stored here to show on opensea
type XsynMetadata struct {
	Hash               string                `json:"hash" db:"hash"`
	UserID             *UserID               `json:"user_id" db:"user_id"`
	Minted             bool                  `json:"minted" db:"minted"`
	Username           *string               `json:"username" db:"username"`
	ExternalTokenID    uint64                `json:"external_token_id" db:"external_token_id"`
	Name               string                `json:"name" db:"name"`
	CollectionID       CollectionID          `json:"collection_id" db:"collection_id"`
	Collection         Collection            `json:"collection" db:"collection"`
	GameObject         interface{}           `json:"game_object" db:"game_object"`
	Description        *string               `json:"description,omitempty" db:"description,omitempty"`
	ExternalUrl        string                `json:"external_url" db:"external_url"`
	Image              string                `json:"image" db:"image"`
	ImageAvatar        string                `json:"image_avatar" db:"image_avatar"`
	AnimationURL       string                `json:"animation_url" db:"animation_url"`
	Durability         int                   `json:"durability" db:"durability"`
	Attributes         []*AttributeOld       `json:"attributes" db:"attributes"`
	AdditionalMetadata []*AdditionalMetadata `json:"additional_metadata" db:"additional_metadata"`
	DeletedAt          *time.Time            `json:"deleted_at" db:"deleted_at"`
	FrozenAt           *time.Time            `json:"frozen_at" db:"frozen_at"`
	LockedByID         *UserID               `json:"locked_by_id" db:"locked_by_id"`
	MintingSignature   string                `json:"minting_signature" db:"minting_signature"`
	SignatureExpiry    string                `json:"signature_expiry" db:"signature_expiry"`
	UpdatedAt          time.Time             `json:"updated_at" db:"updated_at"`
	CreatedAt          time.Time             `json:"created_at" db:"created_at"`
	TxHistory          []string              `json:"tx_history" db:"tx_history"`
}

func (c *Collection) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(b, c)
}

type PurchasedItem struct {
	Hash            string       `json:"hash" db:"hash"`
	UserID          *UserID      `json:"user_id" db:"user_id"`
	Minted          bool         `json:"minted" db:"minted"`
	Username        *string      `json:"username" db:"username,omitempty"`
	ExternalTokenID uint64       `json:"external_token_id" db:"external_token_id"`
	CollectionID    CollectionID `json:"collection_id" db:"collection_id"`
	Collection      *Collection  `json:"collection" db:"collection"`
	GameObject      interface{}  `json:"game_object" db:"game_object"`
	ExternalUrl     string       `json:"external_url" db:"external_url"`
	Image           string       `json:"image" db:"image"`
	ImageAvatar     string       `json:"image_avatar" db:"image_avatar"`
	AnimationURL    string       `json:"animation_url" db:"animation_url"`
	DeletedAt       *time.Time   `json:"deleted_at" db:"deleted_at"`
	FrozenAt        *time.Time   `json:"frozen_at" db:"frozen_at"`
	LockedByID      *UserID      `json:"locked_by_id" db:"locked_by_id"`
	UpdatedAt       time.Time    `json:"updated_at" db:"updated_at"`
	CreatedAt       time.Time    `json:"created_at" db:"created_at"`
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
	TokenID     uint64       `json:"token_id"`
	Collection  Collection   `json:"collection" db:"collection"`
	GameObject  interface{}  `json:"game_object" db:"game_object"`
	Name        string       `json:"name" db:"name"`
	Description string       `json:"description" db:"description"`
	ExternalUrl string       `json:"external_url" db:"external_url"`
	Image       string          `json:"image" db:"image"`
	Attributes  []*AttributeOld `json:"attributes" db:"attributes"`
	DeletedAt   *time.Time      `json:"deleted_at" db:"deleted_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
}

type WarMachineMetadata struct {
	Hash               string             `json:"hash"`
	OwnedByID          UserID             `json:"owned_by_id"`
	Name               string             `json:"name"`
	Description        *string            `json:"description,omitempty"`
	ExternalUrl        string             `json:"external_url"`
	Image              string             `json:"image"`
	Model              string             `json:"model"`
	Skin               string             `json:"skin"`
	MaxHealth          int                `json:"max_health"`
	Health             int                `json:"health"`
	MaxShield          int                `json:"max_shield"`
	Shield             int                `json:"shield"`
	ShieldRechargeRate float64            `json:"shield_recharge_rate"`
	Speed              int                `json:"speed"`
	Durability         int                `json:"durability"`
	PowerGrid          int                `json:"power_grid"`
	CPU                int                `json:"cpu"`
	WeaponHardpoint    int                `json:"weapon_hardpoint"`
	WeaponNames        []string           `json:"weapon_names"`
	TurretHardpoint    int                `json:"turret_hardpoint"`
	UtilitySlots       int                `json:"utility_slots"`
	FactionID          FactionID          `json:"faction_id"`
	Faction            *Faction           `json:"faction"`
	Abilities          []*AbilityMetadata `json:"abilities"`
}

type WarMachineAttField string

const (
	WarMachineAttName                    WarMachineAttField = "Name"
	WarMachineAttModel                   WarMachineAttField = "Model"
	WarMachineAttSubModel                WarMachineAttField = "SubModel"
	WarMachineAttFieldMaxHitPoint        WarMachineAttField = "Max Structure Hit Points"
	WarMachineAttFieldMaxShieldHitPoint  WarMachineAttField = "Max Shield Hit Points"
	WarMachineAttFieldSpeed              WarMachineAttField = "Speed"
	WarMachineAttFieldPowerGrid          WarMachineAttField = "Power Grid"
	WarMachineAttFieldShieldRechargeRate WarMachineAttField = "Shield Recharge Rate"
	WarMachineAttFieldCPU                WarMachineAttField = "CPU"
	WarMachineAttFieldWeaponHardpoints   WarMachineAttField = "Weapon Hardpoints"
	WarMachineAttFieldTurretHardpoints   WarMachineAttField = "Turret Hardpoints"
	WarMachineAttFieldUtilitySlots       WarMachineAttField = "Utility Slots"
	WarMachineAttFieldWeapon01           WarMachineAttField = "Weapon One"
	WarMachineAttFieldWeapon02           WarMachineAttField = "Weapon Two"
	WarMachineAttFieldTurret01           WarMachineAttField = "Turret One"
	WarMachineAttFieldTurret02           WarMachineAttField = "Turret Two"
	WarMachineAttFieldAbility01          WarMachineAttField = "Ability One"
	WarMachineAttFieldAbility02          WarMachineAttField = "Ability Two"
)

type AbilityMetadata struct {
	Hash              string  `json:"hash"`
	TokenID           uint64  `json:"token_id"`
	Name              string  `json:"name"`
	Description       *string `json:"description"`
	ExternalUrl       string  `json:"external_url"`
	Image             string  `json:"image"`
	SupsCost          string  `json:"sups_cost"`
	GameClientID      int     `json:"game_client_id"`
	RequiredSlot      string  `json:"required_slot"`
	RequiredPowerGrid int     `json:"required_power_grid"`
	RequiredCPU       int     `json:"required_cpu"`
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
	abilityMetadata.TokenID = metadata.ExternalTokenID
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

type WarMachineContract struct {
	CurrentReward big.Int
}

type RepairMode string

const (
	RepairModeFast     = "FAST"
	RepairModeStandard = "STANDARD"
)

type AssetRepairRecord struct {
	Hash              string     `json:"hash"`
	StartedAt         time.Time  `json:"started_at"` // this is calculated on fly value
	ExpectCompletedAt time.Time  `json:"expect_complete_at"`
	RepairMode        RepairMode `json:"repair_mode"`
	IsPaidToComplete  bool       `json:"is_paid_to_complete"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}
