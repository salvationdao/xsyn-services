package rpcclient

import (
	"time"

	"github.com/volatiletech/null/v8"
)

type TemplateContainer struct {
	Template         *Template
	BlueprintChassis *BlueprintChassis
	BlueprintWeapons map[int]*BlueprintWeapon
	BlueprintTurrets map[int]*BlueprintWeapon
	BlueprintModules map[int]*BlueprintModule
}
type Template struct {
	ID                 string
	BlueprintChassisID string
	FactionID          string
	Tier               string
	Label              string
	Slug               string
	IsDefault          bool
	ImageURL           string
	AnimationURL       string
	DeletedAt          null.Time
	UpdatedAt          time.Time
	CreatedAt          time.Time
}
type BlueprintChassis struct {
	ID                 string
	BrandID            string
	Label              string
	Slug               string
	Model              string
	Skin               string
	ShieldRechargeRate int
	WeaponHardpoints   int
	TurretHardpoints   int
	UtilitySlots       int
	Speed              int
	MaxHitpoints       int
	MaxShield          int
	DeletedAt          null.Time
	UpdatedAt          time.Time
	CreatedAt          time.Time
}
type BlueprintWeapon struct {
	ID         string
	BrandID    null.String
	Label      string
	Slug       string
	Damage     int
	WeaponType string
	DeletedAt  null.Time
	UpdatedAt  time.Time
	CreatedAt  time.Time
}
type BlueprintModule struct {
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
