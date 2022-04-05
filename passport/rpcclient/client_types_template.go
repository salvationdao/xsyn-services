package rpcclient

import (
	"time"

	"github.com/volatiletech/null/v8"
)

type TemplateContainer struct {
	Template         *Template                `json:"template"`
	BlueprintChassis *BlueprintChassis        `json:"blueprint_chassis"`
	BlueprintWeapons map[int]*BlueprintWeapon `json:"blueprint_weapons"`
	BlueprintTurrets map[int]*BlueprintWeapon `json:"blueprint_turrets"`
	BlueprintModules map[int]*BlueprintModule `json:"blueprint_modules"`
}
type Template struct {
	ID                 string      `json:"id"`
	BlueprintChassisID string      `json:"blueprint_chassis_id"`
	FactionID          string      `json:"faction_id"`
	Tier               string      `json:"tier"`
	Label              string      `json:"label"`
	Slug               string      `json:"slug"`
	IsDefault          bool        `json:"is_default"`
	LargeImageURL      string      `json:"large_image_url"`
	ImageURL           string      `json:"image_url"`
	AnimationURL       string      `json:"animation_url"`
	CardAnimationURL   string      `json:"card_animation_url"`
	AvatarURL          string      `json:"avatar_url"`
	AssetType          string      `json:"asset_type"`
	CollectionSlug     null.String `json:"collection_slug"`
	DeletedAt          null.Time   `json:"deleted_at"`
	UpdatedAt          time.Time   `json:"updated_at"`
	CreatedAt          time.Time   `json:"created_at"`
}
type BlueprintChassis struct {
	ID                 string    `json:"id"`
	BrandID            string    `json:"brand_id"`
	Label              string    `json:"label"`
	Slug               string    `json:"slug"`
	Model              string    `json:"model"`
	Skin               string    `json:"skin"`
	ShieldRechargeRate int       `json:"shield_recharge_rate"`
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
type BlueprintWeapon struct {
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
type BlueprintModule struct {
	ID               string      `json:"id"`
	BrandID          null.String `json:"brand_id"`
	Slug             string      `json:"slug"`
	Label            string      `json:"label"`
	HitpointModifier int         `json:"hitpoint_modifier"`
	ShieldModifier   int         `json:"shield_modifier"`
	DeletedAt        null.Time   `json:"deleted_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
	CreatedAt        time.Time   `json:"created_at"`
}
