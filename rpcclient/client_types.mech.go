package rpcclient

import (
	"time"

	null "github.com/volatiletech/null/v8"
)

type MechContainer struct {
	Mech    *Mech              `json:"mech"`
	Chassis *Chassis           `json:"chassis"`
	Weapons map[string]*Weapon `json:"weapons"`
	Turrets map[string]*Weapon `json:"turrets"`
	Modules map[string]*Module `json:"modules"`
}

type Mech struct {
	ID               string    `json:"id"`
	OwnerID          string    `json:"owner_id"`
	TemplateID       string    `json:"template_id"`
	ChassisID        string    `json:"chassis_id"`
	ExternalTokenID  int       `json:"external_token_id"`
	Tier             string    `json:"tier"`
	IsDefault        bool      `json:"is_default"`
	ImageURL         string    `json:"image_url"`
	AnimationURL     string    `json:"animation_url"`
	CardAnimationURL string    `json:"card_animation_url"`
	AvatarURL        string    `json:"avatar_url"`
	Hash             string    `json:"hash"`
	Name             string    `json:"name"`
	Label            string    `json:"label"`
	Slug             string    `json:"slug"`
	AssetType        string    `json:"asset_type"`
	DeletedAt        null.Time `json:"deleted_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	CreatedAt        time.Time `json:"created_at"`
}
type Chassis struct {
	ID                 string    `json:"id"`
	BrandID            string    `json:"brand_id"`
	Label              string    `json:"label"`
	Model              string    `json:"model"`
	Skin               string    `json:"skin"`
	Slug               string    `json:"slug"`
	ShieldRechargeRate int       `json:"shield_recharge_rate"`
	HealthRemaining    int       `json:"health_remaining"`
	WeaponHardpoints   int       `json:"weapon_hardpoints"`
	TurretHardpoints   int       `json:"turret_hardpoints"`
	UtilitySlots       int       `json:"utility_slots"`
	Speed              int       `json:"speed"`
	MaxHitpoints       int       `json:"max_hitpoints"`
	MaxShield          int       `json:"max_shield"`
	DeletedAt          null.Time `json:"deleted_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	CreatedAt          time.Time `json:"created_at"`
}

type Weapon struct {
	ID         string      `json:"id"`
	BrandID    null.String `json:"brand_id"`
	Label      string      `json:"label"`
	Slug       string      `json:"slug"`
	Damage     int         `json:"damage"`
	WeaponType string      `json:"weapon_type"`
	DeletedAt  null.Time   `json:"deleted_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
	CreatedAt  time.Time   `json:"created_at"`
}
type Module struct {
	ID               string
	BrandID          null.String
	Slug             string
	Label            string
	HitpointModifier int
	ShieldModifier   int
	DeletedAt        null.Time
	UpdatedAt        time.Time
	CreatedAt        time.Time
}
